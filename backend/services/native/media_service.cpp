/**
 * @file media_service.cpp
 * @brief SMTC media session tracking, optimistic seek handling, and WASAPI audio mixing.
 *
 * Consolidates WASAPI session/endpoint COM boilerplate, caches per-app volume reads
 * and process names to avoid kernel handle churn on media ticks, and enforces an
 * optimistic seek grace window so scrubbing the timeline never snaps back.
 */

#include "media_service.h"
#include "winrt_common.h"

#include <mmdeviceapi.h>
#include <endpointvolume.h>
#include <audiopolicy.h>
#include <propkey.h>
#include <functiondiscoverykeys_devpkey.h>
#include <psapi.h>
#include <shellapi.h>
#include <winrt/Windows.Media.Control.h>

#include <algorithm>
#include <atomic>
#include <cmath>
#include <condition_variable>
#include <cstdio>
#include <iomanip>
#include <mutex>
#include <sstream>
#include <string>
#include <thread>
#include <unordered_map>
#include <unordered_set>
#include <vector>

#ifndef PKEY_Device_FriendlyName
DEFINE_PROPERTYKEY(PKEY_Device_FriendlyName, 0xa45c254e, 0xdf1c, 0x4efd, 0x80, 0x20, 0x67, 0xd1, 0x46, 0xa8, 0x50, 0xe0,
                   14);
#endif

interface DECLSPEC_UUID("f8679f50-850a-41cf-9c72-430f290290c8") IPolicyConfig : public IUnknown {
public:
    virtual HRESULT STDMETHODCALLTYPE GetMixFormat(PCWSTR, WAVEFORMATEX**) = 0;
    virtual HRESULT STDMETHODCALLTYPE GetDeviceFormat(PCWSTR, INT, WAVEFORMATEX**) = 0;
    virtual HRESULT STDMETHODCALLTYPE ResetDeviceFormat(PCWSTR) = 0;
    virtual HRESULT STDMETHODCALLTYPE SetDeviceFormat(PCWSTR, WAVEFORMATEX*, WAVEFORMATEX*) = 0;
    virtual HRESULT STDMETHODCALLTYPE GetProcessingPeriod(PCWSTR, INT, PINT64, PINT64) = 0;
    virtual HRESULT STDMETHODCALLTYPE SetProcessingPeriod(PCWSTR, PINT64) = 0;
    virtual HRESULT STDMETHODCALLTYPE GetShareMode(PCWSTR, struct DeviceShareMode**) = 0;
    virtual HRESULT STDMETHODCALLTYPE SetShareMode(PCWSTR, struct DeviceShareMode*) = 0;
    virtual HRESULT STDMETHODCALLTYPE GetPropertyValue(PCWSTR, const PROPERTYKEY&, PROPVARIANT*) = 0;
    virtual HRESULT STDMETHODCALLTYPE SetPropertyValue(PCWSTR, const PROPERTYKEY&, PROPVARIANT*) = 0;
    virtual HRESULT STDMETHODCALLTYPE SetDefaultEndpoint(PCWSTR wszDeviceId, ERole eRole) = 0;
    virtual HRESULT STDMETHODCALLTYPE SetEndpointVisibility(PCWSTR, INT) = 0;
};

class DECLSPEC_UUID("870af99c-171d-4f9e-af0d-e63df40c2bc9") CPolicyConfigClient;

namespace winrt_media = winrt::Windows::Media::Control;

static std::atomic<NativeMediaCallback> g_mediaCb{nullptr};
static std::atomic<bool> g_mediaRunning{false};
static std::atomic<int> g_activeBgTasks{0};
static std::thread g_mediaSessionPollThread;
static std::condition_variable g_mediaStopCv;
static std::mutex g_mediaStopMutex;

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

static std::unordered_map<std::string, std::string> g_trackArtMap;
static std::vector<std::string> g_trackArtOrder;
static std::string g_currentTrackKey;
static std::string g_previousTrackArt;
static std::chrono::steady_clock::time_point g_trackStartTime{};
static std::mutex g_artCacheMutex;

static std::mutex g_audioVolMutex;

static std::string g_selectedSessionId;
static std::mutex g_selectedSessionMutex;

static std::atomic<uint64_t> g_artCorrectionGeneration{0};
static std::atomic<uint64_t> g_activeWorkerGen{0};
static std::chrono::steady_clock::time_point g_lastArtResyncTime{};
static std::mutex g_artResyncMutex;
static std::atomic<bool> g_mediaPropsDirty{false};

// Optimistic seek state to prevent timeline snap-back while players process seek IPC.
static std::mutex g_seekMutex;
static double g_seekTargetMs = -1.0;
static std::chrono::steady_clock::time_point g_seekTime{};
static std::atomic<uint64_t> g_seekGeneration{0};

static std::string g_cachedSessionsJson = "[]";
static std::chrono::steady_clock::time_point g_lastSessionsFetchTime{};
static std::mutex g_sessionsJsonMutex;

static float g_mediaVolumeStore = 100.0f;
static bool g_mediaMuteStore = false;
static std::string g_mediaVolCachedApp;
static std::chrono::steady_clock::time_point g_lastVolumeSetTime{};
static std::chrono::steady_clock::time_point g_lastVolumeReadTime{};
static std::mutex g_mediaVolStateMutex;

// PID -> process path cache to eliminate repeated OpenProcess calls during audio session scans.
static std::mutex g_pidPathMutex;
static std::unordered_map<DWORD, std::string> g_pidPathCache;
static std::chrono::steady_clock::time_point g_pidCacheClearTime = std::chrono::steady_clock::now();

static void UpdateCurrentSession();
static void EmitMediaState(bool includeVolume = true);
static void ScheduleEmit();
static std::pair<float, bool> GetMediaAudioVolumeAndMute(const std::string& sourceApp);

static std::string ToLowerStr(std::string s) {
    std::transform(s.begin(), s.end(), s.begin(), [](unsigned char c) { return static_cast<char>(std::tolower(c)); });
    return s;
}

static void TrimString(std::string& s) {
    size_t start = s.find_first_not_of(" \t\r\n");
    if (start == std::string::npos) {
        s.clear();
        return;
    }
    s = s.substr(start, s.find_last_not_of(" \t\r\n") - start + 1);
}

/**
 * @brief Sleeps for ms or returns false immediately if the media service is stopping.
 */
static bool SleepInterruptible(int ms) {
    std::unique_lock<std::mutex> lock(g_mediaStopMutex);
    return !g_mediaStopCv.wait_for(lock, std::chrono::milliseconds(ms),
                                   [] { return !g_mediaRunning.load(std::memory_order_acquire); });
}

/**
 * @brief Resolves and caches the full executable path for a process ID.
 */
static std::string GetCachedProcessPath(DWORD pid) {
    if (pid == 0)
        return "";
    auto now = std::chrono::steady_clock::now();
    {
        std::lock_guard<std::mutex> lock(g_pidPathMutex);
        if (std::chrono::duration_cast<std::chrono::seconds>(now - g_pidCacheClearTime).count() < 30) {
            auto it = g_pidPathCache.find(pid);
            if (it != g_pidPathCache.end())
                return it->second;
        }
    }

    std::string result;
    HANDLE hProc = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, FALSE, pid);
    if (hProc) {
        WCHAR wPath[MAX_PATH] = {0};
        DWORD size = MAX_PATH;
        if (QueryFullProcessImageNameW(hProc, 0, wPath, &size)) {
            result = Utf16ToUtf8(wPath);
        }
        CloseHandle(hProc);
    }

    if (!result.empty()) {
        std::lock_guard<std::mutex> lock(g_pidPathMutex);
        if (std::chrono::duration_cast<std::chrono::seconds>(now - g_pidCacheClearTime).count() >= 30 ||
            g_pidPathCache.size() > 256) {
            g_pidPathCache.clear();
            g_pidCacheClearTime = now;
        }
        g_pidPathCache[pid] = result;
    }
    return result;
}

static std::string ComputeSessionUniqueId(const std::string& sourceApp, std::string title, std::string artist) {
    TrimString(title);
    TrimString(artist);
    return sourceApp + "::" + title + "::" + artist;
}

struct ValidTrackSnapshot {
    std::string sourceApp, title, artist, album, albumArt;
    double posMs{0.0}, durMs{0.0};
    bool isPlaying{false}, canGoNext{true}, canGoPrev{true}, canPause{true}, canPlay{true};
    bool isShuffleActive{false}, isShuffleEnabled{true}, isRepeatEnabled{true};
    int autoRepeatMode{0};
    std::chrono::steady_clock::time_point timestamp;
    bool hasValue{false};
};

static std::mutex g_validSnapshotMutex;
static ValidTrackSnapshot g_lastValidTrackSnapshot;

void FlushMediaArtCache() {
    std::lock_guard<std::mutex> artLock(g_artCacheMutex);
    g_trackArtMap.clear();
    g_trackArtOrder.clear();
    g_currentTrackKey.clear();
    g_previousTrackArt.clear();
    g_trackStartTime = {};
    ++g_artCorrectionGeneration;
    {
        std::lock_guard<std::mutex> snapLock(g_validSnapshotMutex);
        g_lastValidTrackSnapshot = ValidTrackSnapshot{};
    }
}

void RequeryMediaArtwork() {
    g_mediaPropsDirty.store(true);
    {
        std::lock_guard<std::mutex> artLock(g_artCacheMutex);
        if (!g_currentTrackKey.empty()) {
            g_trackArtMap.erase(g_currentTrackKey);
            g_trackArtOrder.erase(std::remove(g_trackArtOrder.begin(), g_trackArtOrder.end(), g_currentTrackKey),
                                  g_trackArtOrder.end());
        }
        g_previousTrackArt.clear();
    }
    ++g_artCorrectionGeneration;
    ScheduleEmit();
}

void ForceEmitMediaState() {
    UpdateCurrentSession();
}

static std::string TryExtractThumbnail(
    winrt_media::GlobalSystemMediaTransportControlsSessionMediaProperties const& props,
    const std::string& staleArt = "", bool allowStale = false) {
    if (!props)
        return "";
    try {
        auto thumbRef = props.Thumbnail();
        if (!thumbRef)
            return "";
        winrt_storage::IRandomAccessStreamWithContentType stream{nullptr};
        try {
            stream = thumbRef.OpenReadAsync().get();
        } catch (...) {
            if (!SleepInterruptible(50))
                return "";
            try {
                stream = thumbRef.OpenReadAsync().get();
            } catch (...) {
            }
        }
        if (stream) {
            std::string fetched = StreamToBase64(stream);
            if (!fetched.empty() && (allowStale || staleArt.empty() || fetched != staleArt)) {
                return fetched;
            }
        }
    } catch (...) {
    }
    return "";
}

bool IsMediaPlayingState() {
    return g_isMediaPlaying.load(std::memory_order_acquire);
}

int SelectMediaSession(const char* sessionId) {
    {
        std::lock_guard<std::mutex> lock(g_selectedSessionMutex);
        g_selectedSessionId = sessionId ? sessionId : "";
    }
    {
        std::lock_guard<std::mutex> sLock(g_sessionsJsonMutex);
        g_lastSessionsFetchTime = {};
    }
    UpdateCurrentSession();
    return 1;
}

struct WindowFocusData {
    std::string needle;
    std::string needleAumid;
    HWND foundHwnd = nullptr;
};

int FocusApp(const char* sourceApp, const char* aumid) {
    std::string needle = (sourceApp && *sourceApp) ? ToLowerStr(sourceApp) : "";
    std::string rawAumid = (aumid && *aumid) ? aumid : "";
    std::string aumidToken = ToLowerStr(rawAumid);

    if (size_t slash = needle.find_last_of("\\/"); slash != std::string::npos) {
        needle = needle.substr(slash + 1);
    }
    if (needle.size() > 4 && needle.substr(needle.size() - 4) == ".exe") {
        needle.resize(needle.size() - 4);
    }
    if (size_t dot = aumidToken.find_last_of('.'); dot != std::string::npos && dot + 1 < aumidToken.size()) {
        aumidToken = aumidToken.substr(dot + 1);
    }
    if (size_t bang = aumidToken.find('!'); bang != std::string::npos) {
        aumidToken = aumidToken.substr(0, bang);
    }

    WindowFocusData data{needle, aumidToken, nullptr};

    EnumWindows(
        [](HWND hwnd, LPARAM lParam) -> BOOL {
            if (!IsWindowVisible(hwnd) || (GetWindowLongW(hwnd, GWL_EXSTYLE) & WS_EX_TOOLWINDOW)) {
                return TRUE;
            }
            if (GetWindowTextLengthW(hwnd) <= 0)
                return TRUE;

            DWORD pid = 0;
            GetWindowThreadProcessId(hwnd, &pid);
            std::string exePath = ToLowerStr(GetCachedProcessPath(pid));
            if (exePath.empty())
                return TRUE;

            auto* d = reinterpret_cast<WindowFocusData*>(lParam);
            bool match =
                (!d->needle.empty() && exePath.find(d->needle) != std::string::npos) ||
                (!d->needleAumid.empty() && exePath.find(d->needleAumid) != std::string::npos) ||
                (d->needle.find("spotify") != std::string::npos && exePath.find("spotify") != std::string::npos);
            if (match) {
                d->foundHwnd = hwnd;
                return FALSE;
            }
            return TRUE;
        },
        reinterpret_cast<LPARAM>(&data));

    if (data.foundHwnd) {
        if (IsIconic(data.foundHwnd))
            ShowWindow(data.foundHwnd, SW_RESTORE);
        DWORD foreThread = GetWindowThreadProcessId(GetForegroundWindow(), nullptr);
        DWORD appThread = GetCurrentThreadId();
        if (foreThread != appThread) {
            AttachThreadInput(foreThread, appThread, TRUE);
            BringWindowToTop(data.foundHwnd);
            SetForegroundWindow(data.foundHwnd);
            AttachThreadInput(foreThread, appThread, FALSE);
        } else {
            BringWindowToTop(data.foundHwnd);
            SetForegroundWindow(data.foundHwnd);
        }
        return 1;
    }

    if (!rawAumid.empty()) {
        std::wstring target = L"shell:AppsFolder\\" + Utf8ToUtf16(rawAumid);
        if (reinterpret_cast<INT_PTR>(
                ShellExecuteW(nullptr, L"open", target.c_str(), nullptr, nullptr, SW_SHOWNORMAL)) > 32) {
            return 1;
        }
    }
    return 0;
}

int FocusMediaApp(const char* sourceApp) {
    return FocusApp(sourceApp, "");
}

static void ScheduleEmit() {
    g_activeBgTasks.fetch_add(1, std::memory_order_acq_rel);
    if (!g_mediaRunning.load(std::memory_order_acquire)) {
        g_activeBgTasks.fetch_sub(1, std::memory_order_acq_rel);
        return;
    }
    bool expected = false;
    if (!g_emitPending.compare_exchange_strong(expected, true)) {
        g_activeBgTasks.fetch_sub(1, std::memory_order_acq_rel);
        return;
    }
    std::thread([]() {
        ApartmentScope apt;
        if (SleepInterruptible(40)) {
            g_emitPending.store(false);
            UpdateCurrentSession();
        } else {
            g_emitPending.store(false);
        }
        g_activeBgTasks.fetch_sub(1, std::memory_order_acq_rel);
    }).detach();
}

static bool IsValidMediaSession(
    winrt_media::GlobalSystemMediaTransportControlsSession const& session,
    winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackInfo const* prefetchedInfo = nullptr,
    winrt_media::GlobalSystemMediaTransportControlsSessionMediaProperties const* prefetchedProps = nullptr) {
    if (!session)
        return false;
    try {
        auto info = (prefetchedInfo && *prefetchedInfo) ? *prefetchedInfo : session.GetPlaybackInfo();
        if (!info ||
            info.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Closed) {
            return false;
        }

        auto controls = info.Controls();
        if (!controls)
            return false;
        if (!(controls.IsPlayEnabled() || controls.IsPauseEnabled() || controls.IsPlayPauseToggleEnabled() ||
              controls.IsStopEnabled() || controls.IsNextEnabled() || controls.IsPreviousEnabled())) {
            return false;
        }

        auto props =
            (prefetchedProps && *prefetchedProps) ? *prefetchedProps : session.TryGetMediaPropertiesAsync().get();
        if (!props)
            return false;

        std::string title = Utf16ToUtf8(props.Title().c_str());
        std::string artist = Utf16ToUtf8(props.Artist().c_str());
        TrimString(title);
        TrimString(artist);
        if (title.empty() && artist.empty())
            return false;

        bool hasAlbumArt = (props.Thumbnail() != nullptr);
        bool hasPrevNext = (controls.IsNextEnabled() || controls.IsPreviousEnabled());
        if (artist.empty() && (!hasAlbumArt || !hasPrevNext)) {
            return false;
        }

        if (artist.empty()) {
            std::string appName;
            try {
                appName = Utf16ToUtf8(session.SourceAppUserModelId().c_str());
            } catch (...) {
            }
            if (size_t slash = appName.find_last_of("\\/"); slash != std::string::npos)
                appName = appName.substr(slash + 1);
            if (size_t dot = appName.rfind('.'); dot != std::string::npos)
                appName = appName.substr(0, dot);
            if (!appName.empty() && ToLowerStr(title) == ToLowerStr(appName)) {
                return false;
            }
        }
        return true;
    } catch (...) {
        return false;
    }
}

static bool IsSessionAliveAndValid(
    winrt_media::GlobalSystemMediaTransportControlsSession const& session,
    winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackInfo const* prefetchedInfo = nullptr,
    winrt_media::GlobalSystemMediaTransportControlsSessionMediaProperties const* prefetchedProps = nullptr) {
    if (!session)
        return false;
    try {
        auto info = (prefetchedInfo && *prefetchedInfo) ? *prefetchedInfo : session.GetPlaybackInfo();
        if (!info ||
            info.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Closed) {
            return false;
        }
        if (IsValidMediaSession(session, &info, prefetchedProps))
            return true;

        std::lock_guard<std::mutex> snapLock(g_validSnapshotMutex);
        if (g_lastValidTrackSnapshot.hasValue) {
            std::string sApp;
            try {
                sApp = Utf16ToUtf8(session.SourceAppUserModelId().c_str());
            } catch (...) {
            }
            if (sApp == g_lastValidTrackSnapshot.sourceApp) {
                auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(std::chrono::steady_clock::now() -
                                                                                g_lastValidTrackSnapshot.timestamp)
                                   .count();
                if (elapsed < 15)
                    return true;
            }
        }
        return false;
    } catch (...) {
        return false;
    }
}

static void UpdateCurrentSession() {
    if (!g_mediaRunning.load(std::memory_order_acquire))
        return;
    ApartmentScope apt;
    std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
    if (!g_sessionManager)
        return;

    try {
        std::string selectedId;
        {
            std::lock_guard<std::mutex> sLock(g_selectedSessionMutex);
            selectedId = g_selectedSessionId;
        }

        winrt_media::GlobalSystemMediaTransportControlsSession targetSession{nullptr};
        winrt_media::GlobalSystemMediaTransportControlsSession chosenSession{nullptr};

        if (!selectedId.empty()) {
            try {
                auto sessions = g_sessionManager.GetSessions();
                for (auto const& s : sessions) {
                    try {
                        std::string sId = Utf16ToUtf8(s.SourceAppUserModelId().c_str());
                        std::string sTitle, sArtist;
                        if (auto props = s.TryGetMediaPropertiesAsync().get()) {
                            sTitle = Utf16ToUtf8(props.Title().c_str());
                            sArtist = Utf16ToUtf8(props.Artist().c_str());
                            TrimString(sTitle);
                            TrimString(sArtist);
                        }
                        std::string uniqueId = ComputeSessionUniqueId(sId, sTitle, sArtist);
                        if (selectedId == uniqueId ||
                            (!sTitle.empty() && selectedId.find(sTitle) != std::string::npos)) {
                            chosenSession = s;
                            break;
                        }
                    } catch (...) {
                    }
                }
                if (!chosenSession) {
                    for (auto const& s : sessions) {
                        try {
                            std::string sId = Utf16ToUtf8(s.SourceAppUserModelId().c_str());
                            if (!sId.empty() && selectedId.rfind(sId, 0) == 0) {
                                chosenSession = s;
                                break;
                            }
                        } catch (...) {
                        }
                    }
                }
            } catch (...) {
            }
        }

        if (chosenSession && IsSessionAliveAndValid(chosenSession)) {
            targetSession = chosenSession;
        } else {
            if (!selectedId.empty()) {
                std::lock_guard<std::mutex> sLock(g_selectedSessionMutex);
                g_selectedSessionId.clear();
            }

            bool currentIsPlaying = false;
            if (g_currentSession && IsValidMediaSession(g_currentSession)) {
                try {
                    auto curInfo = g_currentSession.GetPlaybackInfo();
                    currentIsPlaying =
                        (curInfo && curInfo.PlaybackStatus() ==
                                        winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing);
                } catch (...) {
                }
            }

            winrt_media::GlobalSystemMediaTransportControlsSession focusedPlaying{nullptr};
            try {
                auto focused = g_sessionManager.GetCurrentSession();
                if (focused && IsValidMediaSession(focused)) {
                    auto info = focused.GetPlaybackInfo();
                    if (info && info.PlaybackStatus() ==
                                    winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing) {
                        focusedPlaying = focused;
                    }
                }
            } catch (...) {
            }

            if (currentIsPlaying) {
                targetSession =
                    (focusedPlaying && focusedPlaying != g_currentSession) ? focusedPlaying : g_currentSession;
            } else if (focusedPlaying) {
                targetSession = focusedPlaying;
            }

            if (!targetSession) {
                try {
                    for (auto const& s : g_sessionManager.GetSessions()) {
                        try {
                            auto info = s.GetPlaybackInfo();
                            if (info &&
                                info.PlaybackStatus() ==
                                    winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing &&
                                IsValidMediaSession(s, &info)) {
                                targetSession = s;
                                break;
                            }
                        } catch (...) {
                        }
                    }
                } catch (...) {
                }
            }

            if (!targetSession) {
                if (g_currentSession && IsSessionAliveAndValid(g_currentSession)) {
                    targetSession = g_currentSession;
                } else {
                    try {
                        auto focused = g_sessionManager.GetCurrentSession();
                        if (focused && IsValidMediaSession(focused))
                            targetSession = focused;
                    } catch (...) {
                    }

                    if (!targetSession) {
                        try {
                            for (auto const& s : g_sessionManager.GetSessions()) {
                                if (IsValidMediaSession(s)) {
                                    targetSession = s;
                                    break;
                                }
                            }
                        } catch (...) {
                        }
                    }
                }
            }
        }

        if (g_currentSession != targetSession) {
            {
                std::lock_guard<std::mutex> sLock(g_sessionsJsonMutex);
                g_lastSessionsFetchTime = {};
            }
            ++g_artCorrectionGeneration;
            g_mediaPropsDirty.store(true);
            if (g_currentSession) {
                try {
                    g_currentSession.MediaPropertiesChanged(g_mediaPropsChangedToken);
                } catch (...) {
                }
                try {
                    g_currentSession.PlaybackInfoChanged(g_playbackInfoChangedToken);
                } catch (...) {
                }
                try {
                    g_currentSession.TimelinePropertiesChanged(g_timelinePropsChangedToken);
                } catch (...) {
                }
            }

            g_currentSession = targetSession;

            if (g_currentSession) {
                try {
                    g_mediaPropsChangedToken = g_currentSession.MediaPropertiesChanged([](auto const&, auto const&) {
                        g_mediaPropsDirty.store(true);
                        ScheduleEmit();
                    });
                    g_playbackInfoChangedToken =
                        g_currentSession.PlaybackInfoChanged([](auto const&, auto const&) { ScheduleEmit(); });
                    g_timelinePropsChangedToken =
                        g_currentSession.TimelinePropertiesChanged([](auto const&, auto const&) { ScheduleEmit(); });
                } catch (...) {
                }
            }
        }
    } catch (...) {
    }

    EmitMediaState(true);
}

static void EmitMediaState(bool includeVolume) {
    (void)includeVolume;
    NativeMediaCallback cb = g_mediaCb.load(std::memory_order_acquire);
    if (!cb || !g_mediaRunning.load(std::memory_order_acquire))
        return;

    std::string sourceApp, title, artist, album, albumArt;
    bool isPlaying = false, canGoNext = true, canGoPrev = true, canPause = true, canPlay = true;
    bool isShuffleActive = false, isShuffleEnabled = true, isRepeatEnabled = true;
    int autoRepeatMode = 0;
    double posMs = 0.0, durMs = 0.0;

    winrt_media::GlobalSystemMediaTransportControlsSession session{nullptr};
    {
        std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
        session = g_currentSession;
    }

    winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackInfo prefetchInfo{nullptr};
    winrt_media::GlobalSystemMediaTransportControlsSessionMediaProperties prefetchProps{nullptr};
    if (session) {
        try {
            prefetchInfo = session.GetPlaybackInfo();
        } catch (...) {
        }
        try {
            prefetchProps = session.TryGetMediaPropertiesAsync().get();
        } catch (...) {
        }
    }

    bool useSnapshot = false;
    if (!session || !IsValidMediaSession(session, &prefetchInfo, &prefetchProps)) {
        std::lock_guard<std::mutex> snapLock(g_validSnapshotMutex);
        if (session && g_lastValidTrackSnapshot.hasValue && prefetchInfo &&
            prefetchInfo.PlaybackStatus() !=
                winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Closed) {
            auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(std::chrono::steady_clock::now() -
                                                                            g_lastValidTrackSnapshot.timestamp)
                               .count();
            if (elapsed < 15)
                useSnapshot = true;
        }

        if (!useSnapshot) {
            g_isMediaPlaying.store(false, std::memory_order_release);
            cb("{\"type\":\"media\",\"isAvailable\":false}");
            return;
        }
    }

    if (useSnapshot) {
        std::lock_guard<std::mutex> snapLock(g_validSnapshotMutex);
        const auto& snap = g_lastValidTrackSnapshot;
        sourceApp = snap.sourceApp;
        title = snap.title;
        artist = snap.artist;
        album = snap.album;
        albumArt = snap.albumArt;
        isPlaying = snap.isPlaying;
        durMs = snap.durMs;
        canGoNext = snap.canGoNext;
        canGoPrev = snap.canGoPrev;
        canPause = snap.canPause;
        canPlay = snap.canPlay;
        isShuffleActive = snap.isShuffleActive;
        isShuffleEnabled = snap.isShuffleEnabled;
        autoRepeatMode = snap.autoRepeatMode;
        isRepeatEnabled = snap.isRepeatEnabled;

        if (isPlaying) {
            double elapsedMs =
                std::chrono::duration<double, std::milli>(std::chrono::steady_clock::now() - snap.timestamp).count();
            posMs = (durMs > 0.0) ? std::min(durMs, snap.posMs + elapsedMs) : (snap.posMs + elapsedMs);
        } else {
            posMs = snap.posMs;
        }
        g_isMediaPlaying.store(isPlaying, std::memory_order_release);
    } else {
        try {
            sourceApp = Utf16ToUtf8(session.SourceAppUserModelId().c_str());
        } catch (...) {
        }

        if (prefetchProps) {
            title = Utf16ToUtf8(prefetchProps.Title().c_str());
            artist = Utf16ToUtf8(prefetchProps.Artist().c_str());
            album = Utf16ToUtf8(prefetchProps.AlbumTitle().c_str());

            std::string t = title, a = artist;
            TrimString(t);
            TrimString(a);
            std::string trackKey = t + "_" + a;
            std::string staleArt;
            bool isNewTrack = false;
            auto now = std::chrono::steady_clock::now();

            {
                std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                if (trackKey != g_currentTrackKey) {
                    isNewTrack = true;
                    // Clear any active seek override when switching tracks
                    {
                        std::lock_guard<std::mutex> skLock(g_seekMutex);
                        g_seekTargetMs = -1.0;
                    }
                    auto prevIt = g_trackArtMap.find(g_currentTrackKey);
                    if (prevIt != g_trackArtMap.end() && !prevIt->second.empty()) {
                        g_previousTrackArt = prevIt->second;
                    }
                    g_currentTrackKey = trackKey;
                    g_trackStartTime = now;
                    ++g_artCorrectionGeneration;
                } else if (std::chrono::duration_cast<std::chrono::milliseconds>(now - g_trackStartTime).count() >
                           1000) {
                    g_previousTrackArt.clear();
                }
                staleArt = g_previousTrackArt;

                auto it = g_trackArtMap.find(trackKey);
                if (it != g_trackArtMap.end())
                    albumArt = it->second;
            }

            bool allowStale =
                (std::chrono::duration_cast<std::chrono::milliseconds>(now - g_trackStartTime).count() >= 400);
            bool propsDirty = g_mediaPropsDirty.exchange(false);

            if ((albumArt.empty() || propsDirty) && prefetchProps.Thumbnail() != nullptr) {
                std::string fetchedArt = TryExtractThumbnail(prefetchProps, staleArt, allowStale);
                if (!fetchedArt.empty() && fetchedArt != albumArt) {
                    std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                    g_trackArtMap[trackKey] = fetchedArt;
                    g_trackArtOrder.push_back(trackKey);
                    while (g_trackArtMap.size() > 30 && !g_trackArtOrder.empty()) {
                        std::string oldest = g_trackArtOrder.front();
                        g_trackArtOrder.erase(g_trackArtOrder.begin());
                        if (oldest != trackKey)
                            g_trackArtMap.erase(oldest);
                    }
                    albumArt = fetchedArt;
                }
            }

            if (albumArt.empty() && !trackKey.empty() && !isNewTrack) {
                bool shouldResync = false;
                {
                    std::lock_guard<std::mutex> resyncLock(g_artResyncMutex);
                    if (std::chrono::duration_cast<std::chrono::seconds>(now - g_lastArtResyncTime).count() >= 10) {
                        g_lastArtResyncTime = now;
                        shouldResync = true;
                    }
                }
                if (shouldResync) {
                    std::string freshArt = TryExtractThumbnail(prefetchProps, staleArt, true);
                    if (!freshArt.empty()) {
                        std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                        g_trackArtMap[trackKey] = freshArt;
                        g_trackArtOrder.push_back(trackKey);
                        while (g_trackArtMap.size() > 30 && !g_trackArtOrder.empty()) {
                            std::string oldest = g_trackArtOrder.front();
                            g_trackArtOrder.erase(g_trackArtOrder.begin());
                            if (oldest != trackKey)
                                g_trackArtMap.erase(oldest);
                        }
                        albumArt = freshArt;
                    }
                }
            }

            uint64_t curGen = g_artCorrectionGeneration.load();
            bool needsWorker = (albumArt.empty() || albumArt.size() < 10000 || propsDirty) && !trackKey.empty();
            if (needsWorker && g_activeWorkerGen.load() != curGen) {
                g_activeWorkerGen.store(curGen);
                {
                    std::lock_guard<std::mutex> resyncLock(g_artResyncMutex);
                    g_lastArtResyncTime = now;
                }
                std::string currentDeliveredArt = albumArt;
                g_activeBgTasks.fetch_add(1, std::memory_order_acq_rel);
                std::thread([curGen, trackKey, currentDeliveredArt]() mutable {
                    ApartmentScope apt;
                    const int delaysMs[] = {300, 600, 1000, 1500, 2200, 3200, 4500};

                    for (int delay : delaysMs) {
                        if (!SleepInterruptible(delay) || g_artCorrectionGeneration.load() != curGen)
                            break;
                        try {
                            winrt_media::GlobalSystemMediaTransportControlsSession curSession{nullptr};
                            {
                                std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
                                curSession = g_currentSession;
                            }
                            if (!curSession)
                                break;
                            auto p = curSession.TryGetMediaPropertiesAsync().get();
                            if (!p)
                                continue;

                            std::string pTitle = Utf16ToUtf8(p.Title().c_str());
                            std::string pArtist = Utf16ToUtf8(p.Artist().c_str());
                            TrimString(pTitle);
                            TrimString(pArtist);
                            if ((pTitle + "_" + pArtist) != trackKey)
                                break;

                            std::string newArt = TryExtractThumbnail(p, "", delay >= 500);
                            if (!newArt.empty() && newArt != currentDeliveredArt) {
                                bool changed = false;
                                {
                                    std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                                    if (g_trackArtMap[trackKey] != newArt) {
                                        g_trackArtMap[trackKey] = newArt;
                                        g_trackArtOrder.push_back(trackKey);
                                        changed = true;
                                    }
                                }
                                if (changed) {
                                    currentDeliveredArt = newArt;
                                    ScheduleEmit();
                                    if (newArt.size() > 10000)
                                        break;
                                }
                            }
                        } catch (...) {
                        }
                    }
                    g_activeBgTasks.fetch_sub(1, std::memory_order_acq_rel);
                }).detach();
            }
        }

        if (prefetchInfo) {
            isPlaying = (prefetchInfo.PlaybackStatus() ==
                         winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing);
            g_isMediaPlaying.store(isPlaying, std::memory_order_release);

            if (auto shuffleRef = prefetchInfo.IsShuffleActive())
                isShuffleActive = shuffleRef.Value();
            if (auto repeatRef = prefetchInfo.AutoRepeatMode()) {
                auto mode = repeatRef.Value();
                if (mode == winrt::Windows::Media::MediaPlaybackAutoRepeatMode::Track)
                    autoRepeatMode = 1;
                else if (mode == winrt::Windows::Media::MediaPlaybackAutoRepeatMode::List)
                    autoRepeatMode = 2;
                else
                    autoRepeatMode = 0;
            }
            if (auto controls = prefetchInfo.Controls()) {
                canGoNext = controls.IsNextEnabled();
                canGoPrev = controls.IsPreviousEnabled();
                canPause = controls.IsPauseEnabled();
                canPlay = controls.IsPlayEnabled();
                if (controls.IsPlayPauseToggleEnabled()) {
                    if (isPlaying)
                        canPause = true;
                    else
                        canPlay = true;
                }
                isShuffleEnabled = controls.IsShuffleEnabled();
                isRepeatEnabled = controls.IsRepeatEnabled();
            }
        }

        if (title.empty() && artist.empty()) {
            g_isMediaPlaying.store(false, std::memory_order_release);
            cb("{\"type\":\"media\",\"isAvailable\":false}");
            return;
        }

        try {
            if (auto timeline = session.GetTimelineProperties()) {
                double rawPosMs = std::chrono::duration<double, std::milli>(timeline.Position()).count();
                durMs = std::chrono::duration<double, std::milli>(timeline.EndTime()).count();

                if (isPlaying) {
                    auto lastUpdated = timeline.LastUpdatedTime();
                    auto now = winrt::clock::now();
                    if (now > lastUpdated) {
                        double elapsedMs = std::chrono::duration<double, std::milli>(now - lastUpdated).count();
                        posMs = (durMs > 0.0) ? std::min(durMs, rawPosMs + elapsedMs) : (rawPosMs + elapsedMs);
                    } else {
                        posMs = rawPosMs;
                    }
                } else {
                    posMs = rawPosMs;
                }
            }
        } catch (...) {
        }

        // Apply optimistic seek lock if the user recently scrubbed the timeline and SMTC hasn't caught up yet.
        {
            std::lock_guard<std::mutex> skLock(g_seekMutex);
            if (g_seekTargetMs >= 0.0) {
                double seekElapsed =
                    std::chrono::duration<double, std::milli>(std::chrono::steady_clock::now() - g_seekTime).count();
                if (seekElapsed < 1200.0) {
                    double expectedPos = isPlaying ? (g_seekTargetMs + seekElapsed) : g_seekTargetMs;
                    if (durMs > 0.0 && expectedPos > durMs)
                        expectedPos = durMs;
                    if (std::fabs(posMs - expectedPos) > 1500.0) {
                        posMs = expectedPos;
                    } else {
                        g_seekTargetMs = -1.0; // Player caught up to the seek target
                    }
                } else {
                    g_seekTargetMs = -1.0;
                }
            }
        }

        if (!title.empty() && (!artist.empty() || !albumArt.empty())) {
            std::lock_guard<std::mutex> snapLock(g_validSnapshotMutex);
            g_lastValidTrackSnapshot = ValidTrackSnapshot{sourceApp,
                                                          title,
                                                          artist,
                                                          album,
                                                          albumArt,
                                                          posMs,
                                                          durMs,
                                                          isPlaying,
                                                          canGoNext,
                                                          canGoPrev,
                                                          canPause,
                                                          canPlay,
                                                          isShuffleActive,
                                                          isShuffleEnabled,
                                                          isRepeatEnabled,
                                                          autoRepeatMode,
                                                          std::chrono::steady_clock::now(),
                                                          true};
        }
    }

    std::string sessionsJson = "[]";
    std::string activeSessionId = ComputeSessionUniqueId(sourceApp, title, artist);
    try {
        if (g_sessionManager) {
            auto now = std::chrono::steady_clock::now();
            std::lock_guard<std::mutex> sLock(g_sessionsJsonMutex);
            auto elapsedMs =
                std::chrono::duration_cast<std::chrono::milliseconds>(now - g_lastSessionsFetchTime).count();

            if (elapsedMs < 1500 && g_cachedSessionsJson != "[]") {
                sessionsJson = g_cachedSessionsJson;
            } else {
                std::ostringstream sJson;
                sJson << "[";
                bool first = true;
                std::unordered_set<std::string> seenUniqueIds;

                for (auto const& s : g_sessionManager.GetSessions()) {
                    if (!s)
                        continue;
                    try {
                        auto pInfo = s.GetPlaybackInfo();
                        if (!pInfo ||
                            pInfo.PlaybackStatus() ==
                                winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Closed) {
                            continue;
                        }
                        auto props =
                            (s == session && prefetchProps) ? prefetchProps : s.TryGetMediaPropertiesAsync().get();
                        if (!props || !IsValidMediaSession(s, &pInfo, &props))
                            continue;

                        std::string sId = Utf16ToUtf8(s.SourceAppUserModelId().c_str());
                        std::string sApp = sId;
                        if (size_t slash = sApp.find_last_of("\\/"); slash != std::string::npos)
                            sApp = sApp.substr(slash + 1);

                        std::string sTitle = Utf16ToUtf8(props.Title().c_str());
                        std::string sArtist = Utf16ToUtf8(props.Artist().c_str());
                        TrimString(sTitle);
                        TrimString(sArtist);
                        if (sTitle.empty() && sArtist.empty())
                            continue;

                        std::string uniqueId = ComputeSessionUniqueId(sId, sTitle, sArtist);
                        if (!seenUniqueIds.insert(uniqueId).second)
                            continue;

                        if (s == session || (sId == sourceApp && sTitle == title && sArtist == artist)) {
                            activeSessionId = uniqueId;
                        }

                        bool sPlaying = (pInfo.PlaybackStatus() ==
                                         winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing);
                        if (!first)
                            sJson << ",";
                        first = false;
                        sJson << "{\"id\":\"" << EscapeJsonString(uniqueId) << "\",\"sourceApp\":\""
                              << EscapeJsonString(sApp) << "\",\"title\":\"" << EscapeJsonString(sTitle)
                              << "\",\"artist\":\"" << EscapeJsonString(sArtist)
                              << "\",\"isPlaying\":" << (sPlaying ? "true" : "false") << "}";
                    } catch (...) {
                    }
                }
                sJson << "]";
                g_cachedSessionsJson = sJson.str();
                g_lastSessionsFetchTime = now;
                sessionsJson = g_cachedSessionsJson;
            }
        }
    } catch (...) {
    }

    auto [volume, isMuted] = GetMediaAudioVolumeAndMute(sourceApp);

    std::ostringstream json;
    json << "{\"type\":\"media\",\"isAvailable\":true,"
         << "\"sourceApp\":\"" << EscapeJsonString(sourceApp) << "\","
         << "\"title\":\"" << EscapeJsonString(title) << "\","
         << "\"artist\":\"" << EscapeJsonString(artist) << "\","
         << "\"album\":\"" << EscapeJsonString(album) << "\","
         << "\"albumArt\":\"" << EscapeJsonString(albumArt) << "\","
         << "\"isPlaying\":" << (isPlaying ? "true" : "false") << ","
         << "\"isShuffleActive\":" << (isShuffleActive ? "true" : "false") << ","
         << "\"autoRepeatMode\":" << autoRepeatMode << ","
         << "\"isShuffleEnabled\":" << (isShuffleEnabled ? "true" : "false") << ","
         << "\"isRepeatEnabled\":" << (isRepeatEnabled ? "true" : "false") << ","
         << "\"audioPeak\":0.0,"
         << "\"positionMs\":" << std::fixed << std::setprecision(0) << posMs << ","
         << "\"durationMs\":" << std::fixed << std::setprecision(0) << durMs << ","
         << "\"volume\":" << static_cast<int>(std::round(volume)) << ","
         << "\"isMuted\":" << (isMuted ? "true" : "false") << ","
         << "\"canGoNext\":" << (canGoNext ? "true" : "false") << ","
         << "\"canGoPrev\":" << (canGoPrev ? "true" : "false") << ","
         << "\"canPause\":" << (canPause ? "true" : "false") << ","
         << "\"canPlay\":" << (canPlay ? "true" : "false") << ","
         << "\"sessions\":" << sessionsJson << ","
         << "\"activeSessionId\":\"" << EscapeJsonString(activeSessionId) << "\"}";

    cb(json.str().c_str());
}

// --- WASAPI Audio Session & Endpoint Helpers ---

/**
 * @brief Acquires the WASAPI IAudioSessionEnumerator for the default multimedia render endpoint.
 */
static IAudioSessionEnumerator* AcquireAudioSessionEnumerator() {
    IMMDeviceEnumerator* pEnumerator = nullptr;
    if (FAILED(CoCreateInstance(__uuidof(MMDeviceEnumerator), nullptr, CLSCTX_INPROC_SERVER,
                                __uuidof(IMMDeviceEnumerator), reinterpret_cast<void**>(&pEnumerator))) ||
        !pEnumerator) {
        return nullptr;
    }
    IMMDevice* pDevice = nullptr;
    HRESULT hr = pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDevice);
    pEnumerator->Release();
    if (FAILED(hr) || !pDevice)
        return nullptr;

    IAudioSessionManager2* pSessionMgr = nullptr;
    hr = pDevice->Activate(__uuidof(IAudioSessionManager2), CLSCTX_INPROC_SERVER, nullptr,
                           reinterpret_cast<void**>(&pSessionMgr));
    pDevice->Release();
    if (FAILED(hr) || !pSessionMgr)
        return nullptr;

    IAudioSessionEnumerator* pSessionEnum = nullptr;
    pSessionMgr->GetSessionEnumerator(&pSessionEnum);
    pSessionMgr->Release();
    return pSessionEnum;
}

/**
 * @brief Acquires the master IAudioEndpointVolume interface for the default multimedia render device.
 */
static IAudioEndpointVolume* AcquireMasterEndpointVolume() {
    IMMDeviceEnumerator* pEnumerator = nullptr;
    if (FAILED(CoCreateInstance(__uuidof(MMDeviceEnumerator), nullptr, CLSCTX_INPROC_SERVER,
                                __uuidof(IMMDeviceEnumerator), reinterpret_cast<void**>(&pEnumerator))) ||
        !pEnumerator) {
        return nullptr;
    }
    IMMDevice* pDevice = nullptr;
    HRESULT hr = pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDevice);
    pEnumerator->Release();
    if (FAILED(hr) || !pDevice)
        return nullptr;

    IAudioEndpointVolume* pEndVol = nullptr;
    pDevice->Activate(__uuidof(IAudioEndpointVolume), CLSCTX_INPROC_SERVER, nullptr,
                      reinterpret_cast<void**>(&pEndVol));
    pDevice->Release();
    return pEndVol;
}

static std::vector<std::string> ExtractAppTokens(const std::string& sourceApp) {
    std::vector<std::string> tokens;
    if (sourceApp.empty())
        return tokens;

    std::string s = ToLowerStr(sourceApp);
    const std::pair<const char*, const char*> aliases[] = {{"308046b0af4a39cb", "firefox"},
                                                           {"firefox", "firefox"},
                                                           {"spotify", "spotify"},
                                                           {"apple", "applemusic"},
                                                           {"apple", "apple"},
                                                           {"chrome", "chrome"},
                                                           {"edge", "msedge"},
                                                           {"edge", "edge"},
                                                           {"vlc", "vlc"},
                                                           {"zune", "zune"},
                                                           {"zune", "wmplayer"},
                                                           {"zune", "media"},
                                                           {"tidal", "tidal"},
                                                           {"brave", "brave"},
                                                           {"opera", "opera"}};
    for (const auto& kv : aliases) {
        if (s.find(kv.first) != std::string::npos)
            tokens.push_back(kv.second);
    }

    std::string cur;
    auto flushToken = [&]() {
        if (cur.length() >= 3 && cur != "exe" && cur != "app" && cur != "win" && cur != "inc") {
            tokens.push_back(cur);
        }
        cur.clear();
    };
    for (char c : s) {
        if (std::isalnum(static_cast<unsigned char>(c)))
            cur += c;
        else
            flushToken();
    }
    flushToken();

    size_t lastSlash = s.find_last_of("\\/");
    std::string base = (lastSlash != std::string::npos) ? s.substr(lastSlash + 1) : s;
    if (size_t pExe = base.find(".exe"); pExe != std::string::npos)
        base.resize(pExe);
    if (base.length() > 2)
        tokens.push_back(base);

    return tokens;
}

struct SessionMatch {
    ISimpleAudioVolume* pVol = nullptr;
    int score = 0;
    float peak = 0.0f;
    bool isActive = false;
};

static bool GetAllAppSessionVolumes(const std::string& sourceApp, std::vector<ISimpleAudioVolume*>& outVolumes,
                                    ISimpleAudioVolume** ppPrimaryVolume) {
    if (ppPrimaryVolume)
        *ppPrimaryVolume = nullptr;
    outVolumes.clear();

    IAudioSessionEnumerator* pSessionEnum = AcquireAudioSessionEnumerator();
    if (!pSessionEnum)
        return false;

    int count = 0;
    if (FAILED(pSessionEnum->GetCount(&count))) {
        pSessionEnum->Release();
        return false;
    }

    std::string appLower = ToLowerStr(sourceApp);
    auto tokens = ExtractAppTokens(sourceApp);
    std::vector<SessionMatch> matches;
    SessionMatch bestFallback;

    for (int i = 0; i < count; ++i) {
        IAudioSessionControl* pControl = nullptr;
        if (FAILED(pSessionEnum->GetSession(i, &pControl)) || !pControl)
            continue;

        IAudioSessionControl2* pControl2 = nullptr;
        if (SUCCEEDED(
                pControl->QueryInterface(__uuidof(IAudioSessionControl2), reinterpret_cast<void**>(&pControl2))) &&
            pControl2) {
            AudioSessionState state = AudioSessionStateInactive;
            pControl2->GetState(&state);
            if (pControl2->IsSystemSoundsSession() != S_OK && state != AudioSessionStateExpired) {
                bool isActive = (state == AudioSessionStateActive);
                float peak = 0.0f;
                IAudioMeterInformation* pMeter = nullptr;
                if (SUCCEEDED(pControl->QueryInterface(__uuidof(IAudioMeterInformation),
                                                       reinterpret_cast<void**>(&pMeter))) &&
                    pMeter) {
                    pMeter->GetPeakValue(&peak);
                    pMeter->Release();
                }

                LPWSTR pswSessId = nullptr, pswDispName = nullptr;
                std::string sessId, dispName;
                if (SUCCEEDED(pControl2->GetSessionIdentifier(&pswSessId)) && pswSessId) {
                    sessId = ToLowerStr(Utf16ToUtf8(pswSessId));
                    CoTaskMemFree(pswSessId);
                }
                if (SUCCEEDED(pControl2->GetDisplayName(&pswDispName)) && pswDispName) {
                    dispName = ToLowerStr(Utf16ToUtf8(pswDispName));
                    CoTaskMemFree(pswDispName);
                }

                DWORD pid = 0;
                pControl2->GetProcessId(&pid);
                std::string procPath = ToLowerStr(GetCachedProcessPath(pid));
                std::string procName = procPath;
                if (size_t slash = procName.find_last_of("\\/"); slash != std::string::npos)
                    procName = procName.substr(slash + 1);
                if (size_t pExe = procName.find(".exe"); pExe != std::string::npos)
                    procName.resize(pExe);

                int appScore = 0;
                if (!appLower.empty()) {
                    if (!procName.empty() &&
                        (appLower.find(procName) != std::string::npos || procName.find(appLower) != std::string::npos))
                        appScore += 100;
                    if (!sessId.empty() && sessId.find(appLower) != std::string::npos)
                        appScore += 80;
                    if (!dispName.empty() &&
                        (appLower.find(dispName) != std::string::npos || dispName.find(appLower) != std::string::npos))
                        appScore += 70;
                }
                for (const auto& token : tokens) {
                    if (token.empty())
                        continue;
                    if (!procName.empty() && procName.find(token) != std::string::npos)
                        appScore += 60;
                    if (!procPath.empty() && procPath.find(token) != std::string::npos)
                        appScore += 40;
                    if (!sessId.empty() && sessId.find(token) != std::string::npos)
                        appScore += 40;
                    if (!dispName.empty() && dispName.find(token) != std::string::npos)
                        appScore += 30;
                }

                int score = appScore + (isActive ? 40 : 0) + (peak > 0.001f ? 20 : 0);
                bool isMatch = (!appLower.empty() || !tokens.empty()) ? (appScore > 0) : (score >= 50);

                ISimpleAudioVolume* pVol = nullptr;
                pControl2->QueryInterface(__uuidof(ISimpleAudioVolume), reinterpret_cast<void**>(&pVol));
                if (pVol) {
                    if (isMatch) {
                        matches.push_back({pVol, score, peak, isActive});
                    } else {
                        int fbScore = (isActive ? 20 : 0) + (peak > 0.001f ? 50 : 0);
                        if (fbScore > bestFallback.score || (!bestFallback.pVol && isActive)) {
                            if (bestFallback.pVol)
                                bestFallback.pVol->Release();
                            bestFallback = {pVol, fbScore, peak, isActive};
                        } else {
                            pVol->Release();
                        }
                    }
                }
            }
            pControl2->Release();
        }
        pControl->Release();
    }
    pSessionEnum->Release();

    if (!matches.empty()) {
        if (bestFallback.pVol)
            bestFallback.pVol->Release();
        std::stable_sort(matches.begin(), matches.end(), [](const SessionMatch& a, const SessionMatch& b) {
            return (a.score != b.score) ? (a.score > b.score) : (a.peak > b.peak);
        });
        for (const auto& m : matches)
            outVolumes.push_back(m.pVol);
        if (ppPrimaryVolume)
            *ppPrimaryVolume = outVolumes[0];
        return true;
    }

    if (bestFallback.pVol) {
        if (appLower.empty()) {
            outVolumes.push_back(bestFallback.pVol);
            if (ppPrimaryVolume)
                *ppPrimaryVolume = bestFallback.pVol;
            return true;
        }
        bestFallback.pVol->Release();
    }
    return false;
}

static std::pair<float, bool> GetMediaAudioVolumeAndMute(const std::string& sourceApp) {
    auto now = std::chrono::steady_clock::now();
    {
        std::lock_guard<std::mutex> stLock(g_mediaVolStateMutex);
        auto sinceSet = std::chrono::duration_cast<std::chrono::milliseconds>(now - g_lastVolumeSetTime).count();
        auto sinceRead = std::chrono::duration_cast<std::chrono::milliseconds>(now - g_lastVolumeReadTime).count();
        if (sourceApp == g_mediaVolCachedApp && (sinceSet < 700 || sinceRead < 1500)) {
            return {g_mediaVolumeStore, g_mediaMuteStore};
        }
    }

    std::lock_guard<std::mutex> lock(g_audioVolMutex);
    ComInitScope com;

    std::vector<ISimpleAudioVolume*> volumes;
    ISimpleAudioVolume* pPrimary = nullptr;

    if (GetAllAppSessionVolumes(sourceApp, volumes, &pPrimary) && pPrimary) {
        float vol = 1.0f;
        BOOL muted = FALSE;
        pPrimary->GetMasterVolume(&vol);
        pPrimary->GetMute(&muted);
        for (auto* v : volumes)
            if (v)
                v->Release();

        float volPercent = vol * 100.0f;
        std::lock_guard<std::mutex> stLock(g_mediaVolStateMutex);
        g_mediaVolumeStore = volPercent;
        g_mediaMuteStore = (muted == TRUE);
        g_mediaVolCachedApp = sourceApp;
        g_lastVolumeReadTime = now;
        return {volPercent, (muted == TRUE)};
    }

    std::lock_guard<std::mutex> stLock(g_mediaVolStateMutex);
    return {g_mediaVolumeStore, g_mediaMuteStore};
}

static void SetMediaAudioVolume(const std::string& sourceApp, float percent) {
    float clamped = std::max(0.0f, std::min(100.0f, percent));
    {
        std::lock_guard<std::mutex> stLock(g_mediaVolStateMutex);
        g_mediaVolumeStore = clamped;
        g_mediaVolCachedApp = sourceApp;
        g_lastVolumeSetTime = std::chrono::steady_clock::now();
        g_lastVolumeReadTime = g_lastVolumeSetTime;
    }

    std::lock_guard<std::mutex> lock(g_audioVolMutex);
    ComInitScope com;

    std::vector<ISimpleAudioVolume*> volumes;
    if (GetAllAppSessionVolumes(sourceApp, volumes, nullptr)) {
        float scalar = clamped / 100.0f;
        for (auto* pVol : volumes) {
            if (pVol) {
                pVol->SetMasterVolume(scalar, nullptr);
                pVol->Release();
            }
        }
    }
}

static void ToggleMediaAudioMute(const std::string& sourceApp) {
    bool nextMute = false;
    {
        std::lock_guard<std::mutex> stLock(g_mediaVolStateMutex);
        g_mediaMuteStore = !g_mediaMuteStore;
        nextMute = g_mediaMuteStore;
        g_mediaVolCachedApp = sourceApp;
        g_lastVolumeSetTime = std::chrono::steady_clock::now();
    }

    std::lock_guard<std::mutex> lock(g_audioVolMutex);
    ComInitScope com;

    std::vector<ISimpleAudioVolume*> volumes;
    if (GetAllAppSessionVolumes(sourceApp, volumes, nullptr)) {
        BOOL bMute = nextMute ? TRUE : FALSE;
        for (auto* pVol : volumes) {
            if (pVol) {
                pVol->SetMute(bMute, nullptr);
                pVol->Release();
            }
        }
    }
}

void StartMediaService(NativeMediaCallback cb) {
    if (g_mediaRunning.exchange(true, std::memory_order_acq_rel))
        return;
    if (g_mediaSessionPollThread.joinable())
        g_mediaSessionPollThread.join();
    g_mediaCb.store(cb, std::memory_order_release);

    g_mediaSessionPollThread = std::thread([]() {
        ApartmentScope apt;
        try {
            g_sessionManager = winrt_media::GlobalSystemMediaTransportControlsSessionManager::RequestAsync().get();
            if (g_sessionManager) {
                g_currentSessionChangedToken =
                    g_sessionManager.CurrentSessionChanged([](auto const&, auto const&) { ScheduleEmit(); });
                g_sessionsChangedToken = g_sessionManager.SessionsChanged([](auto const&, auto const&) {
                    {
                        std::lock_guard<std::mutex> sLock(g_sessionsJsonMutex);
                        g_lastSessionsFetchTime = {};
                    }
                    ScheduleEmit();
                });
            }
        } catch (...) {
        }

        UpdateCurrentSession();

        while (g_mediaRunning.load(std::memory_order_acquire)) {
            if (!SleepInterruptible(3000))
                break;
            if (g_mediaRunning.load(std::memory_order_acquire)) {
                UpdateCurrentSession();
            }
        }
    });
}

void StopMediaService() {
    {
        std::lock_guard<std::mutex> stopLock(g_mediaStopMutex);
        if (!g_mediaRunning.exchange(false, std::memory_order_acq_rel))
            return;
    }
    g_mediaCb.store(nullptr, std::memory_order_release);
    ++g_artCorrectionGeneration;
    g_mediaStopCv.notify_all();

    if (g_mediaSessionPollThread.joinable())
        g_mediaSessionPollThread.join();
    while (g_activeBgTasks.load(std::memory_order_acquire) > 0) {
        std::this_thread::sleep_for(std::chrono::milliseconds(10));
    }

    std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
    if (g_sessionManager) {
        try {
            g_sessionManager.CurrentSessionChanged(g_currentSessionChangedToken);
        } catch (...) {
        }
        try {
            g_sessionManager.SessionsChanged(g_sessionsChangedToken);
        } catch (...) {
        }
    }
    if (g_currentSession) {
        try {
            g_currentSession.MediaPropertiesChanged(g_mediaPropsChangedToken);
        } catch (...) {
        }
        try {
            g_currentSession.PlaybackInfoChanged(g_playbackInfoChangedToken);
        } catch (...) {
        }
        try {
            g_currentSession.TimelinePropertiesChanged(g_timelinePropsChangedToken);
        } catch (...) {
        }
        g_currentSession = nullptr;
    }
    g_sessionManager = nullptr;
}

int ExecuteMediaCommand(const char* command) {
    if (!command)
        return 0;
    ApartmentScope apt;
    std::string cmd(command);

    std::string sourceApp;
    {
        std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
        if (g_currentSession) {
            try {
                sourceApp = Utf16ToUtf8(g_currentSession.SourceAppUserModelId().c_str());
            } catch (...) {
            }
        }
    }

    if (cmd.rfind("set_volume:", 0) == 0) {
        std::string payload = cmd.substr(11);
        size_t colon = payload.find(':');
        float vol = std::stof(colon != std::string::npos ? payload.substr(0, colon) : payload);
        if (colon != std::string::npos && colon + 1 < payload.size())
            sourceApp = payload.substr(colon + 1);
        SetMediaAudioVolume(sourceApp, vol);
        EmitMediaState(true);
        return 1;
    }
    if (cmd == "volume_up" || cmd == "volume_down") {
        auto [curVol, _] = GetMediaAudioVolumeAndMute(sourceApp);
        SetMediaAudioVolume(sourceApp, curVol + (cmd == "volume_up" ? 5.0f : -5.0f));
        EmitMediaState(true);
        return 1;
    }
    if (cmd.rfind("toggle_mute:", 0) == 0) {
        if (cmd.size() > 12)
            sourceApp = cmd.substr(12);
        ToggleMediaAudioMute(sourceApp);
        EmitMediaState(true);
        return 1;
    }
    if (cmd == "mute" || cmd == "toggle_mute") {
        ToggleMediaAudioMute(sourceApp);
        EmitMediaState(true);
        return 1;
    }
    if (cmd.rfind("seek:", 0) == 0) {
        std::string payload = cmd.substr(5);
        size_t colon = payload.find(':');
        double posMs = std::stod(colon != std::string::npos ? payload.substr(0, colon) : payload);
        std::string explicitTarget = (colon != std::string::npos) ? payload.substr(colon + 1) : "";

        winrt_media::GlobalSystemMediaTransportControlsSession session{nullptr};
        winrt_media::GlobalSystemMediaTransportControlsSessionManager mgr{nullptr};
        if (!explicitTarget.empty()) {
            std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
            mgr = g_sessionManager;
        }
        if (!explicitTarget.empty() && mgr) {
            try {
                for (auto const& s : mgr.GetSessions()) {
                    if (!s)
                        continue;
                    try {
                        std::string sId = Utf16ToUtf8(s.SourceAppUserModelId().c_str());
                        std::string sTitle, sArtist;
                        if (auto props = s.TryGetMediaPropertiesAsync().get()) {
                            sTitle = Utf16ToUtf8(props.Title().c_str());
                            sArtist = Utf16ToUtf8(props.Artist().c_str());
                            TrimString(sTitle);
                            TrimString(sArtist);
                        }
                        std::string uniqueId = ComputeSessionUniqueId(sId, sTitle, sArtist);
                        if (explicitTarget == uniqueId || explicitTarget == sId ||
                            (!sTitle.empty() && explicitTarget.find(sTitle) != std::string::npos)) {
                            session = s;
                            break;
                        }
                    } catch (...) {
                    }
                }
            } catch (...) {
            }
        }
        if (!session) {
            std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
            session = g_currentSession;
        }

        if (session) {
            // Lock in the optimistic seek position before calling into WinRT so concurrent ticks never emit the old
            // position.
            auto now = std::chrono::steady_clock::now();
            uint64_t seekGen = g_seekGeneration.fetch_add(1, std::memory_order_acq_rel) + 1;
            {
                std::lock_guard<std::mutex> skLock(g_seekMutex);
                g_seekTargetMs = posMs;
                g_seekTime = now;
            }
            {
                std::lock_guard<std::mutex> snapLock(g_validSnapshotMutex);
                if (g_lastValidTrackSnapshot.hasValue) {
                    g_lastValidTrackSnapshot.posMs = posMs;
                    g_lastValidTrackSnapshot.timestamp = now;
                }
            }

            try {
                if (auto asyncOp = session.TryChangePlaybackPositionAsync(static_cast<int64_t>(posMs * 10000.0))) {
                    asyncOp.get();
                }
            } catch (...) {
            }

            EmitMediaState(true);

            g_activeBgTasks.fetch_add(1, std::memory_order_acq_rel);
            std::thread([session, posMs, seekGen]() {
                ApartmentScope apt;
                for (int i = 0; i < 6; ++i) {
                    if (!SleepInterruptible(80) || g_seekGeneration.load(std::memory_order_acquire) != seekGen) {
                        g_activeBgTasks.fetch_sub(1, std::memory_order_acq_rel);
                        return;
                    }
                    try {
                        if (auto tl = session.GetTimelineProperties()) {
                            double cur = std::chrono::duration<double, std::milli>(tl.Position()).count();
                            if (std::fabs(cur - posMs) < 1500.0)
                                break;
                        }
                    } catch (...) {
                    }
                }
                if (g_seekGeneration.load(std::memory_order_acquire) == seekGen) {
                    UpdateCurrentSession();
                    if (SleepInterruptible(400) && g_seekGeneration.load(std::memory_order_acquire) == seekGen) {
                        UpdateCurrentSession();
                    }
                }
                g_activeBgTasks.fetch_sub(1, std::memory_order_acq_rel);
            }).detach();
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
            if (cmd == "play_pause")
                session.TryTogglePlayPauseAsync();
            else if (cmd == "play")
                session.TryPlayAsync();
            else if (cmd == "pause")
                session.TryPauseAsync();
            else if (cmd == "next")
                session.TrySkipNextAsync();
            else if (cmd == "previous" || cmd == "prev")
                session.TrySkipPreviousAsync();
            else if (cmd == "shuffle" || cmd == "toggle_shuffle") {
                auto info = session.GetPlaybackInfo();
                bool cur = (info && info.IsShuffleActive()) ? info.IsShuffleActive().Value() : false;
                session.TryChangeShuffleActiveAsync(!cur);
            } else if (cmd == "repeat" || cmd == "toggle_repeat") {
                auto info = session.GetPlaybackInfo();
                auto target = winrt::Windows::Media::MediaPlaybackAutoRepeatMode::Track;
                if (info && info.AutoRepeatMode()) {
                    auto cur = info.AutoRepeatMode().Value();
                    if (cur == winrt::Windows::Media::MediaPlaybackAutoRepeatMode::Track) {
                        target = winrt::Windows::Media::MediaPlaybackAutoRepeatMode::List;
                    } else if (cur == winrt::Windows::Media::MediaPlaybackAutoRepeatMode::List) {
                        target = winrt::Windows::Media::MediaPlaybackAutoRepeatMode::None;
                    }
                }
                session.TryChangeAutoRepeatModeAsync(target);
            }
        } catch (...) {
        }

        g_activeBgTasks.fetch_add(1, std::memory_order_acq_rel);
        std::thread([]() {
            ApartmentScope apt;
            const int delays[] = {40, 140, 320};
            for (int d : delays) {
                if (!SleepInterruptible(d))
                    break;
                UpdateCurrentSession();
            }
            g_activeBgTasks.fetch_sub(1, std::memory_order_acq_rel);
        }).detach();
    }

    return 1;
}

int GetAllAudioSessionsJson(char* outBuf, int maxLen) {
    if (!outBuf || maxLen <= 0)
        return 0;
    ComInitScope com;

    IAudioSessionEnumerator* pSessionEnum = AcquireAudioSessionEnumerator();
    if (!pSessionEnum) {
        std::snprintf(outBuf, maxLen, "[]");
        return 0;
    }

    int count = 0;
    if (FAILED(pSessionEnum->GetCount(&count))) {
        pSessionEnum->Release();
        std::snprintf(outBuf, maxLen, "[]");
        return 0;
    }

    std::ostringstream json;
    json << "[";
    bool first = true;

    for (int i = 0; i < count; ++i) {
        IAudioSessionControl* pControl = nullptr;
        if (FAILED(pSessionEnum->GetSession(i, &pControl)) || !pControl)
            continue;

        IAudioSessionControl2* pControl2 = nullptr;
        if (SUCCEEDED(
                pControl->QueryInterface(__uuidof(IAudioSessionControl2), reinterpret_cast<void**>(&pControl2))) &&
            pControl2) {
            DWORD pid = 0;
            pControl2->GetProcessId(&pid);
            bool isSystem = (pControl2->IsSystemSoundsSession() == S_OK);

            if (isSystem || pid != 0) {
                AudioSessionState state = AudioSessionStateInactive;
                pControl2->GetState(&state);
                bool isActive = (state == AudioSessionStateActive);

                float volume = 100.0f;
                bool isMuted = false;
                ISimpleAudioVolume* pVol = nullptr;
                if (SUCCEEDED(
                        pControl2->QueryInterface(__uuidof(ISimpleAudioVolume), reinterpret_cast<void**>(&pVol))) &&
                    pVol) {
                    float v = 1.0f;
                    BOOL m = FALSE;
                    pVol->GetMasterVolume(&v);
                    pVol->GetMute(&m);
                    volume = v * 100.0f;
                    isMuted = (m == TRUE);
                    pVol->Release();
                }

                float peak = 0.0f;
                IAudioMeterInformation* pMeter = nullptr;
                if (SUCCEEDED(pControl->QueryInterface(__uuidof(IAudioMeterInformation),
                                                       reinterpret_cast<void**>(&pMeter))) &&
                    pMeter) {
                    pMeter->GetPeakValue(&peak);
                    pMeter->Release();
                }

                std::string procPath = (pid != 0) ? GetCachedProcessPath(pid) : "";
                std::string procName = procPath;
                if (size_t slash = procName.find_last_of("\\/"); slash != std::string::npos) {
                    procName = procName.substr(slash + 1);
                }
                if (procName.empty()) {
                    procName = isSystem ? "System Sounds" : ("App (" + std::to_string(pid) + ")");
                }

                LPWSTR pswDisp = nullptr;
                std::string dispName;
                if (SUCCEEDED(pControl2->GetDisplayName(&pswDisp)) && pswDisp) {
                    dispName = Utf16ToUtf8(pswDisp);
                    CoTaskMemFree(pswDisp);
                }
                if (dispName.empty() || dispName[0] == '@')
                    dispName = procName;

                LPWSTR pswInst = nullptr;
                std::string instanceId;
                if (SUCCEEDED(pControl2->GetSessionInstanceIdentifier(&pswInst)) && pswInst) {
                    instanceId = Utf16ToUtf8(pswInst);
                    CoTaskMemFree(pswInst);
                }
                if (instanceId.empty())
                    instanceId = std::to_string(pid) + "-" + std::to_string(i);

                if (!first)
                    json << ",";
                first = false;
                json << "{\"id\":\"" << EscapeJsonString(instanceId) << "\",\"pid\":" << pid << ",\"name\":\""
                     << EscapeJsonString(dispName) << "\",\"processName\":\"" << EscapeJsonString(procName)
                     << "\",\"path\":\"" << EscapeJsonString(procPath) << "\",\"volume\":" << std::fixed
                     << std::setprecision(1) << volume << ",\"isMuted\":" << (isMuted ? "true" : "false")
                     << ",\"peak\":" << std::fixed << std::setprecision(3) << peak
                     << ",\"isActive\":" << (isActive ? "true" : "false") << "}";
            }
            pControl2->Release();
        }
        pControl->Release();
    }

    pSessionEnum->Release();
    json << "]";

    std::string out = json.str();
    if (out.length() >= static_cast<size_t>(maxLen))
        out.resize(maxLen - 1);
    strncpy_s(outBuf, maxLen, out.c_str(), _TRUNCATE);
    return static_cast<int>(out.length());
}

/**
 * @brief Shared helper that applies a volume or mute change to all audio sessions matching pid.
 */
static int ModifyAudioSessionByPID(uint32_t pid, bool setVolume, int volumePercent, bool mute) {
    ComInitScope com;
    IAudioSessionEnumerator* pSessionEnum = AcquireAudioSessionEnumerator();
    if (!pSessionEnum)
        return 0;

    int count = 0;
    if (FAILED(pSessionEnum->GetCount(&count))) {
        pSessionEnum->Release();
        return 0;
    }

    float clamped = std::max(0.0f, std::min(1.0f, static_cast<float>(volumePercent) / 100.0f));
    int applied = 0;

    for (int i = 0; i < count; ++i) {
        IAudioSessionControl* pControl = nullptr;
        if (FAILED(pSessionEnum->GetSession(i, &pControl)) || !pControl)
            continue;

        IAudioSessionControl2* pControl2 = nullptr;
        if (SUCCEEDED(
                pControl->QueryInterface(__uuidof(IAudioSessionControl2), reinterpret_cast<void**>(&pControl2))) &&
            pControl2) {
            DWORD sPid = 0;
            pControl2->GetProcessId(&sPid);
            if (sPid == pid || (pid == 0 && pControl2->IsSystemSoundsSession() == S_OK)) {
                ISimpleAudioVolume* pVol = nullptr;
                if (SUCCEEDED(
                        pControl2->QueryInterface(__uuidof(ISimpleAudioVolume), reinterpret_cast<void**>(&pVol))) &&
                    pVol) {
                    if (setVolume)
                        pVol->SetMasterVolume(clamped, nullptr);
                    else
                        pVol->SetMute(mute ? TRUE : FALSE, nullptr);
                    pVol->Release();
                    applied++;
                }
            }
            pControl2->Release();
        }
        pControl->Release();
    }

    pSessionEnum->Release();
    return applied;
}

int SetAppAudioVolumeByPID(uint32_t pid, int volumePercent) {
    return ModifyAudioSessionByPID(pid, true, volumePercent, false);
}

int SetAppAudioMuteByPID(uint32_t pid, bool mute) {
    return ModifyAudioSessionByPID(pid, false, 0, mute);
}

int GetAudioEndpointsJson(char* outBuf, int maxLen) {
    if (!outBuf || maxLen <= 0)
        return 0;
    ComInitScope com;

    IMMDeviceEnumerator* pEnumerator = nullptr;
    if (FAILED(CoCreateInstance(__uuidof(MMDeviceEnumerator), nullptr, CLSCTX_INPROC_SERVER,
                                __uuidof(IMMDeviceEnumerator), reinterpret_cast<void**>(&pEnumerator))) ||
        !pEnumerator) {
        std::snprintf(outBuf, maxLen, "[]");
        return 0;
    }

    std::wstring defaultId;
    IMMDevice* pDefDev = nullptr;
    if (SUCCEEDED(pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDefDev)) && pDefDev) {
        LPWSTR pStr = nullptr;
        if (SUCCEEDED(pDefDev->GetId(&pStr)) && pStr) {
            defaultId = pStr;
            CoTaskMemFree(pStr);
        }
        pDefDev->Release();
    }

    IMMDeviceCollection* pCollection = nullptr;
    if (FAILED(pEnumerator->EnumAudioEndpoints(eRender, DEVICE_STATE_ACTIVE, &pCollection)) || !pCollection) {
        pEnumerator->Release();
        std::snprintf(outBuf, maxLen, "[]");
        return 0;
    }

    UINT count = 0;
    pCollection->GetCount(&count);

    std::ostringstream json;
    json << "[";
    bool first = true;

    for (UINT i = 0; i < count; ++i) {
        IMMDevice* pEndpoint = nullptr;
        if (SUCCEEDED(pCollection->Item(i, &pEndpoint)) && pEndpoint) {
            LPWSTR pId = nullptr;
            std::string idStr;
            bool isDef = false;
            if (SUCCEEDED(pEndpoint->GetId(&pId)) && pId) {
                isDef = (defaultId == pId);
                idStr = Utf16ToUtf8(pId);
                CoTaskMemFree(pId);
            }

            std::string friendlyName;
            IPropertyStore* pProps = nullptr;
            if (SUCCEEDED(pEndpoint->OpenPropertyStore(STGM_READ, &pProps)) && pProps) {
                PROPVARIANT varName;
                PropVariantInit(&varName);
                if (SUCCEEDED(pProps->GetValue(PKEY_Device_FriendlyName, &varName)) && varName.vt == VT_LPWSTR &&
                    varName.pwszVal) {
                    friendlyName = Utf16ToUtf8(varName.pwszVal);
                }
                PropVariantClear(&varName);
                pProps->Release();
            }
            if (friendlyName.empty())
                friendlyName = "Audio Output Device";

            if (!first)
                json << ",";
            first = false;
            json << "{\"id\":\"" << EscapeJsonString(idStr) << "\",\"name\":\"" << EscapeJsonString(friendlyName)
                 << "\",\"isDefault\":" << (isDef ? "true" : "false") << ",\"flow\":\"render\"}";
            pEndpoint->Release();
        }
    }

    pCollection->Release();
    pEnumerator->Release();
    json << "]";

    std::string out = json.str();
    if (out.length() >= static_cast<size_t>(maxLen))
        out.resize(maxLen - 1);
    strncpy_s(outBuf, maxLen, out.c_str(), _TRUNCATE);
    return static_cast<int>(out.length());
}

int SetDefaultAudioEndpoint(const wchar_t* endpointId) {
    if (!endpointId || !*endpointId)
        return 0;
    ComInitScope com;
    IPolicyConfig* pPolicyConfig = nullptr;
    if (SUCCEEDED(CoCreateInstance(__uuidof(CPolicyConfigClient), nullptr, CLSCTX_ALL, __uuidof(IPolicyConfig),
                                   reinterpret_cast<void**>(&pPolicyConfig))) &&
        pPolicyConfig) {
        pPolicyConfig->SetDefaultEndpoint(endpointId, eConsole);
        pPolicyConfig->SetDefaultEndpoint(endpointId, eMultimedia);
        pPolicyConfig->SetDefaultEndpoint(endpointId, eCommunications);
        pPolicyConfig->Release();
        return 1;
    }
    return 0;
}

int SetMasterVolumeNative(int percent) {
    ComInitScope com;
    if (IAudioEndpointVolume* pEndVol = AcquireMasterEndpointVolume()) {
        float clamped = std::max(0.0f, std::min(1.0f, static_cast<float>(percent) / 100.0f));
        pEndVol->SetMasterVolumeLevelScalar(clamped, nullptr);
        pEndVol->Release();
        return 1;
    }
    return 0;
}

int ToggleMasterMuteNative() {
    ComInitScope com;
    if (IAudioEndpointVolume* pEndVol = AcquireMasterEndpointVolume()) {
        BOOL muted = FALSE;
        pEndVol->GetMute(&muted);
        pEndVol->SetMute(!muted, nullptr);
        pEndVol->Release();
        return 1;
    }
    return 0;
}