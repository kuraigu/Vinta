/**
 * @file notification_service.cpp
 * @brief UserNotificationListener toast notification interception implementation.
 */

#include "notification_service.h"
#include "winrt_common.h"
#include <shlobj.h>
#include <propkey.h>
#include <propvarutil.h>
#include <wrl/client.h>

#include <winrt/Windows.UI.Notifications.Management.h>
#include <winrt/Windows.UI.Notifications.h>
#include <winrt/Windows.Data.Xml.Dom.h>

#include <atomic>
#include <regex>
#include <sstream>
#include <thread>
#include <unordered_set>

namespace winrt_notif = winrt::Windows::UI::Notifications;
namespace winrt_notif_mgmt = winrt::Windows::UI::Notifications::Management;

static NativeNotificationCallback g_notifCb = nullptr;
static std::atomic<bool> g_notifRunning{false};
static std::thread g_notifThread;

typedef struct sqlite3 sqlite3;
typedef struct sqlite3_stmt sqlite3_stmt;

typedef int (*fn_sqlite3_open_v2)(const char *filename, sqlite3 **ppDb, int flags, const char *zVfs);
typedef int (*fn_sqlite3_prepare_v2)(sqlite3 *db, const char *zSql, int nByte, sqlite3_stmt **ppStmt, const char **pzTail);
typedef int (*fn_sqlite3_bind_int64)(sqlite3_stmt *pStmt, int iCol, int64_t val);
typedef int (*fn_sqlite3_step)(sqlite3_stmt *pStmt);
typedef const unsigned char *(*fn_sqlite3_column_text)(sqlite3_stmt *pStmt, int iCol);
typedef const void *(*fn_sqlite3_column_blob)(sqlite3_stmt *pStmt, int iCol);
typedef int (*fn_sqlite3_column_bytes)(sqlite3_stmt *pStmt, int iCol);
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
 * Extracts toast payload information with retry polling.
 * When Windows broadcasts UserNotificationListener notifications,
 * the Action Center service is concurrently committing the toast XML record
 * to wpndatabase.db and its write-ahead log.
 * Retrying across short intervals captures the committed payload records.
 */
static DbToastInfo ExtractToastWithRetry(uint32_t notifId, const std::string& targetTitle, const std::string& aumid) {
    DbToastInfo dbInfo;
    const int delaysMs[] = { 0, 40, 80, 140 };
    for (int attempt = 0; attempt < 4; ++attempt) {
        if (delaysMs[attempt] > 0) {
            std::this_thread::sleep_for(std::chrono::milliseconds(delaysMs[attempt]));
        }
        dbInfo = ExtractToastFromDb(notifId, targetTitle, aumid);
        if (!dbInfo.appLogoOverride.empty() || !dbInfo.heroImage.empty() || !dbInfo.inlineImages.empty()) {
            return dbInfo;
        }
    }
    return dbInfo;
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

        while (g_notifRunning) {
            try {
                auto notifs = listener.GetNotificationsAsync(winrt_notif::NotificationKinds::Toast).get();

                // Prune records that are no longer active in Action Center
                std::unordered_set<uint32_t> activeIds;
                for (auto const& notif : notifs) {
                    activeIds.insert(notif.Id());
                }
                for (auto it = seenRecords.begin(); it != seenRecords.end(); ) {
                    if (activeIds.find(it->first) == activeIds.end()) {
                        it = seenRecords.erase(it);
                    } else {
                        ++it;
                    }
                }

                for (auto const& notif : notifs) {
                    uint32_t id = notif.Id();

                    std::string title;
                    std::string message;
                    try {
                        if (notif.Notification() && notif.Notification().Visual()) {
                            auto bindings = notif.Notification().Visual().Bindings();
                            for (auto const& binding : bindings) {
                                auto texts = binding.GetTextElements();
                                uint32_t textCount = texts.Size();
                                if (textCount > 0) {
                                    title = Utf16ToUtf8(texts.GetAt(0).Text().c_str());
                                    for (uint32_t t = 1; t < textCount; t++) {
                                        if (!message.empty()) message += "\n";
                                        message += Utf16ToUtf8(texts.GetAt(t).Text().c_str());
                                    }
                                }
                            }
                        }
                    } catch (...) {}

                    int64_t creationTicks = 0;
                    try {
                        creationTicks = notif.CreationTime().time_since_epoch().count();
                    } catch (...) {}

                    std::string currentSig = title + "\x1F" + message + "\x1F" + std::to_string(creationTicks);
                    auto it = seenRecords.find(id);
                    if (it != seenRecords.end()) {
                        if (it->second.signature == "__INITIAL__" || it->second.signature == currentSig) {
                            continue;
                        }
                    }

                    seenRecords[id] = { currentSig };

                    std::string appName;
                    std::string aumid;
                    try {
                        if (notif.AppInfo() && notif.AppInfo().DisplayInfo()) {
                            appName = Utf16ToUtf8(notif.AppInfo().DisplayInfo().DisplayName().c_str());
                        }
                    } catch (...) {}

                    try {
                        if (notif.AppInfo()) {
                            aumid = Utf16ToUtf8(notif.AppInfo().AppUserModelId().c_str());
                            if (appName.empty()) appName = aumid;
                        }
                    } catch (...) {}

                    std::string icon;
                    std::string heroImage;
                    std::wstring packageFamilyName;

                    try {
                        if (notif.AppInfo() && notif.AppInfo().PackageFamilyName() != L"") {
                            packageFamilyName = notif.AppInfo().PackageFamilyName().c_str();
                        }
                    } catch (...) {}

                    try {
                        DbToastInfo dbInfo = ExtractToastWithRetry(id, title, aumid);
                        if (!dbInfo.appLogoOverride.empty()) {
                            ProcessedImage appLogoImg = ProcessImageSource(Utf8ToUtf16(dbInfo.appLogoOverride), packageFamilyName);
                            if (!appLogoImg.dataUri.empty()) {
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
                    } catch (...) {}

                        std::string appLogo;
                        try {
                            if (notif.AppInfo() && notif.AppInfo().DisplayInfo()) {
                                auto logoStreamRef = notif.AppInfo().DisplayInfo().GetLogo(winrt::Windows::Foundation::Size{48.0f, 48.0f});
                                if (logoStreamRef) {
                                    auto stream = logoStreamRef.OpenReadAsync().get();
                                    appLogo = StreamToBase64(stream);
                                }
                            }
                        } catch (...) {}

                        if (icon.empty()) {
                            icon = appLogo;
                        }

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

                        /**
                         * Delayed SQLite transaction polling:
                         * Windows Action Center writes toast records to wpndatabase.db asynchronously.
                         * When an initial notification read catches only the generic app logo, polling
                         * across short intervals captures the committed avatar or hero image and
                         * emits an in-place update to the frontend.
                         */
                        if (icon == appLogo && heroImage.empty() && g_notifCb) {
                            std::thread([id, title, aumid, appName, message, packageFamilyName, appLogo]() {
                                ApartmentScope apt;
                                const int retryDelays[] = { 300, 700, 1300 };
                                for (int delay : retryDelays) {
                                    std::this_thread::sleep_for(std::chrono::milliseconds(delay));
                                    if (!g_notifRunning) return;

                                    DbToastInfo retryDb = ExtractToastFromDb(id, title, aumid);
                                    std::string lateIcon;
                                    std::string lateHero;

                                    if (!retryDb.appLogoOverride.empty()) {
                                        ProcessedImage img = ProcessImageSource(Utf8ToUtf16(retryDb.appLogoOverride), packageFamilyName);
                                        if (!img.dataUri.empty()) lateIcon = img.dataUri;
                                    }

                                    if (!retryDb.heroImage.empty()) {
                                        ProcessedImage heroImg = ProcessImageSource(Utf8ToUtf16(retryDb.heroImage), packageFamilyName);
                                        if (!heroImg.dataUri.empty()) {
                                            bool isAvatarHint = retryDb.heroImage.find("avatar") != std::string::npos ||
                                                                retryDb.heroImage.find("avatars") != std::string::npos ||
                                                                retryDb.heroImage.find("profile") != std::string::npos ||
                                                                retryDb.heroImage.find("user") != std::string::npos;
                                            if ((heroImg.isAvatarSize() || isAvatarHint) && lateIcon.empty()) {
                                                lateIcon = heroImg.dataUri;
                                            } else {
                                                lateHero = heroImg.dataUri;
                                            }
                                        }
                                    }

                                    for (const auto& inlineSrc : retryDb.inlineImages) {
                                        ProcessedImage inImg = ProcessImageSource(Utf8ToUtf16(inlineSrc), packageFamilyName);
                                        if (!inImg.dataUri.empty()) {
                                            bool isAvatarHint = inlineSrc.find("avatar") != std::string::npos ||
                                                                inlineSrc.find("avatars") != std::string::npos ||
                                                                inlineSrc.find("profile") != std::string::npos ||
                                                                inlineSrc.find("user") != std::string::npos;
                                            if ((inImg.isAvatarSize() || isAvatarHint) && lateIcon.empty()) {
                                                lateIcon = inImg.dataUri;
                                            } else if (lateHero.empty()) {
                                                lateHero = inImg.dataUri;
                                            }
                                        }
                                    }

                                    if (!lateIcon.empty() || !lateHero.empty()) {
                                        if (g_notifCb) {
                                            std::ostringstream json;
                                            json << "{\"type\":\"notification_update\","
                                                 << "\"id\":" << id << ","
                                                 << "\"appName\":\"" << EscapeJsonString(appName) << "\","
                                                 << "\"aumid\":\"" << EscapeJsonString(aumid) << "\","
                                                 << "\"icon\":\"" << EscapeJsonString(lateIcon.empty() ? appLogo : lateIcon) << "\","
                                                 << "\"appLogo\":\"" << EscapeJsonString(appLogo) << "\","
                                                 << "\"title\":\"" << EscapeJsonString(title) << "\","
                                                 << "\"message\":\"" << EscapeJsonString(message) << "\","
                                                 << "\"heroImage\":\"" << EscapeJsonString(lateHero) << "\"}";
                                            g_notifCb(json.str().c_str());
                                        }
                                        break;
                                    }
                                }
                            }).detach();
                        }
                }
            } catch (...) {}

            std::this_thread::sleep_for(std::chrono::milliseconds(120));
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
