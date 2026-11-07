/**
 * @file notification_service.cpp
 * @brief UserNotificationListener toast notification interception implementation.
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

#include <atomic>
#include <condition_variable>
#include <mutex>
#include <regex>
#include <sstream>
#include <thread>
#include <unordered_set>

namespace winrt_notif = winrt::Windows::UI::Notifications;
namespace winrt_notif_mgmt = winrt::Windows::UI::Notifications::Management;

static NativeNotificationCallback g_notifCb = nullptr;
static std::atomic<bool> g_notifRunning{false};
static std::thread g_notifThread;
static std::mutex g_notifWakeMutex;
static std::condition_variable g_notifWakeCv;
static std::atomic<bool> g_notifWakePending{false};

typedef struct sqlite3 sqlite3;
typedef struct sqlite3_stmt sqlite3_stmt;

typedef int (*fn_sqlite3_open_v2)(const char *filename, sqlite3 **ppDb, int flags, const char *zVfs);
typedef int (*fn_sqlite3_prepare_v2)(sqlite3 *db, const char *zSql, int nByte, sqlite3_stmt **ppStmt, const char **pzTail);
typedef int (*fn_sqlite3_bind_int64)(sqlite3_stmt *pStmt, int iCol, int64_t val);
typedef int (*fn_sqlite3_step)(sqlite3_stmt *pStmt);
typedef const unsigned char *(*fn_sqlite3_column_text)(sqlite3_stmt *pStmt, int iCol);
typedef const void *(*fn_sqlite3_column_blob)(sqlite3_stmt *pStmt, int iCol);
typedef int (*fn_sqlite3_column_bytes)(sqlite3_stmt *pStmt, int iCol);
typedef int64_t (*fn_sqlite3_column_int64)(sqlite3_stmt *pStmt, int iCol);
typedef int (*fn_sqlite3_finalize)(sqlite3_stmt *pStmt);
typedef int (*fn_sqlite3_close)(sqlite3 *db);

struct DbToastInfo {
    std::string appLogoOverride;
    std::string heroImage;
    std::vector<std::string> inlineImages;
};

static void ParseToastXmlPayload(const std::string& payload, DbToastInfo& outInfo) {
    if (payload.empty()) return;
    std::regex imgRegex("<image[^>]+>", std::regex::icase);
    std::regex srcRegex("src\\s*=\\s*[\"']([^\"']+)[\"']", std::regex::icase);
    std::regex placementRegex("placement\\s*=\\s*[\"']([^\"']+)[\"']", std::regex::icase);
    std::regex cropRegex("hint-crop\\s*=\\s*[\"']circle[\"']", std::regex::icase);

    auto imgBegin = std::sregex_iterator(payload.begin(), payload.end(), imgRegex);
    auto imgEnd = std::sregex_iterator();
    for (auto it = imgBegin; it != imgEnd; ++it) {
        std::string imgTag = (*it).str();
        std::smatch smSrc, smPlacement;
        std::string src, placement;
        if (std::regex_search(imgTag, smSrc, srcRegex)) src = smSrc[1].str();
        if (std::regex_search(imgTag, smPlacement, placementRegex)) placement = smPlacement[1].str();

        if (src.empty()) continue;

        bool hasCircleCrop = std::regex_search(imgTag, cropRegex);
        bool isAvatarHint = src.find("avatar") != std::string::npos ||
                            src.find("avatars") != std::string::npos ||
                            src.find("profile") != std::string::npos ||
                            src.find("user") != std::string::npos;

        if (placement == "appLogoOverride" || placement == "applogo" || hasCircleCrop || isAvatarHint) {
            // Circle crop, appLogoOverride, or avatar path explicitly designates a contact profile avatar.
            if (outInfo.appLogoOverride.empty()) {
                outInfo.appLogoOverride = src;
            }
        } else if (placement == "hero") {
            outInfo.heroImage = src;
        } else {
            outInfo.inlineImages.push_back(src);
        }
    }
}

static std::string DecodeXmlEntities(const std::string& in) {
    std::string out = in;
    size_t pos = 0;
    while ((pos = out.find("&amp;", pos)) != std::string::npos) { out.replace(pos, 5, "&"); pos += 1; }
    pos = 0;
    while ((pos = out.find("&lt;", pos)) != std::string::npos) { out.replace(pos, 4, "<"); pos += 1; }
    pos = 0;
    while ((pos = out.find("&gt;", pos)) != std::string::npos) { out.replace(pos, 4, ">"); pos += 1; }
    pos = 0;
    while ((pos = out.find("&quot;", pos)) != std::string::npos) { out.replace(pos, 6, "\""); pos += 1; }
    pos = 0;
    while ((pos = out.find("&apos;", pos)) != std::string::npos) { out.replace(pos, 6, "'"); pos += 1; }
    return out;
}

static void ParseToastXmlText(const std::string& payload, std::string& outTitle, std::string& outMessage) {
    if (payload.empty()) return;
    std::regex textRegex("<text[^>]*>([\\s\\S]*?)</text>", std::regex::icase);
    auto textBegin = std::sregex_iterator(payload.begin(), payload.end(), textRegex);
    auto textEnd = std::sregex_iterator();
    int idx = 0;
    for (auto it = textBegin; it != textEnd; ++it, ++idx) {
        std::string raw = (*it)[1].str();
        if (raw.rfind("<![CDATA[", 0) == 0 && raw.size() >= 12 && raw.substr(raw.size() - 3) == "]]>") {
            raw = raw.substr(9, raw.size() - 12);
        } else {
            raw = DecodeXmlEntities(raw);
        }
        if (idx == 0) {
            outTitle = raw;
        } else {
            if (!outMessage.empty()) outMessage += "\n";
            outMessage += raw;
        }
    }
}

static DbToastInfo ExtractToastFromDb(uint32_t notifId, const std::string& targetTitle, const std::string& aumid) {
    DbToastInfo outInfo;
    HMODULE hSqlite = LoadLibraryW(L"winsqlite3.dll");
    if (!hSqlite) return outInfo;

    auto p_open = (fn_sqlite3_open_v2)GetProcAddress(hSqlite, "sqlite3_open_v2");
    auto p_prepare = (fn_sqlite3_prepare_v2)GetProcAddress(hSqlite, "sqlite3_prepare_v2");
    auto p_bind_int64 = (fn_sqlite3_bind_int64)GetProcAddress(hSqlite, "sqlite3_bind_int64");
    auto p_step = (fn_sqlite3_step)GetProcAddress(hSqlite, "sqlite3_step");
    auto p_text = (fn_sqlite3_column_text)GetProcAddress(hSqlite, "sqlite3_column_text");
    auto p_blob = (fn_sqlite3_column_blob)GetProcAddress(hSqlite, "sqlite3_column_blob");
    auto p_bytes = (fn_sqlite3_column_bytes)GetProcAddress(hSqlite, "sqlite3_column_bytes");
    auto p_finalize = (fn_sqlite3_finalize)GetProcAddress(hSqlite, "sqlite3_finalize");
    auto p_close = (fn_sqlite3_close)GetProcAddress(hSqlite, "sqlite3_close");

    if (!p_open || !p_prepare || !p_step || !p_finalize || !p_close) return outInfo;

    char localAppData[MAX_PATH];
    if (GetEnvironmentVariableA("LOCALAPPDATA", localAppData, MAX_PATH) == 0) return outInfo;
    std::string dbPath = std::string(localAppData) + "\\Microsoft\\Windows\\Notifications\\wpndatabase.db";

    sqlite3 *db = nullptr;
    std::string tempDbFile;
    int rc = p_open(dbPath.c_str(), &db, 0x00000001, nullptr);
    if (rc != 0) {
        char tempPath[MAX_PATH];
        GetTempPathA(MAX_PATH, tempPath);
        std::string copyPath = std::string(tempPath) + "wpn_vinta_" + std::to_string(GetCurrentThreadId()) + ".db";
        std::string walPath = dbPath + "-wal";
        std::string copyWal = copyPath + "-wal";
        CopyFileA(dbPath.c_str(), copyPath.c_str(), FALSE);
        CopyFileA(walPath.c_str(), copyWal.c_str(), FALSE);
        tempDbFile = copyPath;
        if (p_open(copyPath.c_str(), &db, 0x00000001, nullptr) != 0) {
            DeleteFileA(copyPath.c_str());
            DeleteFileA(copyWal.c_str());
            return outInfo;
        }
    }

    bool found = false;

    if (notifId > 0 && p_bind_int64) {
        const char *sqlId = "SELECT n.Payload FROM Notification n WHERE n.Id = ? LIMIT 1;";
        sqlite3_stmt *stmtId = nullptr;
        if (p_prepare(db, sqlId, -1, &stmtId, nullptr) == 0) {
            p_bind_int64(stmtId, 1, static_cast<int64_t>(notifId));
            if (p_step(stmtId) == 100) {
                const void *blob = p_blob(stmtId, 0);
                int blobSize = p_bytes(stmtId, 0);
                if (blob && blobSize > 0) {
                    std::string payload((const char *)blob, blobSize);
                    ParseToastXmlPayload(payload, outInfo);
                    found = true;
                }
            }
            p_finalize(stmtId);
        }
    }

    auto cleanStr = [](const std::string& in) {
        std::string out;
        for (size_t i = 0; i < in.size(); ++i) {
            unsigned char c = in[i];
            if (c < 0x20) continue;
            // Strip directional isolate bytes U+2068 and U+2069
            if (c == 0xE2 && i + 2 < in.size() && static_cast<unsigned char>(in[i+1]) == 0x81 &&
                (static_cast<unsigned char>(in[i+2]) == 0xA8 || static_cast<unsigned char>(in[i+2]) == 0xA9)) {
                i += 2;
                continue;
            }
            out += tolower(c);
        }
        return out;
    };

    if (!found) {
        const char *sql = 
            "SELECT h.PrimaryId, n.Payload "
            "FROM Notification n "
            "LEFT JOIN NotificationHandler h ON n.HandlerId = h.RecordId "
            "WHERE (n.Type = 'toast' OR n.Type = 'toastCondensed') "
            "ORDER BY n.ArrivalTime DESC LIMIT 100;";

        sqlite3_stmt *stmt = nullptr;
        if (p_prepare(db, sql, -1, &stmt, nullptr) == 0) {
            while (p_step(stmt) == 100) {
                const char *prim = (const char *)p_text(stmt, 0);
                const void *blob = p_blob(stmt, 1);
                int blobSize = p_bytes(stmt, 1);
                std::string payload;
                if (blob && blobSize > 0) {
                    payload = std::string((const char *)blob, blobSize);
                }

                std::string primStr = prim ? prim : "";
                
                std::string dbTitle;
                std::regex textRegex("<text[^>]*>([\\s\\S]*?)</text>", std::regex::icase);
                auto textBegin = std::sregex_iterator(payload.begin(), payload.end(), textRegex);
                auto textEnd = std::sregex_iterator();
                if (textBegin != textEnd) {
                    dbTitle = (*textBegin)[1].str();
                }

                std::string cDbTitle = cleanStr(dbTitle);
                std::string cTargetTitle = cleanStr(targetTitle);
                std::string cPayload = cleanStr(payload);

                bool match = false;
                if (!cTargetTitle.empty()) {
                    if (cDbTitle.find(cTargetTitle) != std::string::npos ||
                        cTargetTitle.find(cDbTitle) != std::string::npos ||
                        cPayload.find(cTargetTitle) != std::string::npos) {
                        match = true;
                    }
                } else if (!aumid.empty() && !primStr.empty()) {
                    if (primStr.find(aumid) != std::string::npos || aumid.find(primStr) != std::string::npos) {
                        match = true;
                    }
                }

                if (match) {
                    ParseToastXmlPayload(payload, outInfo);
                    if (!outInfo.appLogoOverride.empty() || !outInfo.heroImage.empty() || !outInfo.inlineImages.empty()) {
                        found = true;
                        break;
                    }
                }
            }
            p_finalize(stmt);
        }
    }

    if (!found && !targetTitle.empty()) {
        std::vector<std::string> searchPaths;
        if (!tempDbFile.empty()) {
            searchPaths.push_back(tempDbFile + "-wal");
            searchPaths.push_back(tempDbFile);
        } else {
            searchPaths.push_back(dbPath + "-wal");
            searchPaths.push_back(dbPath);
        }

        std::string needle = targetTitle;
        size_t spacePos = targetTitle.find_first_of(" (#\r\n\t");
        std::string shortNeedle = (spacePos != std::string::npos && spacePos >= 3) ? targetTitle.substr(0, spacePos) : "";

        for (const auto& searchFile : searchPaths) {
            HANDLE hFile = CreateFileA(searchFile.c_str(), GENERIC_READ, FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
            if (hFile == INVALID_HANDLE_VALUE) continue;

            LARGE_INTEGER fsize;
            if (GetFileSizeEx(hFile, &fsize) && fsize.QuadPart > 0 && fsize.QuadPart < 50 * 1024 * 1024) {
                std::string rawData(static_cast<size_t>(fsize.QuadPart), '\0');
                DWORD bytesRead = 0;
                if (ReadFile(hFile, &rawData[0], static_cast<DWORD>(rawData.size()), &bytesRead, NULL) && bytesRead > 0) {
                    size_t pos = rawData.rfind(needle);
                    if (pos == std::string::npos && !shortNeedle.empty()) {
                        pos = rawData.rfind(shortNeedle);
                    }

                    if (pos != std::string::npos) {
                        size_t bPos = rawData.rfind("<binding", pos);
                        size_t tPos = rawData.rfind("<toast", pos);
                        size_t toastStart = std::string::npos;
                        if (bPos != std::string::npos && tPos != std::string::npos) {
                            toastStart = (bPos > tPos) ? bPos : tPos;
                        } else if (bPos != std::string::npos) {
                            toastStart = bPos;
                        } else {
                            toastStart = tPos;
                        }

                        if (toastStart != std::string::npos && pos - toastStart < 2500) {
                            size_t bEnd = rawData.find("</binding>", pos);
                            size_t tEnd = rawData.find("</toast>", pos);
                            size_t toastEnd = std::string::npos;
                            if (bEnd != std::string::npos && tEnd != std::string::npos) {
                                toastEnd = (bEnd < tEnd) ? bEnd + 10 : tEnd + 8;
                            } else if (bEnd != std::string::npos) {
                                toastEnd = bEnd + 10;
                            } else if (tEnd != std::string::npos) {
                                toastEnd = tEnd + 8;
                            }

                            if (toastEnd != std::string::npos && toastEnd > toastStart) {
                                std::string payload = rawData.substr(toastStart, toastEnd - toastStart);
                                ParseToastXmlPayload(payload, outInfo);
                                if (!outInfo.appLogoOverride.empty() || !outInfo.heroImage.empty() || !outInfo.inlineImages.empty()) {
                                    found = true;
                                }
                            }
                        }
                    }
                }
            }
            CloseHandle(hFile);
            if (found) break;
        }
    }

    p_close(db);
    if (!tempDbFile.empty()) {
        DeleteFileA(tempDbFile.c_str());
        DeleteFileA((tempDbFile + "-wal").c_str());
    }
    return outInfo;
}

/**
 * Extracts toast payload information directly from wpndatabase.db
 * without artificial sleep delays to ensure instant notification dispatch.
 */
static DbToastInfo ExtractToastFast(uint32_t notifId, const std::string& targetTitle, const std::string& aumid) {
    return ExtractToastFromDb(notifId, targetTitle, aumid);
}

static std::mutex g_appNameMutex;
static std::unordered_map<std::string, std::string> g_appNameCache;

static std::string FormatPascalCase(const std::string& in) {
    if (in.empty()) return in;
    std::string out;
    for (size_t i = 0; i < in.size(); ++i) {
        char c = in[i];
        if (i > 0 && isupper(static_cast<unsigned char>(c)) &&
            (islower(static_cast<unsigned char>(in[i - 1])) ||
             (i + 1 < in.size() && islower(static_cast<unsigned char>(in[i + 1])) && isupper(static_cast<unsigned char>(in[i - 1]))))) {
            out += ' ';
        }
        out += c;
    }
    return out;
}

static std::string ParseAumidFallback(const std::string& aumid) {
    if (aumid.empty()) return "Notification";

    if (aumid == "Windows.SystemToast.StartupApp") return "Startup Apps";
    if (aumid == "Windows.SystemToast.Bthprops" || aumid == "Windows.SystemToast.BthQuickPair") return "Bluetooth";
    if (aumid == "Windows.SystemToast.AudioTroubleshooter") return "Audio Troubleshooter";
    if (aumid == "Windows.SystemToast.AutoPlay") return "AutoPlay";
    if (aumid == "Windows.SystemToast.BackgroundAccess") return "Background Apps";
    if (aumid == "Windows.SystemToast.BdeUnlock") return "BitLocker";
    if (aumid == "Windows.SystemToast.Calling" || aumid == "Windows.SystemToast.Calling.SystemAlertNotification") return "Phone Link";
    if (aumid == "Windows.Defender") return "Windows Security";

    std::string s = aumid;

    size_t dColon = s.find("::");
    if (dColon != std::string::npos) s = s.substr(0, dColon);

    size_t bangIdx = s.rfind('!');
    if (bangIdx != std::string::npos) {
        std::string appId = s.substr(bangIdx + 1);
        std::string lowerAppId = appId;
        for (auto& c : lowerAppId) c = tolower(static_cast<unsigned char>(c));

        if (lowerAppId == "app" || lowerAppId == "application" || lowerAppId == "default" || lowerAppId == "views") {
            std::string pkg = s.substr(0, bangIdx);
            size_t uscore = pkg.find('_');
            if (uscore != std::string::npos) pkg = pkg.substr(0, uscore);
            size_t dot = pkg.rfind('.');
            if (dot != std::string::npos) pkg = pkg.substr(dot + 1);
            s = pkg;
        } else {
            s = appId;
        }
    }

    if (s.size() > 4) {
        std::string ext = s.substr(s.size() - 4);
        for (auto& c : ext) c = tolower(static_cast<unsigned char>(c));
        if (ext == ".exe" || ext == ".app") {
            s = s.substr(0, s.size() - 4);
        }
    }

    size_t dot = s.rfind('.');
    if (dot != std::string::npos && dot + 1 < s.size()) {
        s = s.substr(dot + 1);
    }

    size_t uscore = s.find('_');
    if (uscore != std::string::npos && uscore > 0) {
        s = s.substr(0, uscore);
    }

    s = FormatPascalCase(s);

    if (s.empty()) return "Notification";
    s[0] = toupper(static_cast<unsigned char>(s[0]));
    return s;
}

static std::string ResolveAppDisplayName(const std::string& aumid, const std::string& handlerDisplayName) {
    if (!handlerDisplayName.empty()) {
        return handlerDisplayName;
    }
    if (aumid.empty()) return "Notification";

    {
        std::lock_guard<std::mutex> lock(g_appNameMutex);
        auto it = g_appNameCache.find(aumid);
        if (it != g_appNameCache.end()) {
            return it->second;
        }
    }

    std::string resolved;

    try {
        std::wstring wAumid = Utf8ToUtf16(aumid);
        std::wstring parsingName = L"shell:AppsFolder\\" + wAumid;
        winrt::com_ptr<IShellItem> pItem;
        HRESULT hr = SHCreateItemFromParsingName(parsingName.c_str(), NULL, IID_PPV_ARGS(pItem.put()));
        if (SUCCEEDED(hr) && pItem) {
            PWSTR pszName = nullptr;
            hr = pItem->GetDisplayName(SIGDN_NORMALDISPLAY, &pszName);
            if (SUCCEEDED(hr) && pszName) {
                resolved = Utf16ToUtf8(pszName);
                CoTaskMemFree(pszName);
            }
        }
    } catch (...) {}

    // Check Windows Registry AppUserModelId
    if (resolved.empty()) {
        std::string regName, regIcon;
        if (QueryAppUserModelIdRegistry(aumid, regName, regIcon) && !regName.empty()) {
            resolved = regName;
        }
    }

    if (resolved.empty()) {
        resolved = ParseAumidFallback(aumid);
    }

    if (!resolved.empty()) {
        std::lock_guard<std::mutex> lock(g_appNameMutex);
        g_appNameCache[aumid] = resolved;
    }

    return resolved;
}

static void NotificationWorker() {
    ApartmentScope apt;
    try {
        auto listener = winrt_notif_mgmt::UserNotificationListener::Current();
        auto access = listener.RequestAccessAsync().get();
        if (access != winrt_notif_mgmt::UserNotificationListenerAccessStatus::Allowed) {
            return;
        }

        struct NotifRecord {
            std::string signature;
        };
        std::unordered_map<uint32_t, NotifRecord> seenRecords;
        try {
            auto initialNotifs = listener.GetNotificationsAsync(winrt_notif::NotificationKinds::Toast).get();
            for (auto const& notif : initialNotifs) {
                seenRecords[notif.Id()] = { "__INITIAL__" };
            }
        } catch (...) {}

        std::mutex seenMutex;

        auto dispatchNotification = [&](uint32_t id, const std::string& aumid, const std::string& appName, const std::string& title, const std::string& message, int64_t arrivalTime, const std::string& payloadXml) {
            if (id == 0) return;
            std::string currentSig = title + "\x1F" + message + "\x1F" + std::to_string(arrivalTime);
            {
                std::lock_guard<std::mutex> lock(seenMutex);
                auto it = seenRecords.find(id);
                if (it != seenRecords.end()) {
                    if (it->second.signature == "__INITIAL__" || it->second.signature == currentSig) {
                        return;
                    }
                }
                seenRecords[id] = { currentSig };
            }

            std::wstring packageFamilyName;
            size_t bangPos = aumid.find('!');
            if (bangPos != std::string::npos) {
                packageFamilyName = Utf8ToUtf16(aumid.substr(0, bangPos));
            }

            DbToastInfo dbInfo;
            if (!payloadXml.empty()) {
                ParseToastXmlPayload(payloadXml, dbInfo);
            } else {
                dbInfo = ExtractToastFast(id, title, aumid);
            }

            std::string icon;
            std::string appLogo;
            std::string heroImage;

            if (!dbInfo.appLogoOverride.empty()) {
                ProcessedImage appLogoImg = ProcessImageSource(Utf8ToUtf16(dbInfo.appLogoOverride), packageFamilyName);
                if (!appLogoImg.dataUri.empty()) {
                    bool isAvatarHint = dbInfo.appLogoOverride.find("avatar") != std::string::npos ||
                                        dbInfo.appLogoOverride.find("avatars") != std::string::npos ||
                                        dbInfo.appLogoOverride.find("profile") != std::string::npos ||
                                        dbInfo.appLogoOverride.find("user") != std::string::npos;
                    if (isAvatarHint) {
                        icon = appLogoImg.dataUri;
                    } else {
                        icon = appLogoImg.dataUri;
                        appLogo = appLogoImg.dataUri;
                    }
                    icon = appLogoImg.dataUri;
                }
            }

            if (!dbInfo.heroImage.empty()) {
                ProcessedImage heroImg = ProcessImageSource(Utf8ToUtf16(dbInfo.heroImage), packageFamilyName);
                if (!heroImg.dataUri.empty()) {
                    bool isAvatarHint = dbInfo.heroImage.find("avatar") != std::string::npos ||
                                        dbInfo.heroImage.find("avatars") != std::string::npos ||
                                        dbInfo.heroImage.find("profile") != std::string::npos ||
                                        dbInfo.heroImage.find("user") != std::string::npos;
                    if ((heroImg.isAvatarSize() || isAvatarHint) && icon.empty()) {
                        icon = heroImg.dataUri;
                    } else {
                        heroImage = heroImg.dataUri;
                    }
                }
            }

            for (const auto& inlineSrc : dbInfo.inlineImages) {
                ProcessedImage inImg = ProcessImageSource(Utf8ToUtf16(inlineSrc), packageFamilyName);
                if (!inImg.dataUri.empty()) {
                    bool isAvatarHint = inlineSrc.find("avatar") != std::string::npos ||
                                        inlineSrc.find("avatars") != std::string::npos ||
                                        inlineSrc.find("profile") != std::string::npos ||
                                        inlineSrc.find("user") != std::string::npos;
                    if ((inImg.isAvatarSize() || isAvatarHint) && icon.empty()) {
                        icon = inImg.dataUri;
                    } else if (heroImage.empty()) {
                        heroImage = inImg.dataUri;
                    }
                }
            }

            // Resolve authentic parent application icon directly via Win32 Shell / Registry
            std::string nativeAppLogo = ExtractAppIconAsDataUri(appName.c_str(), aumid.c_str(), 48);
            if (!nativeAppLogo.empty()) {
                appLogo = nativeAppLogo;
            }
            if (icon.empty()) {
                icon = appLogo;
            }

            // Immediately emit toast with 0ms delay without blocking on WinRT RPC
            if (g_notifCb) {
                std::ostringstream json;
                json << "{\"type\":\"notification\","
                     << "\"id\":" << id << ","
                     << "\"appName\":\"" << EscapeJsonString(appName) << "\","
                     << "\"aumid\":\"" << EscapeJsonString(aumid) << "\","
                     << "\"icon\":\"" << EscapeJsonString(icon) << "\","
                     << "\"appLogo\":\"" << EscapeJsonString(appLogo) << "\","
                     << "\"title\":\"" << EscapeJsonString(title) << "\","
                     << "\"message\":\"" << EscapeJsonString(message) << "\","
                     << "\"heroImage\":\"" << EscapeJsonString(heroImage) << "\"}";
                g_notifCb(json.str().c_str());
            }

            // Asynchronously resolve WinRT display name, app logo, and late avatar in background
            std::thread([id, aumid, appName, title, message, packageFamilyName, icon, appLogo, heroImage]() {
                ApartmentScope apt;
                std::string resolvedName = appName;
                std::string finalAppLogo = appLogo;
                std::wstring pfn = packageFamilyName;

                for (int attempt = 0; attempt < 3; ++attempt) {
                    std::this_thread::sleep_for(std::chrono::milliseconds(250 * (attempt + 1)));
                    if (!g_notifRunning) return;

                    try {
                        auto l = winrt_notif_mgmt::UserNotificationListener::Current();
                        auto notif = l.GetNotification(id);
                        if (notif && notif.AppInfo()) {
                            if (notif.AppInfo().DisplayInfo()) {
                                std::string winrtName = Utf16ToUtf8(notif.AppInfo().DisplayInfo().DisplayName().c_str());
                                if (!winrtName.empty()) {
                                    resolvedName = winrtName;
                                }
                                if (finalAppLogo.empty()) {
                                    auto logoRef = notif.AppInfo().DisplayInfo().GetLogo(winrt::Windows::Foundation::Size{48.0f, 48.0f});
                                    if (logoRef) {
                                        auto s = logoRef.OpenReadAsync().get();
                                        finalAppLogo = StreamToBase64(s);
                                    }
                                }
                            }
                            if (notif.AppInfo().PackageFamilyName() != L"") {
                                pfn = notif.AppInfo().PackageFamilyName().c_str();
                            }
                            break;
                        }
                    } catch (...) {}
                }

                if (finalAppLogo.empty()) {
                    finalAppLogo = ExtractAppIconAsDataUri(resolvedName.c_str(), aumid.c_str(), 48);
                }

                std::string finalIcon = icon;
                std::string finalHero = heroImage;

                if (finalIcon.empty()) {
                    DbToastInfo retryDb = ExtractToastFromDb(id, title, aumid);
                    if (!retryDb.appLogoOverride.empty()) {
                        ProcessedImage img = ProcessImageSource(Utf8ToUtf16(retryDb.appLogoOverride), pfn);
                        if (!img.dataUri.empty()) finalIcon = img.dataUri;
                    }
                }
                if (finalIcon.empty()) {
                    finalIcon = finalAppLogo;
                }

                if (g_notifCb && (!finalAppLogo.empty() || resolvedName != appName || finalIcon != icon)) {
                    std::ostringstream json;
                    json << "{\"type\":\"notification_update\","
                         << "\"id\":" << id << ","
                         << "\"appName\":\"" << EscapeJsonString(resolvedName) << "\","
                         << "\"aumid\":\"" << EscapeJsonString(aumid) << "\","
                         << "\"icon\":\"" << EscapeJsonString(finalIcon) << "\","
                         << "\"appLogo\":\"" << EscapeJsonString(finalAppLogo) << "\","
                         << "\"title\":\"" << EscapeJsonString(title) << "\","
                         << "\"message\":\"" << EscapeJsonString(message) << "\","
                         << "\"heroImage\":\"" << EscapeJsonString(finalHero) << "\"}";
                    g_notifCb(json.str().c_str());
                }
            }).detach();
        };

        uint32_t lastProcessedId = 0;

        auto pollWpnDatabase = [&](bool isInitial) {
            HMODULE hSqlite = LoadLibraryW(L"winsqlite3.dll");
            if (!hSqlite) return;

            auto p_open_v2 = (fn_sqlite3_open_v2)GetProcAddress(hSqlite, "sqlite3_open_v2");
            auto p_prepare = (fn_sqlite3_prepare_v2)GetProcAddress(hSqlite, "sqlite3_prepare_v2");
            auto p_bind_int64 = (fn_sqlite3_bind_int64)GetProcAddress(hSqlite, "sqlite3_bind_int64");
            auto p_step = (fn_sqlite3_step)GetProcAddress(hSqlite, "sqlite3_step");
            auto p_text = (fn_sqlite3_column_text)GetProcAddress(hSqlite, "sqlite3_column_text");
            auto p_blob = (fn_sqlite3_column_blob)GetProcAddress(hSqlite, "sqlite3_column_blob");
            auto p_bytes = (fn_sqlite3_column_bytes)GetProcAddress(hSqlite, "sqlite3_column_bytes");
            auto p_int64 = (fn_sqlite3_column_int64)GetProcAddress(hSqlite, "sqlite3_column_int64");
            auto p_finalize = (fn_sqlite3_finalize)GetProcAddress(hSqlite, "sqlite3_finalize");
            auto p_close = (fn_sqlite3_close)GetProcAddress(hSqlite, "sqlite3_close");

            if (!p_open_v2 || !p_prepare || !p_step || !p_finalize || !p_close) return;

            char localAppData[MAX_PATH];
            if (GetEnvironmentVariableA("LOCALAPPDATA", localAppData, MAX_PATH) == 0) return;
            std::string dbPath = std::string(localAppData) + "\\Microsoft\\Windows\\Notifications\\wpndatabase.db";
            std::string uri = "file:" + dbPath + "?mode=ro";

            sqlite3* db = nullptr;
            int rc = p_open_v2(uri.c_str(), &db, 0x00000001 | 0x00000040, nullptr);
            if (rc != 0 || !db) {
                rc = p_open_v2(dbPath.c_str(), &db, 0x00000001, nullptr);
            }
            if (rc != 0 || !db) return;

            if (isInitial) {
                const char* maxSql = "SELECT MAX(Id) FROM Notification WHERE (Type = 'toast' OR Type = 'toastCondensed');";
                sqlite3_stmt* maxStmt = nullptr;
                if (p_prepare(db, maxSql, -1, &maxStmt, nullptr) == 0) {
                    if (p_step(maxStmt) == 100 && p_int64) {
                        lastProcessedId = static_cast<uint32_t>(p_int64(maxStmt, 0));
                    }
                    p_finalize(maxStmt);
                }
                const char* seedSql =
                    "SELECT n.Id "
                    "FROM Notification n "
                    "WHERE (n.Type = 'toast' OR n.Type = 'toastCondensed') "
                    "ORDER BY n.Id DESC LIMIT 50;";
                sqlite3_stmt* seedStmt = nullptr;
                if (p_prepare(db, seedSql, -1, &seedStmt, nullptr) == 0) {
                    std::lock_guard<std::mutex> lock(seenMutex);
                    while (p_step(seedStmt) == 100) {
                        uint32_t seedId = static_cast<uint32_t>(p_int64 ? p_int64(seedStmt, 0) : 0);
                        seenRecords[seedId] = { "__INITIAL__" };
                    }
                    p_finalize(seedStmt);
                }
                p_close(db);
                return;
            }

            const char* sql =
                "SELECT n.Id, h.PrimaryId, n.ArrivalTime, n.Payload, ha.AssetValue "
                "FROM Notification n "
                "LEFT JOIN NotificationHandler h ON n.HandlerId = h.RecordId "
                "LEFT JOIN HandlerAssets ha ON n.HandlerId = ha.HandlerId AND ha.AssetKey = 'DisplayName' "
                "WHERE (n.Type = 'toast' OR n.Type = 'toastCondensed') AND n.Id > ? "
                "ORDER BY n.Id ASC;";

            sqlite3_stmt* stmt = nullptr;
            if (p_prepare(db, sql, -1, &stmt, nullptr) == 0) {
                if (p_bind_int64) {
                    p_bind_int64(stmt, 1, static_cast<int64_t>(lastProcessedId));
                }
                while (p_step(stmt) == 100) {
                    uint32_t id = static_cast<uint32_t>(p_int64 ? p_int64(stmt, 0) : 0);
                    if (id > lastProcessedId) {
                        lastProcessedId = id;
                    }
                    const char* aumidPtr = (const char*)p_text(stmt, 1);
                    std::string aumid = aumidPtr ? aumidPtr : "";
                    int64_t arrivalTime = p_int64 ? p_int64(stmt, 2) : 0;
                    const void* blob = p_blob(stmt, 3);
                    int blobBytes = p_bytes(stmt, 3);
                    std::string payload = (blob && blobBytes > 0) ? std::string((const char*)blob, blobBytes) : "";
                    const char* assetValPtr = (const char*)p_text(stmt, 4);
                    std::string handlerDisplayName = assetValPtr ? assetValPtr : "";

                    std::string title;
                    std::string message;
                    ParseToastXmlText(payload, title, message);
                    std::string resolvedAppName = ResolveAppDisplayName(aumid, handlerDisplayName);
                    dispatchNotification(id, aumid, resolvedAppName, title, message, arrivalTime, payload);
                }
                p_finalize(stmt);
            }
            p_close(db);
        };

        // Seed initial notifications to prevent re-firing history on startup
        pollWpnDatabase(true);

        char localAppData[MAX_PATH];
        GetEnvironmentVariableA("LOCALAPPDATA", localAppData, MAX_PATH);
        std::string notifDir = std::string(localAppData) + "\\Microsoft\\Windows\\Notifications";

        HANDLE hChange = FindFirstChangeNotificationA(
            notifDir.c_str(),
            FALSE,
            FILE_NOTIFY_CHANGE_LAST_WRITE | FILE_NOTIFY_CHANGE_SIZE | FILE_NOTIFY_CHANGE_FILE_NAME
        );

        while (g_notifRunning) {
            DWORD waitRes = (hChange != INVALID_HANDLE_VALUE)
                ? WaitForSingleObject(hChange, 80)
                : (std::this_thread::sleep_for(std::chrono::milliseconds(80)), WAIT_TIMEOUT);

            if (waitRes == WAIT_OBJECT_0 && hChange != INVALID_HANDLE_VALUE) {
                FindNextChangeNotification(hChange);
            }

            // Real-time instant notification check from SQLite (fires < 80ms)
            try {
                pollWpnDatabase(false);
            } catch (...) {}
        }

        if (hChange != INVALID_HANDLE_VALUE) {
            FindCloseChangeNotification(hChange);
        }
    } catch (...) {}
}

void StartNotificationService(NativeNotificationCallback cb) {
    if (g_notifRunning) return;
    g_notifCb = cb;
    g_notifRunning = true;
    g_notifThread = std::thread(NotificationWorker);
}

void StopNotificationService() {
    if (!g_notifRunning) return;
    g_notifRunning = false;
    g_notifWakePending = true;
    g_notifWakeCv.notify_all();
    if (g_notifThread.joinable()) {
        g_notifThread.join();
    }
}

static std::wstring GetVintaIconPath() {
    wchar_t appData[MAX_PATH] = {0};
    if (GetEnvironmentVariableW(L"APPDATA", appData, MAX_PATH) > 0) {
        std::wstring iconPath = std::wstring(appData) + L"\\FreeMatrix Studios\\Vinta\\appicon.png";
        DWORD attribs = GetFileAttributesW(iconPath.c_str());
        if (attribs != INVALID_FILE_ATTRIBUTES && !(attribs & FILE_ATTRIBUTE_DIRECTORY)) {
            return iconPath;
        }
    }
    return L"";
}

static std::wstring EscapeXml(const std::wstring& src) {
    std::wstring out;
    out.reserve(src.size() + 16);
    for (wchar_t c : src) {
        switch (c) {
            case L'&': out += L"&amp;"; break;
            case L'<': out += L"&lt;"; break;
            case L'>': out += L"&gt;"; break;
            case L'\"': out += L"&quot;"; break;
            case L'\'': out += L"&apos;"; break;
            default: out += c; break;
        }
    }
    return out;
}

static void EnsureAumidRegistered() {
    static bool registered = false;
    if (registered) return;
    registered = true;

    const wchar_t* aumid = L"FreeMatrixStudios.Vinta";
    SetCurrentProcessExplicitAppUserModelID(aumid);

    std::wstring iconPath = GetVintaIconPath();
    wchar_t exePath[MAX_PATH] = {0};
    GetModuleFileNameW(nullptr, exePath, MAX_PATH);

    const wchar_t* iconToUse = !iconPath.empty() ? iconPath.c_str() : exePath;
    const wchar_t* displayName = L"Vinta";

    // 1. Register in HKCU\Software\Classes\AppUserModelId\FreeMatrixStudios.Vinta
    HKEY hKeyClasses = nullptr;
    const wchar_t* subkeyClasses = L"Software\\Classes\\AppUserModelId\\FreeMatrixStudios.Vinta";
    if (RegCreateKeyExW(HKEY_CURRENT_USER, subkeyClasses, 0, nullptr, REG_OPTION_NON_VOLATILE, KEY_SET_VALUE, nullptr, &hKeyClasses, nullptr) == ERROR_SUCCESS) {
        RegSetValueExW(hKeyClasses, L"DisplayName", 0, REG_SZ, reinterpret_cast<const BYTE*>(displayName), static_cast<DWORD>((wcslen(displayName) + 1) * sizeof(wchar_t)));
        RegSetValueExW(hKeyClasses, L"IconUri", 0, REG_SZ, reinterpret_cast<const BYTE*>(iconToUse), static_cast<DWORD>((wcslen(iconToUse) + 1) * sizeof(wchar_t)));
        RegSetValueExW(hKeyClasses, L"IconBackgroundColor", 0, REG_SZ, reinterpret_cast<const BYTE*>(L"0"), static_cast<DWORD>((wcslen(L"0") + 1) * sizeof(wchar_t)));
        RegCloseKey(hKeyClasses);
    }

    // 2. Register in HKCU\Software\Microsoft\Windows\CurrentVersion\Notifications\Settings\FreeMatrixStudios.Vinta
    HKEY hKeySettings = nullptr;
    const wchar_t* subkeySettings = L"Software\\Microsoft\\Windows\\CurrentVersion\\Notifications\\Settings\\FreeMatrixStudios.Vinta";
    if (RegCreateKeyExW(HKEY_CURRENT_USER, subkeySettings, 0, nullptr, REG_OPTION_NON_VOLATILE, KEY_SET_VALUE, nullptr, &hKeySettings, nullptr) == ERROR_SUCCESS) {
        DWORD one = 1;
        RegSetValueExW(hKeySettings, L"DisplayName", 0, REG_SZ, reinterpret_cast<const BYTE*>(displayName), static_cast<DWORD>((wcslen(displayName) + 1) * sizeof(wchar_t)));
        RegSetValueExW(hKeySettings, L"IconUri", 0, REG_SZ, reinterpret_cast<const BYTE*>(iconToUse), static_cast<DWORD>((wcslen(iconToUse) + 1) * sizeof(wchar_t)));
        RegSetValueExW(hKeySettings, L"ShowInActionCenter", 0, REG_DWORD, reinterpret_cast<const BYTE*>(&one), sizeof(DWORD));
        RegSetValueExW(hKeySettings, L"Enabled", 0, REG_DWORD, reinterpret_cast<const BYTE*>(&one), sizeof(DWORD));
        RegCloseKey(hKeySettings);
    }

    // 3. Ensure Start Menu Programs shortcut exists with System.AppUserModel.ID
    PWSTR programsPath = nullptr;
    if (SUCCEEDED(SHGetKnownFolderPath(FOLDERID_Programs, 0, nullptr, &programsPath))) {
        std::wstring shortcutPath = std::wstring(programsPath) + L"\\Vinta.lnk";
        CoTaskMemFree(programsPath);

        if (exePath[0] != L'\0') {
            Microsoft::WRL::ComPtr<IShellLinkW> shellLink;
            if (SUCCEEDED(CoCreateInstance(CLSID_ShellLink, nullptr, CLSCTX_INPROC_SERVER, IID_PPV_ARGS(&shellLink)))) {
                shellLink->SetPath(exePath);
                if (!iconPath.empty()) {
                    shellLink->SetIconLocation(iconPath.c_str(), 0);
                } else {
                    shellLink->SetIconLocation(exePath, 0);
                }

                Microsoft::WRL::ComPtr<IPropertyStore> propertyStore;
                if (SUCCEEDED(shellLink.As(&propertyStore))) {
                    PROPVARIANT pv;
                    if (SUCCEEDED(InitPropVariantFromString(aumid, &pv))) {
                        propertyStore->SetValue(PKEY_AppUserModel_ID, pv);
                        propertyStore->Commit();
                        PropVariantClear(&pv);
                    }
                }

                Microsoft::WRL::ComPtr<IPersistFile> persistFile;
                if (SUCCEEDED(shellLink.As(&persistFile))) {
                    persistFile->Save(shortcutPath.c_str(), TRUE);
                }
            }
        }
    }
}

int SendToastNotification(const wchar_t* title, const wchar_t* message) {
    try {
        if (!title && !message) return 0;
        std::wstring wTitle = title ? title : L"Vinta";
        std::wstring wMessage = message ? message : L"";

        EnsureAumidRegistered();

        std::wstring iconPath = GetVintaIconPath();
        std::wstring iconUri = iconPath;
        for (wchar_t& ch : iconUri) {
            if (ch == L'\\') ch = L'/';
        }

        std::wstring xmlStr = L"<toast><visual><binding template=\"ToastGeneric\">";
        xmlStr += L"<text>" + EscapeXml(wTitle) + L"</text>";
        if (!wMessage.empty()) {
            xmlStr += L"<text>" + EscapeXml(wMessage) + L"</text>";
        }
        if (!iconUri.empty()) {
            xmlStr += L"<image placement=\"appLogoOverride\" hint-crop=\"circle\" src=\"file:///" + iconUri + L"\"/>";
        }
        xmlStr += L"</binding></visual></toast>";

        winrt::Windows::Data::Xml::Dom::XmlDocument toastXml;
        toastXml.LoadXml(xmlStr);
        winrt_notif::ToastNotification toast(toastXml);
        winrt_notif::ToastNotificationManager::CreateToastNotifier(L"FreeMatrixStudios.Vinta").Show(toast);
        return 1;
    } catch (...) {
        return 0;
    }
}
