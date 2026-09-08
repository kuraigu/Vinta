#include "notification_service.h"
#include "winrt_common.h"

#include <winrt/Windows.UI.Notifications.Management.h>
#include <winrt/Windows.UI.Notifications.h>

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

        if (placement == "appLogoOverride" || placement == "applogo" || hasCircleCrop) {
            // According to Windows Toast schema guidelines, circle crop or appLogoOverride
            // explicitly designates a contact profile avatar rather than a decorative banner.
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
    int rc = p_open(dbPath.c_str(), &db, 0x00000001, nullptr);
    if (rc != 0) {
        char tempPath[MAX_PATH];
        GetTempPathA(MAX_PATH, tempPath);
        std::string copyPath = std::string(tempPath) + "wpn_vinta_" + std::to_string(GetCurrentThreadId()) + ".db";
        std::string walPath = dbPath + "-wal";
        std::string copyWal = copyPath + "-wal";
        CopyFileA(dbPath.c_str(), copyPath.c_str(), FALSE);
        CopyFileA(walPath.c_str(), copyWal.c_str(), FALSE);
        if (p_open(copyPath.c_str(), &db, 0x00000001, nullptr) != 0) return outInfo;
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

    if (!found) {
        const char *sql = 
            "SELECT h.PrimaryId, n.Payload "
            "FROM Notification n "
            "LEFT JOIN NotificationHandler h ON n.HandlerId = h.RecordId "
            "WHERE n.Type = 'toast' "
            "ORDER BY n.ArrivalTime DESC LIMIT 25;";

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

                bool match = false;
                if (!targetTitle.empty()) {
                    if (dbTitle.find(targetTitle) != std::string::npos ||
                        targetTitle.find(dbTitle) != std::string::npos ||
                        payload.find(targetTitle) != std::string::npos) {
                        match = true;
                    }
                } else if (!aumid.empty() && !primStr.empty()) {
                    if (primStr.find(aumid) != std::string::npos || aumid.find(primStr) != std::string::npos) {
                        match = true;
                    }
                }

                if (match) {
                    ParseToastXmlPayload(payload, outInfo);
                    break;
                }
            }
            p_finalize(stmt);
        }
    }

    p_close(db);
    return outInfo;
}

/**
 * Extracts toast payload information with retry polling.
 * When Windows broadcasts UserNotificationListener notifications,
 * the Action Center service is concurrently committing the toast XML record
 * to wpndatabase.db and its write-ahead log (WAL).
 * A single immediate read often misses the payload on the first notification.
 * Retrying across short intervals allows SQLite to complete the transaction commit,
 * ensuring sender avatars and hero images are captured on the initial message.
 */
static DbToastInfo ExtractToastWithRetry(uint32_t notifId, const std::string& targetTitle, const std::string& aumid) {
    DbToastInfo dbInfo;
    const int delaysMs[] = { 40, 80, 120, 160 };
    for (int attempt = 0; attempt < 4; ++attempt) {
        dbInfo = ExtractToastFromDb(notifId, targetTitle, aumid);
        if (!dbInfo.appLogoOverride.empty() || !dbInfo.heroImage.empty() || !dbInfo.inlineImages.empty()) {
            return dbInfo;
        }
        if (attempt < 3) {
            std::this_thread::sleep_for(std::chrono::milliseconds(delaysMs[attempt]));
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

        std::unordered_set<uint32_t> seenIds;
        try {
            auto initialNotifs = listener.GetNotificationsAsync(winrt_notif::NotificationKinds::Toast).get();
            for (auto const& notif : initialNotifs) {
                seenIds.insert(notif.Id());
            }
        } catch (...) {}

        while (g_notifRunning) {
            try {
                auto notifs = listener.GetNotificationsAsync(winrt_notif::NotificationKinds::Toast).get();
                for (auto const& notif : notifs) {
                    uint32_t id = notif.Id();
                    if (seenIds.find(id) == seenIds.end()) {
                        seenIds.insert(id);

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

                        std::string title;
                        std::string message;
                        std::string icon;
                        std::string heroImage;
                        std::wstring packageFamilyName;

                        try {
                            if (notif.AppInfo() && notif.AppInfo().PackageFamilyName() != L"") {
                                packageFamilyName = notif.AppInfo().PackageFamilyName().c_str();
                            }
                        } catch (...) {}

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
                                    if (heroImg.isAvatarSize()) {
                                        if (icon.empty()) {
                                            icon = heroImg.dataUri;
                                        }
                                    } else {
                                        heroImage = heroImg.dataUri;
                                    }
                                }
                            }

                            for (const auto& inlineSrc : dbInfo.inlineImages) {
                                ProcessedImage inImg = ProcessImageSource(Utf8ToUtf16(inlineSrc), packageFamilyName);
                                if (!inImg.dataUri.empty()) {
                                    if (inImg.isAvatarSize() && icon.empty()) {
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
                                            if (heroImg.isAvatarSize()) {
                                                if (lateIcon.empty()) {
                                                    lateIcon = heroImg.dataUri;
                                                }
                                            } else {
                                                lateHero = heroImg.dataUri;
                                            }
                                        }
                                    }

                                    for (const auto& inlineSrc : retryDb.inlineImages) {
                                        ProcessedImage inImg = ProcessImageSource(Utf8ToUtf16(inlineSrc), packageFamilyName);
                                        if (!inImg.dataUri.empty()) {
                                            if (inImg.isAvatarSize() && lateIcon.empty()) {
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
                }

                if (seenIds.size() > notifs.Size() + 50) {
                    std::unordered_set<uint32_t> currentIds;
                    for (auto const& n : notifs) currentIds.insert(n.Id());
                    std::unordered_set<uint32_t> pruned;
                    for (uint32_t id : seenIds) {
                        if (currentIds.find(id) != currentIds.end()) pruned.insert(id);
                    }
                    seenIds = std::move(pruned);
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
