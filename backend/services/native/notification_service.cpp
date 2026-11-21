/**
 * @file notification_service.cpp
 * @brief Event-driven Windows toast interception and native toast dispatch.
 *
 * Watches the Windows Push Notification database (wpndatabase.db) and WinRT
 * UserNotificationListener events to capture incoming toasts without idle CPU
 * polling. Parses toast XML directly via string scanning, resolves app icons,
 * and emits JSON payloads to the Go backend.
 */

#include "notification_service.h"
#include "icon_service.h"
#include "winrt_common.h"

#include <shlobj.h>
#include <propkey.h>
#include <propvarutil.h>
#include <wrl/client.h>

#include <winrt/Windows.UI.Notifications.Management.h>
#include <winrt/Windows.UI.Notifications.h>
#include <winrt/Windows.Data.Xml.Dom.h>
#include <winrt/Windows.Foundation.h>

#include <atomic>
#include <cctype>
#include <cstring>
#include <cwchar>
#include <mutex>
#include <sstream>
#include <string>
#include <thread>
#include <unordered_map>
#include <vector>

namespace winrt_notif = winrt::Windows::UI::Notifications;
namespace winrt_notif_mgmt = winrt::Windows::UI::Notifications::Management;

static std::atomic<NativeNotificationCallback> g_notifCb{nullptr};
static std::atomic<bool> g_notifRunning{false};
static std::atomic<int> g_activeEnrichTasks{0};
static HANDLE g_stopEvent = nullptr;
static HANDLE g_wakeEvent = nullptr;
static std::thread g_notifThread;

typedef struct sqlite3 sqlite3;
typedef struct sqlite3_stmt sqlite3_stmt;

/**
 * @struct WinSqliteApi
 * @brief Lazily resolves winsqlite3.dll exports once per process lifetime.
 */
struct WinSqliteApi {
    int (*open_v2)(const char*, sqlite3**, int, const char*) = nullptr;
    int (*prepare_v2)(sqlite3*, const char*, int, sqlite3_stmt**, const char**) = nullptr;
    int (*bind_int64)(sqlite3_stmt*, int, int64_t) = nullptr;
    int (*step)(sqlite3_stmt*) = nullptr;
    const unsigned char* (*column_text)(sqlite3_stmt*, int) = nullptr;
    const void* (*column_blob)(sqlite3_stmt*, int) = nullptr;
    int (*column_bytes)(sqlite3_stmt*, int) = nullptr;
    int64_t (*column_int64)(sqlite3_stmt*, int) = nullptr;
    int (*finalize)(sqlite3_stmt*) = nullptr;
    int (*close)(sqlite3*) = nullptr;
    bool valid = false;

    static const WinSqliteApi& Get() {
        static WinSqliteApi api = []() {
            WinSqliteApi a;
            HMODULE h = LoadLibraryW(L"winsqlite3.dll");
            if (!h)
                return a;
            a.open_v2 = reinterpret_cast<decltype(a.open_v2)>(GetProcAddress(h, "sqlite3_open_v2"));
            a.prepare_v2 = reinterpret_cast<decltype(a.prepare_v2)>(GetProcAddress(h, "sqlite3_prepare_v2"));
            a.bind_int64 = reinterpret_cast<decltype(a.bind_int64)>(GetProcAddress(h, "sqlite3_bind_int64"));
            a.step = reinterpret_cast<decltype(a.step)>(GetProcAddress(h, "sqlite3_step"));
            a.column_text = reinterpret_cast<decltype(a.column_text)>(GetProcAddress(h, "sqlite3_column_text"));
            a.column_blob = reinterpret_cast<decltype(a.column_blob)>(GetProcAddress(h, "sqlite3_column_blob"));
            a.column_bytes = reinterpret_cast<decltype(a.column_bytes)>(GetProcAddress(h, "sqlite3_column_bytes"));
            a.column_int64 = reinterpret_cast<decltype(a.column_int64)>(GetProcAddress(h, "sqlite3_column_int64"));
            a.finalize = reinterpret_cast<decltype(a.finalize)>(GetProcAddress(h, "sqlite3_finalize"));
            a.close = reinterpret_cast<decltype(a.close)>(GetProcAddress(h, "sqlite3_close"));
            a.valid = a.open_v2 && a.prepare_v2 && a.step && a.finalize && a.close && a.column_text && a.column_blob &&
                      a.column_bytes && a.column_int64;
            return a;
        }();
        return api;
    }
};

struct DbToastInfo {
    std::string appLogoOverride;
    std::string heroImage;
    std::vector<std::string> inlineImages;

    bool hasAny() const { return !appLogoOverride.empty() || !heroImage.empty() || !inlineImages.empty(); }
};

/**
 * @brief Case-insensitive substring search helper to replace std::regex.
 */
static size_t FindCI(const std::string& hay, const char* needle, size_t startPos = 0) {
    size_t nLen = std::strlen(needle);
    if (nLen == 0 || startPos + nLen > hay.size())
        return std::string::npos;
    for (size_t i = startPos; i + nLen <= hay.size(); ++i) {
        size_t j = 0;
        while (j < nLen && std::tolower(static_cast<unsigned char>(hay[i + j])) ==
                               std::tolower(static_cast<unsigned char>(needle[j]))) {
            ++j;
        }
        if (j == nLen)
            return i;
    }
    return std::string::npos;
}

/**
 * @brief Extracts an XML attribute value (e.g. src="..." or placement='...') from a tag.
 */
static std::string ExtractXmlAttr(const std::string& tag, const char* attrName) {
    size_t pos = FindCI(tag, attrName, 0);
    while (pos != std::string::npos) {
        size_t eq = pos + std::strlen(attrName);
        while (eq < tag.size() && std::isspace(static_cast<unsigned char>(tag[eq])))
            ++eq;
        if (eq < tag.size() && tag[eq] == '=') {
            size_t qStart = eq + 1;
            while (qStart < tag.size() && std::isspace(static_cast<unsigned char>(tag[qStart])))
                ++qStart;
            if (qStart < tag.size() && (tag[qStart] == '"' || tag[qStart] == '\'')) {
                char quote = tag[qStart];
                size_t qEnd = tag.find(quote, qStart + 1);
                if (qEnd != std::string::npos) {
                    return tag.substr(qStart + 1, qEnd - qStart - 1);
                }
            }
        }
        pos = FindCI(tag, attrName, pos + 1);
    }
    return "";
}

static bool IsAvatarPath(const std::string& src) {
    return FindCI(src, "avatar") != std::string::npos || FindCI(src, "profile") != std::string::npos ||
           FindCI(src, "user") != std::string::npos;
}

/**
 * @brief Parses `<image>` tags from toast XML without regex overhead.
 */
static void ParseToastXmlPayload(const std::string& payload, DbToastInfo& outInfo) {
    size_t pos = 0;
    while ((pos = FindCI(payload, "<image", pos)) != std::string::npos) {
        size_t end = payload.find('>', pos);
        if (end == std::string::npos)
            break;

        std::string tag = payload.substr(pos, end - pos + 1);
        pos = end + 1;

        std::string src = ExtractXmlAttr(tag, "src");
        if (src.empty())
            continue;

        std::string placement = ExtractXmlAttr(tag, "placement");
        std::string crop = ExtractXmlAttr(tag, "hint-crop");

        if (FindCI(placement, "applogo") != std::string::npos || FindCI(crop, "circle") != std::string::npos ||
            IsAvatarPath(src)) {
            if (outInfo.appLogoOverride.empty())
                outInfo.appLogoOverride = src;
        } else if (FindCI(placement, "hero") != std::string::npos) {
            outInfo.heroImage = src;
        } else {
            outInfo.inlineImages.push_back(src);
        }
    }
}

static std::string DecodeXmlEntities(std::string out) {
    const struct {
        const char* k;
        const char* v;
        size_t len;
    } ents[] = {{"&amp;", "&", 5}, {"&lt;", "<", 4}, {"&gt;", ">", 4}, {"&quot;", "\"", 6}, {"&apos;", "'", 6}};
    for (const auto& e : ents) {
        for (size_t p = 0; (p = out.find(e.k, p)) != std::string::npos; ++p) {
            out.replace(p, e.len, e.v);
        }
    }
    return out;
}

/**
 * @brief Extracts title and message text from `<text>` nodes in toast XML.
 */
static void ParseToastXmlText(const std::string& payload, std::string& outTitle, std::string& outMessage) {
    size_t pos = 0;
    int idx = 0;
    while ((pos = FindCI(payload, "<text", pos)) != std::string::npos) {
        size_t openEnd = payload.find('>', pos);
        if (openEnd == std::string::npos)
            break;
        if (openEnd > pos && payload[openEnd - 1] == '/') {
            pos = openEnd + 1;
            continue;
        }
        size_t closeStart = FindCI(payload, "</text>", openEnd + 1);
        if (closeStart == std::string::npos)
            break;

        std::string raw = payload.substr(openEnd + 1, closeStart - openEnd - 1);
        pos = closeStart + 7;

        if (raw.rfind("<![CDATA[", 0) == 0 && raw.size() >= 12 && raw.substr(raw.size() - 3) == "]]>") {
            raw = raw.substr(9, raw.size() - 12);
        } else {
            raw = DecodeXmlEntities(raw);
        }
        if (idx++ == 0) {
            outTitle = raw;
        } else {
            if (!outMessage.empty())
                outMessage += "\n";
            outMessage += raw;
        }
    }
}

static std::string CleanMatchString(const std::string& in) {
    std::string out;
    out.reserve(in.size());
    for (size_t i = 0; i < in.size(); ++i) {
        unsigned char c = static_cast<unsigned char>(in[i]);
        if (c < 0x20)
            continue;
        if (c == 0xE2 && i + 2 < in.size() && static_cast<unsigned char>(in[i + 1]) == 0x81 &&
            (static_cast<unsigned char>(in[i + 2]) == 0xA8 || static_cast<unsigned char>(in[i + 2]) == 0xA9)) {
            i += 2;
            continue;
        }
        out += static_cast<char>(std::tolower(c));
    }
    return out;
}

static const std::string& GetWpnDirectory() {
    static std::string dir = []() {
        char buf[MAX_PATH] = {0};
        return GetEnvironmentVariableA("LOCALAPPDATA", buf, MAX_PATH)
                   ? std::string(buf) + "\\Microsoft\\Windows\\Notifications"
                   : std::string();
    }();
    return dir;
}

static sqlite3* OpenWpnDatabase(std::string& outTempDbFile) {
    const auto& sql = WinSqliteApi::Get();
    const std::string& dir = GetWpnDirectory();
    if (!sql.valid || dir.empty())
        return nullptr;

    std::string dbPath = dir + "\\wpndatabase.db";
    std::string uri = "file:" + dbPath + "?mode=ro";
    sqlite3* db = nullptr;

    if (sql.open_v2(uri.c_str(), &db, 0x00000041, nullptr) == 0 && db)
        return db;
    if (db) {
        sql.close(db);
        db = nullptr;
    }
    if (sql.open_v2(dbPath.c_str(), &db, 0x00000001, nullptr) == 0 && db)
        return db;
    if (db) {
        sql.close(db);
        db = nullptr;
    }

    char tmp[MAX_PATH] = {0};
    if (!GetTempPathA(MAX_PATH, tmp))
        return nullptr;
    std::string copyPath = std::string(tmp) + "wpn_vinta_" + std::to_string(GetCurrentThreadId()) + ".db";
    CopyFileA(dbPath.c_str(), copyPath.c_str(), FALSE);
    CopyFileA((dbPath + "-wal").c_str(), (copyPath + "-wal").c_str(), FALSE);

    if (sql.open_v2(copyPath.c_str(), &db, 0x00000001, nullptr) == 0 && db) {
        outTempDbFile = copyPath;
        return db;
    }
    if (db)
        sql.close(db);
    DeleteFileA(copyPath.c_str());
    DeleteFileA((copyPath + "-wal").c_str());
    return nullptr;
}

static void CloseWpnDatabase(sqlite3* db, const std::string& tempFile) {
    const auto& sql = WinSqliteApi::Get();
    if (db && sql.close)
        sql.close(db);
    if (!tempFile.empty()) {
        DeleteFileA(tempFile.c_str());
        DeleteFileA((tempFile + "-wal").c_str());
    }
}

/**
 * @brief Queries wpndatabase.db for image attachments matching a toast notification.
 */
static DbToastInfo ExtractToastFromDb(uint32_t notifId, const std::string& targetTitle, const std::string& aumid) {
    DbToastInfo outInfo;
    const auto& sql = WinSqliteApi::Get();
    std::string tempFile;
    sqlite3* db = OpenWpnDatabase(tempFile);
    if (!db)
        return outInfo;

    bool found = false;
    if (notifId > 0 && sql.bind_int64) {
        sqlite3_stmt* stmt = nullptr;
        if (sql.prepare_v2(db, "SELECT Payload FROM Notification WHERE Id = ? LIMIT 1;", -1, &stmt, nullptr) == 0 &&
            stmt) {
            sql.bind_int64(stmt, 1, static_cast<int64_t>(notifId));
            if (sql.step(stmt) == 100) {
                const void* blob = sql.column_blob(stmt, 0);
                int bytes = sql.column_bytes(stmt, 0);
                if (blob && bytes > 0) {
                    ParseToastXmlPayload(std::string(static_cast<const char*>(blob), bytes), outInfo);
                    found = true;
                }
            }
            sql.finalize(stmt);
        }
    }

    if (!found) {
        const char* q = "SELECT h.PrimaryId, n.Payload FROM Notification n "
                        "LEFT JOIN NotificationHandler h ON n.HandlerId = h.RecordId "
                        "WHERE (n.Type = 'toast' OR n.Type = 'toastCondensed') "
                        "ORDER BY n.ArrivalTime DESC LIMIT 40;";
        sqlite3_stmt* stmt = nullptr;
        if (sql.prepare_v2(db, q, -1, &stmt, nullptr) == 0 && stmt) {
            std::string cTarget = CleanMatchString(targetTitle);
            while (sql.step(stmt) == 100) {
                const char* prim = reinterpret_cast<const char*>(sql.column_text(stmt, 0));
                const void* blob = sql.column_blob(stmt, 1);
                int bytes = sql.column_bytes(stmt, 1);
                if (!blob || bytes <= 0)
                    continue;

                std::string payload(static_cast<const char*>(blob), bytes);
                std::string primStr = prim ? prim : "";
                std::string dbTitle, unusedMsg;
                ParseToastXmlText(payload, dbTitle, unusedMsg);
                std::string cDbTitle = CleanMatchString(dbTitle);

                bool match =
                    !cTarget.empty()
                        ? (cDbTitle.find(cTarget) != std::string::npos ||
                           (!cDbTitle.empty() && cTarget.find(cDbTitle) != std::string::npos) ||
                           CleanMatchString(payload).find(cTarget) != std::string::npos)
                        : (!aumid.empty() && !primStr.empty() &&
                           (primStr.find(aumid) != std::string::npos || aumid.find(primStr) != std::string::npos));

                if (match) {
                    ParseToastXmlPayload(payload, outInfo);
                    if (outInfo.hasAny())
                        break;
                }
            }
            sql.finalize(stmt);
        }
    }

    CloseWpnDatabase(db, tempFile);
    return outInfo;
}

static std::mutex g_appNameMutex;
static std::unordered_map<std::string, std::string> g_appNameCache;

static std::string ParseAumidFallback(const std::string& aumid) {
    if (aumid.empty())
        return "Notification";
    static const std::unordered_map<std::string, std::string> kSystemMap = {
        {"Windows.SystemToast.StartupApp", "Startup Apps"},
        {"Windows.SystemToast.Bthprops", "Bluetooth"},
        {"Windows.SystemToast.BthQuickPair", "Bluetooth"},
        {"Windows.SystemToast.AudioTroubleshooter", "Audio Troubleshooter"},
        {"Windows.SystemToast.AutoPlay", "AutoPlay"},
        {"Windows.SystemToast.BackgroundAccess", "Background Apps"},
        {"Windows.SystemToast.BdeUnlock", "BitLocker"},
        {"Windows.SystemToast.Calling", "Phone Link"},
        {"Windows.SystemToast.Calling.SystemAlertNotification", "Phone Link"},
        {"Windows.Defender", "Windows Security"},
    };
    if (auto it = kSystemMap.find(aumid); it != kSystemMap.end())
        return it->second;

    std::string s = aumid;
    if (size_t p = s.find("::"); p != std::string::npos)
        s = s.substr(0, p);
    if (size_t bang = s.rfind('!'); bang != std::string::npos) {
        std::string appId = s.substr(bang + 1);
        if (FindCI(appId, "app") == 0 || FindCI(appId, "default") == 0 || FindCI(appId, "views") == 0) {
            s = s.substr(0, bang);
            if (size_t u = s.find('_'); u != std::string::npos)
                s = s.substr(0, u);
            if (size_t d = s.rfind('.'); d != std::string::npos)
                s = s.substr(d + 1);
        } else {
            s = appId;
        }
    }
    if (s.size() > 4 && (FindCI(s, ".exe", s.size() - 4) != std::string::npos ||
                         FindCI(s, ".app", s.size() - 4) != std::string::npos)) {
        s.resize(s.size() - 4);
    }
    if (size_t d = s.rfind('.'); d != std::string::npos && d + 1 < s.size())
        s = s.substr(d + 1);
    if (size_t u = s.find('_'); u != std::string::npos && u > 0)
        s = s.substr(0, u);

    std::string out;
    for (size_t i = 0; i < s.size(); ++i) {
        char c = s[i];
        if (i > 0 && std::isupper(static_cast<unsigned char>(c)) &&
            (std::islower(static_cast<unsigned char>(s[i - 1])) ||
             (i + 1 < s.size() && std::islower(static_cast<unsigned char>(s[i + 1])) &&
              std::isupper(static_cast<unsigned char>(s[i - 1]))))) {
            out += ' ';
        }
        out += c;
    }
    if (out.empty())
        return "Notification";
    out[0] = static_cast<char>(std::toupper(static_cast<unsigned char>(out[0])));
    return out;
}

static std::string ResolveAppDisplayName(const std::string& aumid, const std::string& handlerName) {
    if (!handlerName.empty())
        return handlerName;
    if (aumid.empty())
        return "Notification";
    {
        std::lock_guard<std::mutex> lock(g_appNameMutex);
        if (auto it = g_appNameCache.find(aumid); it != g_appNameCache.end())
            return it->second;
    }

    std::string resolved;
    try {
        std::wstring parsingName = L"shell:AppsFolder\\" + Utf8ToUtf16(aumid);
        winrt::com_ptr<IShellItem> item;
        if (SUCCEEDED(SHCreateItemFromParsingName(parsingName.c_str(), nullptr, IID_PPV_ARGS(item.put()))) && item) {
            PWSTR name = nullptr;
            if (SUCCEEDED(item->GetDisplayName(SIGDN_NORMALDISPLAY, &name)) && name) {
                resolved = Utf16ToUtf8(name);
                CoTaskMemFree(name);
            }
        }
    } catch (...) {
    }

    if (resolved.empty()) {
        std::string regName, regIcon;
        if (QueryAppUserModelIdRegistry(aumid, regName, regIcon) && !regName.empty())
            resolved = regName;
    }
    if (resolved.empty())
        resolved = ParseAumidFallback(aumid);

    std::lock_guard<std::mutex> lock(g_appNameMutex);
    if (g_appNameCache.size() > 256)
        g_appNameCache.clear();
    g_appNameCache[aumid] = resolved;
    return resolved;
}

static void EmitToastJson(const char* type, uint32_t id, const std::string& appName, const std::string& aumid,
                          const std::string& icon, const std::string& appLogo, const std::string& title,
                          const std::string& message, const std::string& heroImage) {
    NativeNotificationCallback cb = g_notifCb.load(std::memory_order_acquire);
    if (!cb)
        return;
    std::ostringstream json;
    json << "{\"type\":\"" << type << "\",\"id\":" << id << ",\"appName\":\"" << EscapeJsonString(appName)
         << "\",\"aumid\":\"" << EscapeJsonString(aumid) << "\",\"icon\":\"" << EscapeJsonString(icon)
         << "\",\"appLogo\":\"" << EscapeJsonString(appLogo) << "\",\"title\":\"" << EscapeJsonString(title)
         << "\",\"message\":\"" << EscapeJsonString(message) << "\",\"heroImage\":\"" << EscapeJsonString(heroImage)
         << "\"}";
    cb(json.str().c_str());
}

/**
 * @brief Main worker thread waiting on OS file change and WinRT notification events.
 */
static void NotificationWorker() {
    ApartmentScope apt;
    try {
        auto listener = winrt_notif_mgmt::UserNotificationListener::Current();
        if (listener.RequestAccessAsync().get() != winrt_notif_mgmt::UserNotificationListenerAccessStatus::Allowed)
            return;

        std::unordered_map<uint32_t, std::string> seenRecords;
        try {
            for (auto const& n : listener.GetNotificationsAsync(winrt_notif::NotificationKinds::Toast).get()) {
                seenRecords[n.Id()] = "__INITIAL__";
            }
        } catch (...) {
        }

        std::mutex seenMutex;

        auto dispatchNotification = [&](uint32_t id, const std::string& aumid, const std::string& appName,
                                        const std::string& title, const std::string& message, int64_t arrivalTime,
                                        const std::string& payloadXml) {
            if (id == 0)
                return;
            std::string sig = title + "\x1F" + message + "\x1F" + std::to_string(arrivalTime);
            {
                std::lock_guard<std::mutex> lock(seenMutex);
                if (auto it = seenRecords.find(id);
                    it != seenRecords.end() && (it->second == "__INITIAL__" || it->second == sig))
                    return;
                if (seenRecords.size() > 512)
                    seenRecords.clear();
                seenRecords[id] = sig;
            }

            std::wstring pfn;
            if (size_t bang = aumid.find('!'); bang != std::string::npos)
                pfn = Utf8ToUtf16(aumid.substr(0, bang));

            DbToastInfo dbInfo;
            if (!payloadXml.empty())
                ParseToastXmlPayload(payloadXml, dbInfo);
            else
                dbInfo = ExtractToastFromDb(id, title, aumid);

            std::string icon, appLogo, heroImage;
            auto classifyImg = [&](const std::string& src, bool isOverride) {
                if (src.empty())
                    return;
                ProcessedImage img = ProcessImageSource(Utf8ToUtf16(src), pfn);
                if (img.dataUri.empty())
                    return;
                if (isOverride) {
                    icon = img.dataUri;
                    if (!IsAvatarPath(src))
                        appLogo = img.dataUri;
                } else if ((img.isAvatarSize() || IsAvatarPath(src)) && icon.empty()) {
                    icon = img.dataUri;
                } else if (heroImage.empty()) {
                    heroImage = img.dataUri;
                }
            };

            classifyImg(dbInfo.appLogoOverride, true);
            classifyImg(dbInfo.heroImage, false);
            for (const auto& s : dbInfo.inlineImages)
                classifyImg(s, false);

            if (std::string nativeLogo = ExtractAppIconAsDataUri(appName.c_str(), aumid.c_str(), 48);
                !nativeLogo.empty()) {
                appLogo = nativeLogo;
            }
            if (icon.empty())
                icon = appLogo;

            EmitToastJson("notification", id, appName, aumid, icon, appLogo, title, message, heroImage);

            if (appLogo.empty() || icon.empty() || appName == "Notification") {
                g_activeEnrichTasks.fetch_add(1, std::memory_order_acq_rel);
                std::thread([id, aumid, appName, title, message, pfn, icon, appLogo, heroImage]() mutable {
                    ApartmentScope threadApt;
                    std::string resolvedName = appName, finalLogo = appLogo;
                    for (int attempt = 0; attempt < 2; ++attempt) {
                        if (WaitForSingleObject(g_stopEvent, 200 * (attempt + 1)) == WAIT_OBJECT_0 ||
                            !g_notifRunning.load()) {
                            g_activeEnrichTasks.fetch_sub(1, std::memory_order_acq_rel);
                            return;
                        }
                        try {
                            auto notif = winrt_notif_mgmt::UserNotificationListener::Current().GetNotification(id);
                            if (notif && notif.AppInfo()) {
                                if (auto disp = notif.AppInfo().DisplayInfo()) {
                                    if (std::string wn = Utf16ToUtf8(disp.DisplayName().c_str()); !wn.empty())
                                        resolvedName = wn;
                                    if (finalLogo.empty()) {
                                        if (auto logoRef =
                                                disp.GetLogo(winrt::Windows::Foundation::Size{48.0f, 48.0f})) {
                                            finalLogo = StreamToBase64(logoRef.OpenReadAsync().get());
                                        }
                                    }
                                }
                                if (!notif.AppInfo().PackageFamilyName().empty())
                                    pfn = notif.AppInfo().PackageFamilyName().c_str();
                                break;
                            }
                        } catch (...) {
                        }
                    }
                    if (finalLogo.empty())
                        finalLogo = ExtractAppIconAsDataUri(resolvedName.c_str(), aumid.c_str(), 48);
                    std::string finalIcon = icon;
                    if (finalIcon.empty()) {
                        DbToastInfo retry = ExtractToastFromDb(id, title, aumid);
                        if (!retry.appLogoOverride.empty()) {
                            ProcessedImage img = ProcessImageSource(Utf8ToUtf16(retry.appLogoOverride), pfn);
                            if (!img.dataUri.empty())
                                finalIcon = img.dataUri;
                        }
                    }
                    if (finalIcon.empty())
                        finalIcon = finalLogo;
                    if (g_notifRunning.load() && (!finalLogo.empty() || resolvedName != appName || finalIcon != icon)) {
                        EmitToastJson("notification_update", id, resolvedName, aumid, finalIcon, finalLogo, title,
                                      message, heroImage);
                    }
                    g_activeEnrichTasks.fetch_sub(1, std::memory_order_acq_rel);
                }).detach();
            }
        };

        uint32_t lastProcessedId = 0;
        auto pollWpnDatabase = [&](bool isInitial) {
            const auto& sql = WinSqliteApi::Get();
            std::string tempFile;
            sqlite3* db = OpenWpnDatabase(tempFile);
            if (!db)
                return;

            if (isInitial) {
                sqlite3_stmt* stmt = nullptr;
                if (sql.prepare_v2(db,
                                   "SELECT Id FROM Notification WHERE (Type = 'toast' OR Type = 'toastCondensed') "
                                   "ORDER BY Id DESC LIMIT 50;",
                                   -1, &stmt, nullptr) == 0 &&
                    stmt) {
                    std::lock_guard<std::mutex> lock(seenMutex);
                    while (sql.step(stmt) == 100) {
                        uint32_t sid = static_cast<uint32_t>(sql.column_int64(stmt, 0));
                        if (sid > lastProcessedId)
                            lastProcessedId = sid;
                        seenRecords[sid] = "__INITIAL__";
                    }
                    sql.finalize(stmt);
                }
                CloseWpnDatabase(db, tempFile);
                return;
            }

            const char* q = "SELECT n.Id, h.PrimaryId, n.ArrivalTime, n.Payload, ha.AssetValue "
                            "FROM Notification n "
                            "LEFT JOIN NotificationHandler h ON n.HandlerId = h.RecordId "
                            "LEFT JOIN HandlerAssets ha ON n.HandlerId = ha.HandlerId AND ha.AssetKey = 'DisplayName' "
                            "WHERE (n.Type = 'toast' OR n.Type = 'toastCondensed') AND n.Id > ? ORDER BY n.Id ASC;";
            sqlite3_stmt* stmt = nullptr;
            if (sql.prepare_v2(db, q, -1, &stmt, nullptr) == 0 && stmt) {
                if (sql.bind_int64)
                    sql.bind_int64(stmt, 1, static_cast<int64_t>(lastProcessedId));
                while (sql.step(stmt) == 100) {
                    uint32_t id = static_cast<uint32_t>(sql.column_int64(stmt, 0));
                    if (id > lastProcessedId)
                        lastProcessedId = id;
                    const char* aPtr = reinterpret_cast<const char*>(sql.column_text(stmt, 1));
                    std::string aumid = aPtr ? aPtr : "";
                    int64_t arrival = sql.column_int64(stmt, 2);
                    const void* blob = sql.column_blob(stmt, 3);
                    int bytes = sql.column_bytes(stmt, 3);
                    std::string payload = (blob && bytes > 0) ? std::string(static_cast<const char*>(blob), bytes) : "";
                    const char* dPtr = reinterpret_cast<const char*>(sql.column_text(stmt, 4));
                    std::string dispName = dPtr ? dPtr : "";

                    std::string title, message;
                    ParseToastXmlText(payload, title, message);
                    dispatchNotification(id, aumid, ResolveAppDisplayName(aumid, dispName), title, message, arrival,
                                         payload);
                }
                sql.finalize(stmt);
            }
            CloseWpnDatabase(db, tempFile);
        };

        pollWpnDatabase(true);

        winrt::event_token notifToken{};
        try {
            notifToken = listener.NotificationChanged([](auto const&, auto const& args) {
                if (args.ChangeKind() == winrt_notif::UserNotificationChangedKind::Added && g_wakeEvent) {
                    SetEvent(g_wakeEvent);
                }
            });
        } catch (...) {
        }

        const std::string& notifDir = GetWpnDirectory();
        HANDLE hChange = notifDir.empty()
                             ? INVALID_HANDLE_VALUE
                             : FindFirstChangeNotificationA(notifDir.c_str(), FALSE,
                                                            FILE_NOTIFY_CHANGE_LAST_WRITE | FILE_NOTIFY_CHANGE_SIZE |
                                                                FILE_NOTIFY_CHANGE_FILE_NAME);

        while (g_notifRunning.load(std::memory_order_acquire)) {
            HANDLE handles[3] = {g_stopEvent, g_wakeEvent, hChange};
            DWORD count = (hChange != INVALID_HANDLE_VALUE) ? 3 : 2;
            DWORD waitRes = WaitForMultipleObjects(count, handles, FALSE, INFINITE);
            if (waitRes == WAIT_OBJECT_0 || !g_notifRunning.load(std::memory_order_acquire))
                break;
            if (hChange != INVALID_HANDLE_VALUE && waitRes == WAIT_OBJECT_0 + 2) {
                FindNextChangeNotification(hChange);
                if (WaitForSingleObject(g_stopEvent, 80) == WAIT_OBJECT_0 ||
                    !g_notifRunning.load(std::memory_order_acquire)) {
                    break;
                }
            }
            try {
                pollWpnDatabase(false);
            } catch (...) {
            }
        }

        if (notifToken) {
            try {
                listener.NotificationChanged(notifToken);
            } catch (...) {
            }
        }
        if (hChange != INVALID_HANDLE_VALUE)
            FindCloseChangeNotification(hChange);
    } catch (...) {
    }
}

void StartNotificationService(NativeNotificationCallback cb) {
    if (g_notifRunning.exchange(true, std::memory_order_acq_rel))
        return;
    if (g_notifThread.joinable())
        g_notifThread.join();
    g_notifCb.store(cb, std::memory_order_release);
    if (!g_stopEvent)
        g_stopEvent = CreateEventW(nullptr, TRUE, FALSE, nullptr);
    else
        ResetEvent(g_stopEvent);
    if (!g_wakeEvent)
        g_wakeEvent = CreateEventW(nullptr, FALSE, FALSE, nullptr);
    g_notifThread = std::thread(NotificationWorker);
}

void StopNotificationService() {
    if (!g_notifRunning.exchange(false, std::memory_order_acq_rel))
        return;
    g_notifCb.store(nullptr, std::memory_order_release);
    if (g_stopEvent)
        SetEvent(g_stopEvent);
    if (g_notifThread.joinable())
        g_notifThread.join();
    while (g_activeEnrichTasks.load(std::memory_order_acquire) > 0) {
        std::this_thread::sleep_for(std::chrono::milliseconds(10));
    }
    if (g_stopEvent) {
        CloseHandle(g_stopEvent);
        g_stopEvent = nullptr;
    }
    if (g_wakeEvent) {
        CloseHandle(g_wakeEvent);
        g_wakeEvent = nullptr;
    }
}

static std::wstring GetVintaIconPath() {
    wchar_t appData[MAX_PATH] = {0};
    if (GetEnvironmentVariableW(L"APPDATA", appData, MAX_PATH) > 0) {
        std::wstring path = std::wstring(appData) + L"\\FreeMatrix Studios\\Vinta\\appicon.png";
        DWORD attr = GetFileAttributesW(path.c_str());
        if (attr != INVALID_FILE_ATTRIBUTES && !(attr & FILE_ATTRIBUTE_DIRECTORY))
            return path;
    }
    return L"";
}

static std::wstring EscapeXml(const std::wstring& src) {
    std::wstring out;
    out.reserve(src.size() + 16);
    for (wchar_t c : src) {
        switch (c) {
        case L'&':
            out += L"&amp;";
            break;
        case L'<':
            out += L"&lt;";
            break;
        case L'>':
            out += L"&gt;";
            break;
        case L'\"':
            out += L"&quot;";
            break;
        case L'\'':
            out += L"&apos;";
            break;
        default:
            out += c;
            break;
        }
    }
    return out;
}

static void EnsureAumidRegistered() {
    static std::once_flag once;
    std::call_once(once, []() {
        const wchar_t* aumid = L"FreeMatrixStudios.Vinta";
        SetCurrentProcessExplicitAppUserModelID(aumid);

        std::wstring iconPath = GetVintaIconPath();
        wchar_t exePath[MAX_PATH] = {0};
        GetModuleFileNameW(nullptr, exePath, MAX_PATH);
        const wchar_t* icon = !iconPath.empty() ? iconPath.c_str() : exePath;
        const wchar_t* name = L"Vinta";

        auto setSz = [](HKEY k, const wchar_t* vName, const wchar_t* val) {
            RegSetValueExW(k, vName, 0, REG_SZ, reinterpret_cast<const BYTE*>(val),
                           static_cast<DWORD>((std::wcslen(val) + 1) * sizeof(wchar_t)));
        };

        HKEY hKey = nullptr;
        if (RegCreateKeyExW(HKEY_CURRENT_USER, L"Software\\Classes\\AppUserModelId\\FreeMatrixStudios.Vinta", 0,
                            nullptr, REG_OPTION_NON_VOLATILE, KEY_SET_VALUE, nullptr, &hKey,
                            nullptr) == ERROR_SUCCESS) {
            setSz(hKey, L"DisplayName", name);
            setSz(hKey, L"IconUri", icon);
            setSz(hKey, L"IconBackgroundColor", L"0");
            RegCloseKey(hKey);
        }

        if (RegCreateKeyExW(
                HKEY_CURRENT_USER,
                L"Software\\Microsoft\\Windows\\CurrentVersion\\Notifications\\Settings\\FreeMatrixStudios.Vinta", 0,
                nullptr, REG_OPTION_NON_VOLATILE, KEY_SET_VALUE, nullptr, &hKey, nullptr) == ERROR_SUCCESS) {
            DWORD one = 1;
            setSz(hKey, L"DisplayName", name);
            setSz(hKey, L"IconUri", icon);
            RegSetValueExW(hKey, L"ShowInActionCenter", 0, REG_DWORD, reinterpret_cast<const BYTE*>(&one),
                           sizeof(DWORD));
            RegSetValueExW(hKey, L"Enabled", 0, REG_DWORD, reinterpret_cast<const BYTE*>(&one), sizeof(DWORD));
            RegCloseKey(hKey);
        }

        PWSTR progPath = nullptr;
        if (SUCCEEDED(SHGetKnownFolderPath(FOLDERID_Programs, 0, nullptr, &progPath))) {
            std::wstring lnk = std::wstring(progPath) + L"\\Vinta.lnk";
            CoTaskMemFree(progPath);
            if (exePath[0] != L'\0') {
                Microsoft::WRL::ComPtr<IShellLinkW> sl;
                if (SUCCEEDED(CoCreateInstance(CLSID_ShellLink, nullptr, CLSCTX_INPROC_SERVER, IID_PPV_ARGS(&sl)))) {
                    sl->SetPath(exePath);
                    sl->SetIconLocation(icon, 0);
                    Microsoft::WRL::ComPtr<IPropertyStore> ps;
                    if (SUCCEEDED(sl.As(&ps))) {
                        PROPVARIANT pv;
                        if (SUCCEEDED(InitPropVariantFromString(aumid, &pv))) {
                            ps->SetValue(PKEY_AppUserModel_ID, pv);
                            ps->Commit();
                            PropVariantClear(&pv);
                        }
                    }
                    Microsoft::WRL::ComPtr<IPersistFile> pf;
                    if (SUCCEEDED(sl.As(&pf)))
                        pf->Save(lnk.c_str(), TRUE);
                }
            }
        }
    });
}

int SendToastNotification(const wchar_t* title, const wchar_t* message) {
    try {
        if (!title && !message)
            return 0;
        ApartmentScope apt;
        EnsureAumidRegistered();

        std::wstring iconUri = GetVintaIconPath();
        for (wchar_t& ch : iconUri)
            if (ch == L'\\')
                ch = L'/';

        std::wstring xml = L"<toast><visual><binding template=\"ToastGeneric\"><text>" +
                           EscapeXml(title ? title : L"Vinta") + L"</text>";
        if (message && message[0] != L'\0')
            xml += L"<text>" + EscapeXml(message) + L"</text>";
        if (!iconUri.empty())
            xml += L"<image placement=\"appLogoOverride\" hint-crop=\"circle\" src=\"file:///" + iconUri + L"\"/>";
        xml += L"</binding></visual></toast>";

        winrt::Windows::Data::Xml::Dom::XmlDocument doc;
        doc.LoadXml(xml);
        winrt_notif::ToastNotificationManager::CreateToastNotifier(L"FreeMatrixStudios.Vinta")
            .Show(winrt_notif::ToastNotification(doc));
        return 1;
    } catch (...) {
        return 0;
    }
}