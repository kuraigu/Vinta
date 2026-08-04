#define NOMINMAX
#define WIN32_LEAN_AND_MEAN

#include "winrt_bridge.h"

#include <windows.h>
#include <mmdeviceapi.h>
#include <endpointvolume.h>
#include <audioclient.h>
#include <wincodec.h>

#include <winrt/base.h>
#include <winrt/Windows.Foundation.h>
#include <winrt/Windows.Foundation.Collections.h>
#include <winrt/Windows.ApplicationModel.h>
#include <winrt/Windows.Media.Control.h>
#include <winrt/Windows.Storage.Streams.h>
#include <winrt/Windows.UI.Notifications.Management.h>
#include <winrt/Windows.UI.Notifications.h>

#include <atomic>
#include <chrono>
#include <cmath>
#include <condition_variable>
#include <iomanip>
#include <iostream>
#include <memory>
#include <mutex>
#include <regex>
#include <sstream>
#include <string>
#include <thread>
#include <unordered_set>
#include <vector>

namespace winrt_media = winrt::Windows::Media::Control;
namespace winrt_storage = winrt::Windows::Storage::Streams;
namespace winrt_notif = winrt::Windows::UI::Notifications;
namespace winrt_notif_mgmt = winrt::Windows::UI::Notifications::Management;

struct ApartmentScope {
    ApartmentScope() {
        try {
            winrt::init_apartment(winrt::apartment_type::multi_threaded);
            m_initialized = true;
        } catch (...) {
            m_initialized = false;
        }
    }
    ~ApartmentScope() {
        if (m_initialized) {
            try { winrt::uninit_apartment(); } catch (...) {}
        }
    }
    bool m_initialized{false};
};

struct ComInitScope {
    HRESULT hr;
    ComInitScope() {
        hr = CoInitializeEx(NULL, COINIT_MULTITHREADED);
    }
    ~ComInitScope() {
        if (SUCCEEDED(hr)) {
            CoUninitialize();
        }
    }
};

static NativeMediaCallback g_mediaCb = nullptr;
static NativeNotificationCallback g_notifCb = nullptr;
static NativeAudioPeakCallback g_peakCb = nullptr;

static std::atomic<bool> g_running{false};
static std::thread g_audioThread;
static std::thread g_notifThread;
static std::thread g_mediaSessionPollThread;

static winrt_media::GlobalSystemMediaTransportControlsSessionManager g_sessionManager{nullptr};
static winrt_media::GlobalSystemMediaTransportControlsSession g_currentSession{nullptr};
static std::recursive_mutex g_sessionMutex;

static winrt::event_token g_currentSessionChangedToken;
static winrt::event_token g_sessionsChangedToken;
static winrt::event_token g_mediaPropsChangedToken;
static winrt::event_token g_playbackInfoChangedToken;
static winrt::event_token g_timelinePropsChangedToken;

static std::atomic<bool> g_isMediaPlaying{false};
static std::atomic<bool> g_emitPending{false};

static std::string g_cachedTrackKey;
static std::string g_cachedAlbumArt;
static std::mutex g_artCacheMutex;

// Forward declarations
static void UpdateCurrentSession();
static void EmitMediaState(bool includeVolume = true);
static void ScheduleEmit();
static std::string StreamToBase64(winrt_storage::IRandomAccessStreamWithContentType const& stream);
static std::pair<float, bool> GetMasterAudioVolumeAndMute();
static void SetMasterAudioVolume(float percent);
static void ToggleMasterAudioMute();
static void SendVolumeKey(BYTE vk);

// Base64 encoding helper
static const std::string base64_chars =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    "abcdefghijklmnopqrstuvwxyz"
    "0123456789+/";

static std::string Base64Encode(const unsigned char* bytes_to_encode, size_t in_len) {
    std::string ret;
    int i = 0;
    int j = 0;
    unsigned char char_array_3[3];
    unsigned char char_array_4[4];

    while (in_len--) {
        char_array_3[i++] = *(bytes_to_encode++);
        if (i == 3) {
            char_array_4[0] = (char_array_3[0] & 0xfc) >> 2;
            char_array_4[1] = ((char_array_3[0] & 0x03) << 4) + ((char_array_3[1] & 0xf0) >> 4);
            char_array_4[2] = ((char_array_3[1] & 0x0f) << 2) + ((char_array_3[2] & 0xc0) >> 6);
            char_array_4[3] = char_array_3[2] & 0x3f;

            for (i = 0; (i < 4); i++)
                ret += base64_chars[char_array_4[i]];
            i = 0;
        }
    }

    if (i) {
        for (j = i; j < 3; j++)
            char_array_3[j] = '\0';

        char_array_4[0] = (char_array_3[0] & 0xfc) >> 2;
        char_array_4[1] = ((char_array_3[0] & 0x03) << 4) + ((char_array_3[1] & 0xf0) >> 4);
        char_array_4[2] = ((char_array_3[1] & 0x0f) << 2) + ((char_array_3[2] & 0xc0) >> 6);

        for (j = 0; (j < i + 1); j++)
            ret += base64_chars[char_array_4[j]];

        while ((i++ < 3))
            ret += '=';
    }

    return ret;
}

// JSON string escaper
static std::string EscapeJsonString(const std::string& input) {
    std::ostringstream ss;
    for (char c : input) {
        switch (c) {
            case '"': ss << "\\\""; break;
            case '\\': ss << "\\\\"; break;
            case '\b': ss << "\\b"; break;
            case '\f': ss << "\\f"; break;
            case '\n': ss << "\\n"; break;
            case '\r': ss << "\\r"; break;
            case '\t': ss << "\\t"; break;
            default:
                if (static_cast<unsigned char>(c) < 0x20) {
                    ss << "\\u" << std::hex << std::setw(4) << std::setfill('0') << static_cast<int>(c);
                } else {
                    ss << c;
                }
                break;
        }
    }
    return ss.str();
}

static std::string Utf16ToUtf8(const std::wstring& wstr) {
    if (wstr.empty()) return std::string();
    int size_needed = WideCharToMultiByte(CP_UTF8, 0, &wstr[0], (int)wstr.size(), NULL, 0, NULL, NULL);
    std::string strTo(size_needed, 0);
    WideCharToMultiByte(CP_UTF8, 0, &wstr[0], (int)wstr.size(), &strTo[0], size_needed, NULL, NULL);
    return strTo;
}

struct ProcessedImage {
    std::string dataUri;
    int width{0};
    int height{0};

    bool isAvatarSize() const {
        if (dataUri.empty()) return false;
        if (width <= 0 || height <= 0) return true;
        float ratio = static_cast<float>(width) / static_cast<float>(height);
        return (ratio >= 0.75f && ratio <= 1.33f && (width <= 512 || height <= 512));
    }
};

static bool GetImageDimensions(const unsigned char* data, size_t len, int& outWidth, int& outHeight) {
    if (!data || len < 10) return false;

    // 1. PNG
    if (len >= 24 && data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G') {
        outWidth = (data[16] << 24) | (data[17] << 16) | (data[18] << 8) | data[19];
        outHeight = (data[20] << 24) | (data[21] << 16) | (data[22] << 8) | data[23];
        return (outWidth > 0 && outHeight > 0);
    }

    // 2. GIF
    if (len >= 10 && data[0] == 'G' && data[1] == 'I' && data[2] == 'F') {
        outWidth = data[6] | (data[7] << 8);
        outHeight = data[8] | (data[9] << 8);
        return (outWidth > 0 && outHeight > 0);
    }

    // 3. BMP
    if (len >= 26 && data[0] == 'B' && data[1] == 'M') {
        outWidth = data[18] | (data[19] << 8) | (data[20] << 16) | (data[21] << 24);
        int32_t h = data[22] | (data[23] << 8) | (data[24] << 16) | (data[25] << 24);
        outHeight = std::abs(h);
        return (outWidth > 0 && outHeight > 0);
    }

    // 4. JPEG
    if (len >= 4 && data[0] == 0xFF && data[1] == 0xD8) {
        size_t pos = 2;
        while (pos + 8 < len) {
            if (data[pos] != 0xFF) {
                pos++;
                continue;
            }
            unsigned char marker = data[pos + 1];
            pos += 2;
            if (marker == 0xD9 || marker == 0xDA) break; // EOI or SOS
            if (pos + 2 > len) break;
            size_t segLen = (data[pos] << 8) | data[pos + 1];
            if ((marker >= 0xC0 && marker <= 0xC3) || (marker >= 0xC5 && marker <= 0xC7) || (marker >= 0xC9 && marker <= 0xCB) || (marker >= 0xCD && marker <= 0xCF)) {
                if (pos + 7 < len) {
                    outHeight = (data[pos + 3] << 8) | data[pos + 4];
                    outWidth = (data[pos + 5] << 8) | data[pos + 6];
                    return (outWidth > 0 && outHeight > 0);
                }
            }
            pos += segLen;
        }
    }

    // 5. WebP
    if (len >= 30 && data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
        data[8] == 'W' && data[9] == 'E' && data[10] == 'B' && data[11] == 'P') {
        if (data[12] == 'V' && data[13] == 'P' && data[14] == '8' && data[15] == ' ') {
            if (len >= 30) {
                outWidth = (data[26] | (data[27] << 8)) & 0x3FFF;
                outHeight = (data[28] | (data[29] << 8)) & 0x3FFF;
                return (outWidth > 0 && outHeight > 0);
            }
        } else if (data[12] == 'V' && data[13] == 'P' && data[14] == '8' && data[15] == 'L') {
            if (len >= 25) {
                uint32_t b0 = data[21], b1 = data[22], b2 = data[23], b3 = data[24];
                outWidth = 1 + (((b1 & 0x3F) << 8) | b0);
                outHeight = 1 + (((b3 & 0x0F) << 10) | (b2 << 2) | ((b1 & 0xC0) >> 6));
                return (outWidth > 0 && outHeight > 0);
            }
        } else if (data[12] == 'V' && data[13] == 'P' && data[14] == '8' && data[15] == 'X') {
            if (len >= 30) {
                outWidth = 1 + (data[24] | (data[25] << 8) | (data[26] << 16));
                outHeight = 1 + (data[27] | (data[28] << 8) | (data[29] << 16));
                return (outWidth > 0 && outHeight > 0);
            }
        }
    }

    return false;
}

static ProcessedImage ReadFileToProcessedImage(const std::wstring& filePath) {
    ProcessedImage result;
    if (filePath.empty()) return result;

    DWORD attrs = GetFileAttributesW(filePath.c_str());
    if (attrs == INVALID_FILE_ATTRIBUTES || (attrs & FILE_ATTRIBUTE_DIRECTORY)) {
        return result;
    }

    HANDLE hFile = CreateFileW(filePath.c_str(), GENERIC_READ, FILE_SHARE_READ | FILE_SHARE_WRITE | FILE_SHARE_DELETE, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
    if (hFile == INVALID_HANDLE_VALUE) return result;

    LARGE_INTEGER size;
    if (!GetFileSizeEx(hFile, &size) || size.QuadPart == 0 || size.QuadPart > 20 * 1024 * 1024) {
        CloseHandle(hFile);
        return result;
    }

    std::vector<unsigned char> buffer(static_cast<size_t>(size.QuadPart));
    DWORD bytesRead = 0;
    if (!ReadFile(hFile, buffer.data(), static_cast<DWORD>(buffer.size()), &bytesRead, NULL) || bytesRead == 0) {
        CloseHandle(hFile);
        return result;
    }
    CloseHandle(hFile);

    GetImageDimensions(buffer.data(), bytesRead, result.width, result.height);

    std::string mime = "image/png";
    size_t dotPos = filePath.rfind(L'.');
    if (dotPos != std::wstring::npos) {
        std::wstring ext = filePath.substr(dotPos);
        for (auto& c : ext) c = towlower(c);
        if (ext == L".jpg" || ext == L".jpeg") mime = "image/jpeg";
        else if (ext == L".webp") mime = "image/webp";
        else if (ext == L".gif") mime = "image/gif";
        else if (ext == L".bmp") mime = "image/bmp";
        else if (ext == L".ico") mime = "image/x-icon";
        else if (ext == L".svg") mime = "image/svg+xml";
    }

    result.dataUri = "data:" + mime + ";base64," + Base64Encode(buffer.data(), bytesRead);
    return result;
}

static ProcessedImage ProcessImageSource(const std::wstring& srcUri, const std::wstring& packageFamilyName = L"") {
    ProcessedImage result;
    if (srcUri.empty()) return result;

    if (srcUri.rfind(L"http://", 0) == 0 || srcUri.rfind(L"https://", 0) == 0 || srcUri.rfind(L"data:", 0) == 0) {
        result.dataUri = Utf16ToUtf8(srcUri);
        return result;
    }

    std::wstring path = srcUri;
    if (path.rfind(L"file:///", 0) == 0) {
        path = path.substr(8);
    } else if (path.rfind(L"file://", 0) == 0) {
        path = path.substr(7);
    }

    for (auto& c : path) {
        if (c == L'/') c = L'\\';
    }

    std::wstring unescapedPath;
    for (size_t i = 0; i < path.length(); ++i) {
        if (path[i] == L'%' && i + 2 < path.length()) {
            wchar_t hex[3] = { path[i + 1], path[i + 2], 0 };
            wchar_t* end = nullptr;
            long val = wcstol(hex, &end, 16);
            if (end == hex + 2) {
                unescapedPath += static_cast<wchar_t>(val);
                i += 2;
                continue;
            }
        }
        unescapedPath += path[i];
    }

    result = ReadFileToProcessedImage(unescapedPath);
    if (!result.dataUri.empty()) return result;

    result = ReadFileToProcessedImage(path);
    if (!result.dataUri.empty()) return result;

    if (srcUri.rfind(L"ms-appdata://", 0) == 0 && !packageFamilyName.empty()) {
        wchar_t localAppData[MAX_PATH];
        if (GetEnvironmentVariableW(L"LOCALAPPDATA", localAppData, MAX_PATH) > 0) {
            std::wstring pkgBase = std::wstring(localAppData) + L"\\Packages\\" + packageFamilyName;
            std::wstring relativePath = srcUri.substr(13);
            if (relativePath.rfind(L"/local/", 0) == 0 || relativePath.rfind(L"local/", 0) == 0) {
                size_t offset = (relativePath[0] == L'/') ? 7 : 6;
                std::wstring fullPath = pkgBase + L"\\LocalState\\" + relativePath.substr(offset);
                for (auto& c : fullPath) { if (c == L'/') c = L'\\'; }
                result = ReadFileToProcessedImage(fullPath);
                if (!result.dataUri.empty()) return result;
            } else if (relativePath.rfind(L"/temp/", 0) == 0 || relativePath.rfind(L"temp/", 0) == 0) {
                size_t offset = (relativePath[0] == L'/') ? 6 : 5;
                std::wstring fullPath = pkgBase + L"\\TempState\\" + relativePath.substr(offset);
                for (auto& c : fullPath) { if (c == L'/') c = L'\\'; }
                result = ReadFileToProcessedImage(fullPath);
                if (!result.dataUri.empty()) return result;
            }
        }
    }

    wchar_t tempPath[MAX_PATH];
    if (GetTempPathW(MAX_PATH, tempPath) > 0) {
        std::wstring candidate = std::wstring(tempPath) + unescapedPath;
        result = ReadFileToProcessedImage(candidate);
        if (!result.dataUri.empty()) return result;
    }

    return result;
}

// Audio Peak Meter Background Worker
static void AudioPeakWorker() {
    ComInitScope com;
    IMMDeviceEnumerator* pEnumerator = nullptr;
    HRESULT hr = CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, __uuidof(IMMDeviceEnumerator), (void**)&pEnumerator);
    if (FAILED(hr) || !pEnumerator) {
        return;
    }

    std::vector<IAudioMeterInformation*> meters;
    auto lastDeviceCheck = std::chrono::steady_clock::now();

    while (g_running) {
        std::this_thread::sleep_for(std::chrono::milliseconds(40)); // 25 FPS
        if (!g_isMediaPlaying) {
            continue;
        }

        auto now = std::chrono::steady_clock::now();
        if (now - lastDeviceCheck > std::chrono::seconds(5) || meters.empty()) {
            for (auto* m : meters) {
                if (m) m->Release();
            }
            meters.clear();

            IMMDeviceCollection* pCollection = nullptr;
            if (SUCCEEDED(pEnumerator->EnumAudioEndpoints(eRender, DEVICE_STATE_ACTIVE, &pCollection)) && pCollection) {
                UINT count = 0;
                pCollection->GetCount(&count);
                for (UINT i = 0; i < count; i++) {
                    IMMDevice* pDevice = nullptr;
                    if (SUCCEEDED(pCollection->Item(i, &pDevice)) && pDevice) {
                        IAudioMeterInformation* pMeter = nullptr;
                        if (SUCCEEDED(pDevice->Activate(__uuidof(IAudioMeterInformation), CLSCTX_ALL, NULL, (void**)&pMeter)) && pMeter) {
                            meters.push_back(pMeter);
                        }
                        pDevice->Release();
                    }
                }
                pCollection->Release();
            }
            lastDeviceCheck = now;
        }

        float maxPeak = 0.0f;
        for (auto* m : meters) {
            float peak = 0.0f;
            if (SUCCEEDED(m->GetPeakValue(&peak))) {
                if (peak > maxPeak) maxPeak = peak;
            }
        }

        if (g_peakCb) {
            uint32_t bits = 0;
            std::memcpy(&bits, &maxPeak, sizeof(bits));
            g_peakCb(bits);
        }
    }

    for (auto* m : meters) {
        if (m) m->Release();
    }
    if (pEnumerator) pEnumerator->Release();
}

typedef struct sqlite3 sqlite3;
typedef struct sqlite3_stmt sqlite3_stmt;

typedef int (*fn_sqlite3_open_v2)(const char *filename, sqlite3 **ppDb, int flags, const char *zVfs);
typedef int (*fn_sqlite3_prepare_v2)(sqlite3 *db, const char *zSql, int nByte, sqlite3_stmt **ppStmt, const char **pzTail);
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

static std::wstring Utf8ToUtf16(const std::string& str) {
    if (str.empty()) return std::wstring();
    int size_needed = MultiByteToWideChar(CP_UTF8, 0, &str[0], (int)str.size(), NULL, 0);
    std::wstring wstrTo(size_needed, 0);
    MultiByteToWideChar(CP_UTF8, 0, &str[0], (int)str.size(), &wstrTo[0], size_needed);
    return wstrTo;
}

static DbToastInfo ExtractToastFromDb(const std::string& targetTitle, const std::string& aumid) {
    DbToastInfo outInfo;
    HMODULE hSqlite = LoadLibraryW(L"winsqlite3.dll");
    if (!hSqlite) return outInfo;

    auto p_open = (fn_sqlite3_open_v2)GetProcAddress(hSqlite, "sqlite3_open_v2");
    auto p_prepare = (fn_sqlite3_prepare_v2)GetProcAddress(hSqlite, "sqlite3_prepare_v2");
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

    char tempPath[MAX_PATH];
    GetTempPathA(MAX_PATH, tempPath);
    std::string copyPath = std::string(tempPath) + "wpn_vinta_sync.db";
    CopyFileA(dbPath.c_str(), copyPath.c_str(), FALSE);

    sqlite3 *db = nullptr;
    if (p_open(copyPath.c_str(), &db, 0x00000001, nullptr) != 0) return outInfo;

    const char *sql = 
        "SELECT h.PrimaryId, n.Payload "
        "FROM Notification n "
        "LEFT JOIN NotificationHandler h ON n.HandlerId = h.RecordId "
        "WHERE n.Type = 'toast' "
        "ORDER BY n.ArrivalTime DESC LIMIT 10;";

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
            if (!targetTitle.empty() && (dbTitle.find(targetTitle) != std::string::npos || targetTitle.find(dbTitle) != std::string::npos)) {
                match = true;
            } else if (!aumid.empty() && primStr.find(aumid) != std::string::npos) {
                match = true;
            } else if (targetTitle.empty()) {
                match = true;
            }

            if (match) {
                std::regex imgRegex("<image[^>]+>", std::regex::icase);
                std::regex srcRegex("src=\"([^\"]+)\"", std::regex::icase);
                std::regex placementRegex("placement=\"([^\"]+)\"", std::regex::icase);

                auto imgBegin = std::sregex_iterator(payload.begin(), payload.end(), imgRegex);
                auto imgEnd = std::sregex_iterator();
                for (auto it = imgBegin; it != imgEnd; ++it) {
                    std::string imgTag = (*it).str();
                    std::smatch smSrc, smPlacement;
                    std::string src, placement;
                    if (std::regex_search(imgTag, smSrc, srcRegex)) src = smSrc[1].str();
                    if (std::regex_search(imgTag, smPlacement, placementRegex)) placement = smPlacement[1].str();

                    if (src.empty()) continue;

                    if (placement == "appLogoOverride" || placement == "applogo") {
                        outInfo.appLogoOverride = src;
                    } else if (placement == "hero") {
                        outInfo.heroImage = src;
                    } else {
                        outInfo.inlineImages.push_back(src);
                    }
                }
                break;
            }
        }
        p_finalize(stmt);
    }

    p_close(db);
    return outInfo;
}

// Notification Listener Background Worker
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

        while (g_running) {
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

                        // 1. Try extracting default App Logo
                        try {
                            if (notif.AppInfo() && notif.AppInfo().DisplayInfo()) {
                                auto logoStreamRef = notif.AppInfo().DisplayInfo().GetLogo(winrt::Windows::Foundation::Size{48.0f, 48.0f});
                                if (logoStreamRef) {
                                    auto stream = logoStreamRef.OpenReadAsync().get();
                                    icon = StreamToBase64(stream);
                                }
                            }
                        } catch (...) {}

                        std::wstring packageFamilyName;
                        try {
                            if (notif.AppInfo() && notif.AppInfo().PackageFamilyName() != L"") {
                                packageFamilyName = notif.AppInfo().PackageFamilyName().c_str();
                            }
                        } catch (...) {}

                        // 2. Parse Visual Bindings for Text
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

                        // 3. Extract exact XML payload and image tags from Windows wpndatabase.db
                        try {
                            DbToastInfo dbInfo = ExtractToastFromDb(title, aumid);
                            if (!dbInfo.appLogoOverride.empty()) {
                                ProcessedImage appLogoImg = ProcessImageSource(Utf8ToUtf16(dbInfo.appLogoOverride), packageFamilyName);
                                if (!appLogoImg.dataUri.empty()) {
                                    icon = appLogoImg.dataUri;
                                }
                            }

                            if (!dbInfo.heroImage.empty()) {
                                ProcessedImage heroImg = ProcessImageSource(Utf8ToUtf16(dbInfo.heroImage), packageFamilyName);
                                if (!heroImg.dataUri.empty()) {
                                    if (heroImg.isAvatarSize() && icon.empty()) {
                                        icon = heroImg.dataUri;
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

                        if (g_notifCb) {
                            std::ostringstream json;
                            json << "{\"type\":\"notification\","
                                 << "\"appName\":\"" << EscapeJsonString(appName) << "\","
                                 << "\"icon\":\"" << EscapeJsonString(icon) << "\","
                                 << "\"title\":\"" << EscapeJsonString(title) << "\","
                                 << "\"message\":\"" << EscapeJsonString(message) << "\","
                                 << "\"heroImage\":\"" << EscapeJsonString(heroImage) << "\"}";
                            g_notifCb(json.str().c_str());
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

static std::string StreamToBase64(winrt_storage::IRandomAccessStreamWithContentType const& stream) {
    if (!stream) return "";
    try {
        uint64_t size = stream.Size();
        if (size == 0 || size > 5 * 1024 * 1024) return "";

        winrt_storage::Buffer buffer(static_cast<uint32_t>(size));
        auto readOp = stream.ReadAsync(buffer, static_cast<uint32_t>(size), winrt_storage::InputStreamOptions::None).get();

        auto dataReader = winrt_storage::DataReader::FromBuffer(readOp);
        std::vector<uint8_t> bytes(static_cast<size_t>(size));
        dataReader.ReadBytes(bytes);

        return "data:image/png;base64," + Base64Encode(bytes.data(), bytes.size());
    } catch (...) {
        return "";
    }
}

static void ScheduleEmit() {
    bool expected = false;
    if (g_emitPending.compare_exchange_strong(expected, true)) {
        std::thread([]() {
            ApartmentScope apt;
            std::this_thread::sleep_for(std::chrono::milliseconds(100));
            g_emitPending.store(false);
            EmitMediaState(true);
        }).detach();
    }
}

static void UpdateCurrentSession() {
    std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
    if (!g_sessionManager) return;

    try {
        winrt_media::GlobalSystemMediaTransportControlsSession targetSession{nullptr};
        try {
            targetSession = g_sessionManager.GetCurrentSession();
        } catch (...) {}

        if (!targetSession) {
            try {
                auto sessions = g_sessionManager.GetSessions();
                for (auto const& s : sessions) {
                    try {
                        auto info = s.GetPlaybackInfo();
                        if (info && info.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing) {
                            targetSession = s;
                            break;
                        }
                    } catch (...) {}
                }
                if (!targetSession && sessions.Size() > 0) {
                    targetSession = sessions.GetAt(0);
                }
            } catch (...) {}
        }

        if (g_currentSession != targetSession) {
            if (g_currentSession) {
                try { g_currentSession.MediaPropertiesChanged(g_mediaPropsChangedToken); } catch (...) {}
                try { g_currentSession.PlaybackInfoChanged(g_playbackInfoChangedToken); } catch (...) {}
                try { g_currentSession.TimelinePropertiesChanged(g_timelinePropsChangedToken); } catch (...) {}
            }

            g_currentSession = targetSession;

            if (g_currentSession) {
                try {
                    g_mediaPropsChangedToken = g_currentSession.MediaPropertiesChanged([](auto const&, auto const&) { ScheduleEmit(); });
                    g_playbackInfoChangedToken = g_currentSession.PlaybackInfoChanged([](auto const&, auto const&) { ScheduleEmit(); });
                    g_timelinePropsChangedToken = g_currentSession.TimelinePropertiesChanged([](auto const&, auto const&) { ScheduleEmit(); });
                } catch (...) {}
            }
        }
    } catch (...) {}

    EmitMediaState(true);
}

static void EmitMediaState(bool includeVolume) {
    if (!g_mediaCb) return;

    std::string sourceApp;
    std::string title;
    std::string artist;
    std::string album;
    std::string albumArt;
    bool isPlaying = false;
    double posMs = 0.0;
    double durMs = 0.0;
    bool canGoNext = true;
    bool canGoPrev = true;
    bool canPause = true;
    bool canPlay = true;

    winrt_media::GlobalSystemMediaTransportControlsSession session{nullptr};
    {
        std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
        session = g_currentSession;
    }

    if (!session) {
        std::ostringstream json;
        json << "{\"type\":\"media\",\"isAvailable\":false}";
        g_mediaCb(json.str().c_str());
        return;
    }

    try {
        sourceApp = Utf16ToUtf8(session.SourceAppUserModelId().c_str());
    } catch (...) {}

    try {
        auto props = session.TryGetMediaPropertiesAsync().get();
        if (props) {
            title = Utf16ToUtf8(props.Title().c_str());
            artist = Utf16ToUtf8(props.Artist().c_str());
            album = Utf16ToUtf8(props.AlbumTitle().c_str());

            std::string trackKey = title + "_" + artist;
            {
                std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                if (!trackKey.empty() && trackKey == g_cachedTrackKey && !g_cachedAlbumArt.empty()) {
                    albumArt = g_cachedAlbumArt;
                }
            }

            if (albumArt.empty() && props.Thumbnail()) {
                try {
                    auto stream = props.Thumbnail().OpenReadAsync().get();
                    albumArt = StreamToBase64(stream);
                    if (!albumArt.empty() && !trackKey.empty()) {
                        std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                        g_cachedTrackKey = trackKey;
                        g_cachedAlbumArt = albumArt;
                    }
                } catch (...) {}
            }
        }
    } catch (...) {}

    try {
        auto info = session.GetPlaybackInfo();
        if (info) {
            isPlaying = (info.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing);
            g_isMediaPlaying.store(isPlaying);

            auto controls = info.Controls();
            if (controls) {
                canGoNext = controls.IsNextEnabled();
                canGoPrev = controls.IsPreviousEnabled();
                canPause = controls.IsPauseEnabled();
                canPlay = controls.IsPlayEnabled();
            }
        }
    } catch (...) {}

    try {
        auto timeline = session.GetTimelineProperties();
        if (timeline) {
            posMs = std::chrono::duration<double, std::milli>(timeline.Position()).count();
            durMs = std::chrono::duration<double, std::milli>(timeline.EndTime()).count();
        }
    } catch (...) {}

    auto [volume, isMuted] = GetMasterAudioVolumeAndMute();

    std::ostringstream json;
    json << "{\"type\":\"media\","
         << "\"isAvailable\":true,"
         << "\"sourceApp\":\"" << EscapeJsonString(sourceApp) << "\","
         << "\"title\":\"" << EscapeJsonString(title) << "\","
         << "\"artist\":\"" << EscapeJsonString(artist) << "\","
         << "\"album\":\"" << EscapeJsonString(album) << "\","
         << "\"albumArt\":\"" << EscapeJsonString(albumArt) << "\","
         << "\"isPlaying\":" << (isPlaying ? "true" : "false") << ","
         << "\"audioPeak\":0.0,"
         << "\"positionMs\":" << std::fixed << std::setprecision(0) << posMs << ","
         << "\"durationMs\":" << std::fixed << std::setprecision(0) << durMs << ","
         << "\"volume\":" << static_cast<int>(std::round(volume)) << ","
         << "\"isMuted\":" << (isMuted ? "true" : "false") << ","
         << "\"canGoNext\":" << (canGoNext ? "true" : "false") << ","
         << "\"canGoPrev\":" << (canGoPrev ? "true" : "false") << ","
         << "\"canPause\":" << (canPause ? "true" : "false") << ","
         << "\"canPlay\":" << (canPlay ? "true" : "false") << "}";

    g_mediaCb(json.str().c_str());
}

static std::pair<float, bool> GetMasterAudioVolumeAndMute() {
    float vol = 50.0f;
    bool muted = false;

    ComInitScope com;
    IMMDeviceEnumerator* pEnumerator = nullptr;
    if (SUCCEEDED(CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, __uuidof(IMMDeviceEnumerator), (void**)&pEnumerator)) && pEnumerator) {
        IMMDevice* pDevice = nullptr;
        if (SUCCEEDED(pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDevice)) && pDevice) {
            IAudioEndpointVolume* pEndpointVolume = nullptr;
            if (SUCCEEDED(pDevice->Activate(__uuidof(IAudioEndpointVolume), CLSCTX_ALL, NULL, (void**)&pEndpointVolume)) && pEndpointVolume) {
                float scalar = 0.5f;
                BOOL isMuted = FALSE;
                if (SUCCEEDED(pEndpointVolume->GetMasterVolumeLevelScalar(&scalar))) {
                    vol = scalar * 100.0f;
                }
                if (SUCCEEDED(pEndpointVolume->GetMute(&isMuted))) {
                    muted = (isMuted == TRUE);
                }
                pEndpointVolume->Release();
            }
            pDevice->Release();
        }
        pEnumerator->Release();
    }

    return {vol, muted};
}

static void SetMasterAudioVolume(float percent) {
    float clamped = std::max(0.0f, std::min(100.0f, percent));
    ComInitScope com;
    IMMDeviceEnumerator* pEnumerator = nullptr;
    if (SUCCEEDED(CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, __uuidof(IMMDeviceEnumerator), (void**)&pEnumerator)) && pEnumerator) {
        IMMDevice* pDevice = nullptr;
        if (SUCCEEDED(pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDevice)) && pDevice) {
            IAudioEndpointVolume* pEndpointVolume = nullptr;
            if (SUCCEEDED(pDevice->Activate(__uuidof(IAudioEndpointVolume), CLSCTX_ALL, NULL, (void**)&pEndpointVolume)) && pEndpointVolume) {
                pEndpointVolume->SetMasterVolumeLevelScalar(clamped / 100.0f, NULL);
                pEndpointVolume->Release();
            }
            pDevice->Release();
        }
        pEnumerator->Release();
    }
}

static void ToggleMasterAudioMute() {
    ComInitScope com;
    IMMDeviceEnumerator* pEnumerator = nullptr;
    if (SUCCEEDED(CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, __uuidof(IMMDeviceEnumerator), (void**)&pEnumerator)) && pEnumerator) {
        IMMDevice* pDevice = nullptr;
        if (SUCCEEDED(pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDevice)) && pDevice) {
            IAudioEndpointVolume* pEndpointVolume = nullptr;
            if (SUCCEEDED(pDevice->Activate(__uuidof(IAudioEndpointVolume), CLSCTX_ALL, NULL, (void**)&pEndpointVolume)) && pEndpointVolume) {
                BOOL isMuted = FALSE;
                if (SUCCEEDED(pEndpointVolume->GetMute(&isMuted))) {
                    pEndpointVolume->SetMute(!isMuted, NULL);
                }
                pEndpointVolume->Release();
            }
            pDevice->Release();
        }
        pEnumerator->Release();
    }
}

static void SendVolumeKey(BYTE vk) {
    UINT scanCode = MapVirtualKeyA(vk, 0);
    keybd_event(vk, static_cast<BYTE>(scanCode), 0x0001, 0);
    std::this_thread::sleep_for(std::chrono::milliseconds(15));
    keybd_event(vk, static_cast<BYTE>(scanCode), 0x0001 | 0x0002, 0);
}

extern "C" BRIDGE_API int StartNativeWinRTServices(
    NativeMediaCallback mediaCb,
    NativeNotificationCallback notifCb,
    NativeAudioPeakCallback peakCb
) {
    if (g_running) return 0;
    g_running = true;

    g_mediaCb = mediaCb;
    g_notifCb = notifCb;
    g_peakCb = peakCb;

    g_mediaSessionPollThread = std::thread([]() {
        ApartmentScope apt;
        try {
            g_sessionManager = winrt_media::GlobalSystemMediaTransportControlsSessionManager::RequestAsync().get();
            if (g_sessionManager) {
                g_currentSessionChangedToken = g_sessionManager.CurrentSessionChanged([](auto const&, auto const&) {
                    ScheduleEmit();
                });
                g_sessionsChangedToken = g_sessionManager.SessionsChanged([](auto const&, auto const&) {
                    ScheduleEmit();
                });
            }
        } catch (...) {}

        UpdateCurrentSession();

        while (g_running) {
            std::this_thread::sleep_for(std::chrono::seconds(2));
            if (g_running) {
                UpdateCurrentSession();
            }
        }
    });

    g_audioThread = std::thread(AudioPeakWorker);
    g_notifThread = std::thread(NotificationWorker);

    return 1;
}

extern "C" BRIDGE_API void StopNativeWinRTServices(void) {
    if (!g_running) return;
    g_running = false;

    if (g_mediaSessionPollThread.joinable()) g_mediaSessionPollThread.join();
    if (g_audioThread.joinable()) g_audioThread.join();
    if (g_notifThread.joinable()) g_notifThread.join();

    std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
    if (g_sessionManager) {
        try { g_sessionManager.CurrentSessionChanged(g_currentSessionChangedToken); } catch (...) {}
        try { g_sessionManager.SessionsChanged(g_sessionsChangedToken); } catch (...) {}
    }
    if (g_currentSession) {
        try { g_currentSession.MediaPropertiesChanged(g_mediaPropsChangedToken); } catch (...) {}
        try { g_currentSession.PlaybackInfoChanged(g_playbackInfoChangedToken); } catch (...) {}
        try { g_currentSession.TimelinePropertiesChanged(g_timelinePropsChangedToken); } catch (...) {}
        g_currentSession = nullptr;
    }
    g_sessionManager = nullptr;
}

extern "C" BRIDGE_API int SendNativeMediaCommand(const char* command) {
    if (!command) return 0;
    std::string cmd(command);

    if (cmd.rfind("set_volume:", 0) == 0) {
        float vol = std::stof(cmd.substr(11));
        SetMasterAudioVolume(vol);
        EmitMediaState(true);
        return 1;
    }
    if (cmd == "volume_up") {
        SendVolumeKey(VK_VOLUME_UP);
        EmitMediaState(true);
        return 1;
    }
    if (cmd == "volume_down") {
        SendVolumeKey(VK_VOLUME_DOWN);
        EmitMediaState(true);
        return 1;
    }
    if (cmd == "mute") {
        ToggleMasterAudioMute();
        SendVolumeKey(VK_VOLUME_MUTE);
        EmitMediaState(true);
        return 1;
    }
    if (cmd.rfind("seek:", 0) == 0) {
        double posMs = std::stod(cmd.substr(5));
        winrt_media::GlobalSystemMediaTransportControlsSession session{nullptr};
        {
            std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
            session = g_currentSession;
        }
        if (session) {
            try {
                auto ticks = static_cast<int64_t>(posMs * 10000.0);
                session.TryChangePlaybackPositionAsync(ticks);
            } catch (...) {}
        }
        return 1;
    }

    winrt_media::GlobalSystemMediaTransportControlsSession session{nullptr};
    {
        std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
        session = g_currentSession;
    }

    if (session) {
        try {
            if (cmd == "play_pause") session.TryTogglePlayPauseAsync();
            else if (cmd == "play") session.TryPlayAsync();
            else if (cmd == "pause") session.TryPauseAsync();
            else if (cmd == "next") session.TrySkipNextAsync();
            else if (cmd == "previous" || cmd == "prev") session.TrySkipPreviousAsync();
        } catch (...) {}

        std::thread([]() {
            ApartmentScope apt;
            std::this_thread::sleep_for(std::chrono::milliseconds(100));
            UpdateCurrentSession();
        }).detach();
    }

    return 1;
}
