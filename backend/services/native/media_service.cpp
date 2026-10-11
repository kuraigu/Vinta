#include "media_service.h"
#include "winrt_common.h"

#include <mmdeviceapi.h>
#include <endpointvolume.h>
#include <audiopolicy.h>
#include <psapi.h>
#include <shellapi.h>
#include <winrt/Windows.Media.Control.h>
#include <algorithm>
#include <unordered_map>
#include <unordered_set>

#include <atomic>
#include <mutex>
#include <thread>

namespace winrt_media = winrt::Windows::Media::Control;

static NativeMediaCallback g_mediaCb = nullptr;
static std::atomic<bool> g_mediaRunning{false};
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

static std::unordered_map<std::string, std::string> g_trackArtMap;
static std::vector<std::string> g_trackArtOrder;
static std::string g_currentTrackKey;
static std::string g_previousTrackArt;
static std::chrono::steady_clock::time_point g_trackStartTime{};
static std::mutex g_artCacheMutex;

static std::mutex g_audioVolMutex;

static std::string g_selectedSessionId;
static std::chrono::steady_clock::time_point g_selectedSessionTime;
static std::mutex g_selectedSessionMutex;

static std::atomic<uint64_t> g_artCorrectionGeneration{0};
static std::atomic<uint64_t> g_activeWorkerGen{0};
static std::chrono::steady_clock::time_point g_lastArtResyncTime{};
static std::mutex g_artResyncMutex;
static std::atomic<bool> g_mediaPropsDirty{false};

static void UpdateCurrentSession();
static void EmitMediaState(bool includeVolume = true);
static void ScheduleEmit();

static void TrimString(std::string& s);

/**
 * Constructs a deterministic identifier for an audio/media playback session.
 * Binding the source application ID with track title and artist prevents
 * state collisions and loop-index drift when sessions dynamically change.
 */
static std::string ComputeSessionUniqueId(const std::string& sourceApp, const std::string& title, const std::string& artist) {
    std::string t = title;
    std::string a = artist;
    TrimString(t);
    TrimString(a);
    return sourceApp + "::" + t + "::" + a;
}

struct ValidTrackSnapshot {
    std::string sourceApp;
    std::string title;
    std::string artist;
    std::string album;
    std::string albumArt;
    double posMs{0.0};
    double durMs{0.0};
    bool isPlaying{false};
    bool canGoNext{true};
    bool canGoPrev{true};
    bool canPause{true};
    bool canPlay{true};
    bool isShuffleActive{false};
    int autoRepeatMode{0};
    bool isShuffleEnabled{true};
    bool isRepeatEnabled{true};
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
        g_lastValidTrackSnapshot.hasValue = false;
        g_lastValidTrackSnapshot.title.clear();
        g_lastValidTrackSnapshot.artist.clear();
        g_lastValidTrackSnapshot.album.clear();
        g_lastValidTrackSnapshot.albumArt.clear();
        g_lastValidTrackSnapshot.sourceApp.clear();
        g_lastValidTrackSnapshot.posMs = 0.0;
        g_lastValidTrackSnapshot.durMs = 0.0;
        g_lastValidTrackSnapshot.isPlaying = false;
    }
}

/**
 * Re-queries the media properties and thumbnail stream for the current session.
 * Used when an actively playing media track initially had an empty or loading artwork stream,
 * or when the user or frontend requests a refreshed thumbnail from Windows SMTC.
 */
void RequeryMediaArtwork() {
    g_mediaPropsDirty.store(true);
    {
        std::lock_guard<std::mutex> artLock(g_artCacheMutex);
        if (!g_currentTrackKey.empty()) {
            g_trackArtMap.erase(g_currentTrackKey);
            g_trackArtOrder.erase(std::remove(g_trackArtOrder.begin(), g_trackArtOrder.end(), g_currentTrackKey), g_trackArtOrder.end());
        }
        g_previousTrackArt.clear();
    }
    ++g_artCorrectionGeneration;
    ScheduleEmit();
}

/**
 * Forces an immediate emission of the active media state to the IPC callback.
 * Important during application startup and frontend reconnections, ensuring
 * WebView2 receives the latest playback properties without waiting for the next tick.
 */
void ForceEmitMediaState() {
    UpdateCurrentSession();
}

static std::string g_cachedSessionsJson = "[]";
static std::chrono::steady_clock::time_point g_lastSessionsFetchTime{};
static std::mutex g_sessionsJsonMutex;

static std::string TryExtractThumbnail(winrt_media::GlobalSystemMediaTransportControlsSessionMediaProperties const& props, const std::string& staleArt = "", bool allowStale = false) {
    if (!props) return "";
    try {
        auto thumbRef = props.Thumbnail();
        if (thumbRef) {
            winrt_storage::IRandomAccessStreamWithContentType stream{nullptr};
            try {
                stream = thumbRef.OpenReadAsync().get();
            } catch (...) {
                // Transient sharing violation or stream initialization retry
                std::this_thread::sleep_for(std::chrono::milliseconds(50));
                try {
                    stream = thumbRef.OpenReadAsync().get();
                } catch (...) {}
            }
            if (stream) {
                std::string fetched = StreamToBase64(stream);
                if (!fetched.empty()) {
                    if (allowStale || staleArt.empty() || fetched != staleArt) {
                        return fetched;
                    }
                }
            }
        }
    } catch (...) {}
    return "";
}

static std::pair<float, bool> GetMasterAudioVolumeAndMute();
static void SetMasterAudioVolume(float percent);
static void ToggleMasterAudioMute();
static std::pair<float, bool> GetMediaAudioVolumeAndMute(const std::string& sourceApp);
static void SetMediaAudioVolume(const std::string& sourceApp, float percent);
static void ToggleMediaAudioMute(const std::string& sourceApp);
static void SendVolumeKey(BYTE vk);

bool IsMediaPlayingState() {
    return g_isMediaPlaying.load();
}

int SelectMediaSession(const char* sessionId) {
    {
        std::lock_guard<std::mutex> lock(g_selectedSessionMutex);
        g_selectedSessionId = (sessionId != nullptr) ? sessionId : "";
        g_selectedSessionTime = std::chrono::steady_clock::now();
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
    size_t end = s.find_last_not_of(" \t\r\n");
    s = s.substr(start, end - start + 1);
}


int FocusApp(const char* sourceApp, const char* aumid) {
    std::string needle = (sourceApp && strlen(sourceApp) > 0) ? ToLowerStr(sourceApp) : "";
    std::string rawAumid = (aumid && strlen(aumid) > 0) ? aumid : "";
    std::string needleAumid = ToLowerStr(rawAumid);

    size_t lastSlash = needle.find_last_of("\\/");
    if (lastSlash != std::string::npos) needle = needle.substr(lastSlash + 1);
    if (needle.size() > 4 && needle.substr(needle.size() - 4) == ".exe") {
        needle = needle.substr(0, needle.size() - 4);
    }

    std::string aumidToken = needleAumid;
    size_t lastDot = aumidToken.find_last_of(".");
    if (lastDot != std::string::npos && lastDot + 1 < aumidToken.size()) {
        aumidToken = aumidToken.substr(lastDot + 1);
    }
    size_t bang = aumidToken.find("!");
    if (bang != std::string::npos) {
        aumidToken = aumidToken.substr(0, bang);
    }

    WindowFocusData data;
    data.needle = needle;
    data.needleAumid = aumidToken;

    EnumWindows([](HWND hwnd, LPARAM lParam) -> BOOL {
        if (!IsWindowVisible(hwnd)) return TRUE;
        LONG exStyle = GetWindowLongW(hwnd, GWL_EXSTYLE);
        if (exStyle & WS_EX_TOOLWINDOW) return TRUE;

        DWORD pid = 0;
        GetWindowThreadProcessId(hwnd, &pid);
        if (pid == 0) return TRUE;

        HANDLE hProc = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, FALSE, pid);
        if (hProc) {
            WCHAR path[MAX_PATH] = {0};
            DWORD size = MAX_PATH;
            if (QueryFullProcessImageNameW(hProc, 0, path, &size)) {
                std::string exePath = ToLowerStr(Utf16ToUtf8(path));
                auto* d = reinterpret_cast<WindowFocusData*>(lParam);
                bool match = false;
                if (!d->needle.empty() && exePath.find(d->needle) != std::string::npos) {
                    match = true;
                }
                if (!match && !d->needleAumid.empty() && exePath.find(d->needleAumid) != std::string::npos) {
                    match = true;
                }
                if (!match && d->needle.find("spotify") != std::string::npos && exePath.find("spotify") != std::string::npos) {
                    match = true;
                }
                if (match) {
                    int len = GetWindowTextLengthW(hwnd);
                    if (len > 0) {
                        d->foundHwnd = hwnd;
                        CloseHandle(hProc);
                        return FALSE;
                    }
                }
            }
            CloseHandle(hProc);
        }
        return TRUE;
    }, reinterpret_cast<LPARAM>(&data));

    if (data.foundHwnd) {
        if (IsIconic(data.foundHwnd)) {
            ShowWindow(data.foundHwnd, SW_RESTORE);
        }
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
        HINSTANCE res = ShellExecuteW(nullptr, L"open", target.c_str(), nullptr, nullptr, SW_SHOWNORMAL);
        if ((INT_PTR)res > 32) {
            return 1;
        }
    }

    return 0;
}

int FocusMediaApp(const char* sourceApp) {
    return FocusApp(sourceApp, "");
}

static void ScheduleEmit() {
    bool expected = false;
    if (g_emitPending.compare_exchange_strong(expected, true)) {
        std::thread([]() {
            ApartmentScope apt;
            std::this_thread::sleep_for(std::chrono::milliseconds(50));
            g_emitPending.store(false);
            UpdateCurrentSession();
        }).detach();
    }
}

/**
 * Evaluates whether a GlobalSystemMediaTransportControlsSession represents a legitimate media player.
 * Filters out dummy audio sessions created by game launchers or system background processes,
 * while ensuring valid paused, stopped, or web browser media sessions are retained.
 */
static bool IsValidMediaSession(winrt_media::GlobalSystemMediaTransportControlsSession const& session,
                                winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackInfo const* prefetchedInfo = nullptr,
                                winrt_media::GlobalSystemMediaTransportControlsSessionMediaProperties const* prefetchedProps = nullptr) {
    if (!session) return false;
    try {
        winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackInfo info{nullptr};
        if (prefetchedInfo && *prefetchedInfo) {
            info = *prefetchedInfo;
        } else {
            info = session.GetPlaybackInfo();
        }
        if (!info) return false;

        auto status = info.PlaybackStatus();
        if (status == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Closed) {
            return false;
        }

        // Must support at least one interactive media playback control.
        // Web browsers commonly expose IsPlayPauseToggleEnabled rather than discrete play or pause.
        auto controls = info.Controls();
        if (!controls) return false;
        bool hasControls = controls.IsPlayEnabled() || controls.IsPauseEnabled() ||
                           controls.IsPlayPauseToggleEnabled() || controls.IsStopEnabled() ||
                           controls.IsNextEnabled() || controls.IsPreviousEnabled();
        if (!hasControls) return false;

        winrt_media::GlobalSystemMediaTransportControlsSessionMediaProperties props{nullptr};
        if (prefetchedProps && *prefetchedProps) {
            props = *prefetchedProps;
        } else {
            props = session.TryGetMediaPropertiesAsync().get();
        }
        if (!props) return false;

        std::string title = Utf16ToUtf8(props.Title().c_str());
        std::string artist = Utf16ToUtf8(props.Artist().c_str());

        TrimString(title);
        TrimString(artist);

        if (title.empty() && artist.empty()) {
            return false;
        }

        // Exclude dummy sources that lack an artist and lack either album artwork or prev/next controls.
        // Genuine sources (Spotify, YouTube with uploader as artist) always provide an artist.
        bool hasAlbumArt = (props.Thumbnail() != nullptr);
        bool hasArtist = !artist.empty();
        bool hasPrevNext = (controls.IsNextEnabled() || controls.IsPreviousEnabled());
        if (!hasArtist && (!hasAlbumArt || !hasPrevNext)) {
            return false;
        }

        // If artist is empty, check if title is simply the application or process name.
        // Dummy background audio streams typically publish their own executable name as the title with no artist.
        if (artist.empty()) {
            std::string appName;
            try {
                appName = Utf16ToUtf8(session.SourceAppUserModelId().c_str());
            } catch (...) {}
            size_t slash = appName.find_last_of("\\/");
            if (slash != std::string::npos) appName = appName.substr(slash + 1);
            size_t dot = appName.rfind('.');
            if (dot != std::string::npos) appName = appName.substr(0, dot);

            std::string lowerTitle = ToLowerStr(title);
            std::string lowerAppName = ToLowerStr(appName);

            if (!lowerAppName.empty() && lowerTitle == lowerAppName) {
                return false;
            }
        }

        return true;
    } catch (...) {
        return false;
    }
}

/**
 * Checks whether a session is alive and valid.
 * If transient web interactions (such as video elements or ads on X/Twitter) clobber
 * session metadata, falls back to the recent valid track snapshot to prevent dropping active media.
 */
static bool IsSessionAliveAndValid(winrt_media::GlobalSystemMediaTransportControlsSession const& session,
                                  winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackInfo const* prefetchedInfo = nullptr,
                                  winrt_media::GlobalSystemMediaTransportControlsSessionMediaProperties const* prefetchedProps = nullptr) {
    if (!session) return false;
    try {
        winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackInfo info{nullptr};
        if (prefetchedInfo && *prefetchedInfo) {
            info = *prefetchedInfo;
        } else {
            info = session.GetPlaybackInfo();
        }
        if (!info || info.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Closed) {
            return false;
        }

        if (IsValidMediaSession(session, &info, prefetchedProps)) {
            return true;
        }

        std::lock_guard<std::mutex> snapLock(g_validSnapshotMutex);
        if (g_lastValidTrackSnapshot.hasValue) {
            std::string sApp;
            try { sApp = Utf16ToUtf8(session.SourceAppUserModelId().c_str()); } catch (...) {}
            if (sApp == g_lastValidTrackSnapshot.sourceApp) {
                auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(
                    std::chrono::steady_clock::now() - g_lastValidTrackSnapshot.timestamp).count();
                if (elapsed < 15) {
                    return true;
                }
            }
        }
        return false;
    } catch (...) {
        return false;
    }
}

static void UpdateCurrentSession() {
    ApartmentScope apt;
    std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
    if (!g_sessionManager) return;

    try {
        std::string selectedId;
        std::chrono::steady_clock::time_point selectedTime;
        {
            std::lock_guard<std::mutex> sLock(g_selectedSessionMutex);
            selectedId = g_selectedSessionId;
            selectedTime = g_selectedSessionTime;
        }

        winrt_media::GlobalSystemMediaTransportControlsSession targetSession{nullptr};
        winrt_media::GlobalSystemMediaTransportControlsSession chosenSession{nullptr};

        // If user explicitly picked a session, attempt to find it among current sessions
        if (!selectedId.empty()) {
            try {
                auto sessions = g_sessionManager.GetSessions();
                // Pass 1: exact uniqueId or non-empty title match (prevents multi-tab browser collisions)
                for (auto const& s : sessions) {
                    try {
                        std::string sId = Utf16ToUtf8(s.SourceAppUserModelId().c_str());
                        std::string sTitle = "";
                        std::string sArtist = "";
                        try {
                            auto props = s.TryGetMediaPropertiesAsync().get();
                            if (props) {
                                sTitle = Utf16ToUtf8(props.Title().c_str());
                                sArtist = Utf16ToUtf8(props.Artist().c_str());
                                TrimString(sTitle);
                                TrimString(sArtist);
                            }
                        } catch (...) {}

                        std::string uniqueId = ComputeSessionUniqueId(sId, sTitle, sArtist);
                        if (selectedId == uniqueId ||
                            (!sTitle.empty() && selectedId.find(sTitle) != std::string::npos)) {
                            chosenSession = s;
                            break;
                        }
                    } catch (...) {}
                }
                // Pass 2: fallback to app ID match only if no title was available
                if (!chosenSession) {
                    for (auto const& s : sessions) {
                        try {
                            std::string sId = Utf16ToUtf8(s.SourceAppUserModelId().c_str());
                            if (!sId.empty() && selectedId.rfind(sId, 0) == 0) {
                                chosenSession = s;
                                break;
                            }
                        } catch (...) {}
                    }
                }
            } catch (...) {}
        }

        /**
         * Respect user session selection:
         * If the user explicitly picked a session from the session switcher, we honor that choice
         * as long as that session remains alive and valid. We do not arbitrarily eject the user
         * back to an active session after 4 seconds.
         * If the chosen session was closed or terminated, clear the selection and resume auto-detection.
         */
        if (chosenSession && IsSessionAliveAndValid(chosenSession)) {
            targetSession = chosenSession;
        } else {
            if (!selectedId.empty()) {
                std::lock_guard<std::mutex> sLock(g_selectedSessionMutex);
                g_selectedSessionId.clear();
            }

            // Hysteresis Rule 1:
            // If the current session is valid and actively Playing, prefer keeping it.
            // Do NOT let background tabs, ads, or focus shifts on social feeds (e.g. X/Twitter) interrupt active playback.
            bool currentIsPlaying = false;
            bool currentIsStrictlyValid = (g_currentSession && IsValidMediaSession(g_currentSession));
            if (currentIsStrictlyValid) {
                try {
                    auto curInfo = g_currentSession.GetPlaybackInfo();
                    if (curInfo && curInfo.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing) {
                        currentIsPlaying = true;
                    }
                } catch (...) {}
            }

            // Check if Windows's focused session is actively playing and strictly valid on its own merits
            winrt_media::GlobalSystemMediaTransportControlsSession focusedPlayingSession{nullptr};
            try {
                winrt_media::GlobalSystemMediaTransportControlsSession focused = g_sessionManager.GetCurrentSession();
                if (focused && IsValidMediaSession(focused)) {
                    auto info = focused.GetPlaybackInfo();
                    if (info && info.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing) {
                        focusedPlayingSession = focused;
                    }
                }
            } catch (...) {}

            if (currentIsPlaying) {
                // If another session started playing AND is the focused session, switch to it.
                // Otherwise, retain current playing session to avoid tab jitter.
                if (focusedPlayingSession && focusedPlayingSession != g_currentSession) {
                    targetSession = focusedPlayingSession;
                } else {
                    targetSession = g_currentSession;
                }
            } else if (focusedPlayingSession) {
                targetSession = focusedPlayingSession;
            }

            // If still no target session found, scan other sessions for an actively playing session
            if (!targetSession) {
                try {
                    auto sessions = g_sessionManager.GetSessions();
                    for (auto const& s : sessions) {
                        try {
                            if (!IsValidMediaSession(s)) continue;

                            auto info = s.GetPlaybackInfo();
                            if (info && info.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing) {
                                targetSession = s;
                                break;
                            }
                        } catch (...) {}
                    }
                } catch (...) {}
            }

            // Hysteresis Rule 2:
            // When paused (no session is playing), do NOT randomly bounce between background browser tabs.
            // Retain the existing session if it is still alive and valid.
            if (!targetSession) {
                if (g_currentSession && IsSessionAliveAndValid(g_currentSession)) {
                    targetSession = g_currentSession;
                } else {
                    winrt_media::GlobalSystemMediaTransportControlsSession focusedSession{nullptr};
                    try {
                        focusedSession = g_sessionManager.GetCurrentSession();
                    } catch (...) {}

                    if (focusedSession && IsValidMediaSession(focusedSession)) {
                        targetSession = focusedSession;
                    } else {
                        try {
                            auto sessions = g_sessionManager.GetSessions();
                            for (auto const& s : sessions) {
                                if (IsValidMediaSession(s)) {
                                    targetSession = s;
                                    break;
                                }
                            }
                        } catch (...) {}
                    }
                }
            }
        }

        bool sessionChanged = (g_currentSession != targetSession);

        if (sessionChanged) {
            {
                std::lock_guard<std::mutex> sLock(g_sessionsJsonMutex);
                g_lastSessionsFetchTime = {};
            }
            ++g_artCorrectionGeneration;
            g_mediaPropsDirty.store(true);
            if (g_currentSession) {
                try { g_currentSession.MediaPropertiesChanged(g_mediaPropsChangedToken); } catch (...) {}
                try { g_currentSession.PlaybackInfoChanged(g_playbackInfoChangedToken); } catch (...) {}
                try { g_currentSession.TimelinePropertiesChanged(g_timelinePropsChangedToken); } catch (...) {}
            }

            g_currentSession = targetSession;

            if (g_currentSession) {
                try {
                    g_mediaPropsChangedToken = g_currentSession.MediaPropertiesChanged([](auto const&, auto const&) {
                        g_mediaPropsDirty.store(true);
                        ScheduleEmit();
                    });
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
    bool isShuffleActive = false;
    int autoRepeatMode = 0;
    bool isShuffleEnabled = true;
    bool isRepeatEnabled = true;

    winrt_media::GlobalSystemMediaTransportControlsSession session{nullptr};
    {
        std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
        session = g_currentSession;
    }

    bool useSnapshot = false;
    if (!session || !IsValidMediaSession(session)) {
        std::lock_guard<std::mutex> snapLock(g_validSnapshotMutex);
        if (session && g_lastValidTrackSnapshot.hasValue) {
            try {
                auto info = session.GetPlaybackInfo();
                if (info && info.PlaybackStatus() != winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Closed) {
                    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(
                        std::chrono::steady_clock::now() - g_lastValidTrackSnapshot.timestamp).count();
                    if (elapsed < 15) {
                        useSnapshot = true;
                    }
                }
            } catch (...) {}
        }

        if (!useSnapshot) {
            g_isMediaPlaying.store(false);
            std::ostringstream json;
            json << "{\"type\":\"media\",\"isAvailable\":false}";
            g_mediaCb(json.str().c_str());
            return;
        }
    }

    if (useSnapshot) {
        std::lock_guard<std::mutex> snapLock(g_validSnapshotMutex);
        sourceApp = g_lastValidTrackSnapshot.sourceApp;
        title = g_lastValidTrackSnapshot.title;
        artist = g_lastValidTrackSnapshot.artist;
        album = g_lastValidTrackSnapshot.album;
        albumArt = g_lastValidTrackSnapshot.albumArt;
        isPlaying = g_lastValidTrackSnapshot.isPlaying;
        durMs = g_lastValidTrackSnapshot.durMs;
        canGoNext = g_lastValidTrackSnapshot.canGoNext;
        canGoPrev = g_lastValidTrackSnapshot.canGoPrev;
        canPause = g_lastValidTrackSnapshot.canPause;
        canPlay = g_lastValidTrackSnapshot.canPlay;
        isShuffleActive = g_lastValidTrackSnapshot.isShuffleActive;
        isShuffleEnabled = g_lastValidTrackSnapshot.isShuffleEnabled;
        autoRepeatMode = g_lastValidTrackSnapshot.autoRepeatMode;
        isRepeatEnabled = g_lastValidTrackSnapshot.isRepeatEnabled;

        if (isPlaying) {
            auto now = std::chrono::steady_clock::now();
            double elapsedMs = std::chrono::duration<double, std::milli>(now - g_lastValidTrackSnapshot.timestamp).count();
            posMs = g_lastValidTrackSnapshot.posMs + elapsedMs;
            if (durMs > 0.0 && posMs > durMs) {
                posMs = durMs;
            }
        } else {
            posMs = g_lastValidTrackSnapshot.posMs;
        }
        g_isMediaPlaying.store(isPlaying);
    } else {
        try {
            sourceApp = Utf16ToUtf8(session.SourceAppUserModelId().c_str());
        } catch (...) {}

        try {
            auto props = session.TryGetMediaPropertiesAsync().get();
        if (props) {
            title = Utf16ToUtf8(props.Title().c_str());
            artist = Utf16ToUtf8(props.Artist().c_str());
            album = Utf16ToUtf8(props.AlbumTitle().c_str());

            std::string t = title;
            std::string a = artist;
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
                    auto prevIt = g_trackArtMap.find(g_currentTrackKey);
                    if (prevIt != g_trackArtMap.end() && !prevIt->second.empty()) {
                        g_previousTrackArt = prevIt->second;
                    }
                    g_currentTrackKey = trackKey;
                    g_trackStartTime = now;
                    ++g_artCorrectionGeneration;
                } else {
                    if (std::chrono::duration_cast<std::chrono::milliseconds>(now - g_trackStartTime).count() > 1000) {
                        g_previousTrackArt.clear();
                    }
                }
                staleArt = g_previousTrackArt;

                auto it = g_trackArtMap.find(trackKey);
                if (it != g_trackArtMap.end() && !it->second.empty()) {
                    albumArt = it->second;
                }
            }

            bool isInitialWindow = (std::chrono::duration_cast<std::chrono::milliseconds>(now - g_trackStartTime).count() < 400);
            bool allowStale = !isInitialWindow;

            // Attempt immediate extraction from current props if art is missing or properties changed
            bool propsDirty = g_mediaPropsDirty.exchange(false);
            if ((albumArt.empty() || propsDirty) && props.Thumbnail() != nullptr) {
                std::string fetchedArt = TryExtractThumbnail(props, staleArt, allowStale);
                if (!fetchedArt.empty() && fetchedArt != albumArt) {
                    std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                    g_trackArtMap[trackKey] = fetchedArt;
                    g_trackArtOrder.push_back(trackKey);
                    while (g_trackArtMap.size() > 30 && !g_trackArtOrder.empty()) {
                        std::string oldest = g_trackArtOrder.front();
                        g_trackArtOrder.erase(g_trackArtOrder.begin());
                        if (oldest != trackKey) g_trackArtMap.erase(oldest);
                    }
                    albumArt = fetchedArt;
                }
            }

            // Continuous 10-second auto-resync if playing track is still missing art
            if (albumArt.empty() && !trackKey.empty() && !isNewTrack) {
                bool shouldResync = false;
                {
                    std::lock_guard<std::mutex> resyncLock(g_artResyncMutex);
                    auto elapsedSec = std::chrono::duration_cast<std::chrono::seconds>(now - g_lastArtResyncTime).count();
                    if (elapsedSec >= 10) {
                        g_lastArtResyncTime = now;
                        shouldResync = true;
                    }
                }
                if (shouldResync) {
                    try {
                        auto freshProps = session.TryGetMediaPropertiesAsync().get();
                        if (freshProps) {
                            std::string freshArt = TryExtractThumbnail(freshProps, staleArt, true);
                            if (!freshArt.empty()) {
                                {
                                    std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                                    g_trackArtMap[trackKey] = freshArt;
                                    g_trackArtOrder.push_back(trackKey);
                                }
                                albumArt = freshArt;
                            }
                        }
                    } catch (...) {}
                }
            }

            // Dedicated background retry worker for asynchronously delivered poster thumbnails (such as YouTube in browsers)
            uint64_t curGen = g_artCorrectionGeneration.load();
            bool needsWorker = (albumArt.empty() || albumArt.size() < 10000 || propsDirty) && !trackKey.empty();
            if (needsWorker && g_activeWorkerGen.load() != curGen) {
                g_activeWorkerGen.store(curGen);
                {
                    std::lock_guard<std::mutex> resyncLock(g_artResyncMutex);
                    g_lastArtResyncTime = now;
                }
                std::string currentDeliveredArt = albumArt;
                std::thread([curGen, trackKey, currentDeliveredArt]() mutable {
                    ApartmentScope apt;
                    const int delaysMs[] = { 300, 600, 1000, 1500, 2200, 3200, 4500 };

                    for (int delay : delaysMs) {
                        std::this_thread::sleep_for(std::chrono::milliseconds(delay));
                        if (g_artCorrectionGeneration.load() != curGen) return; // Switched track

                        try {
                            winrt_media::GlobalSystemMediaTransportControlsSession curSession{nullptr};
                            {
                                std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
                                curSession = g_currentSession;
                            }

                            if (curSession) {
                                auto p = curSession.TryGetMediaPropertiesAsync().get();
                                if (p) {
                                    std::string pTitle = Utf16ToUtf8(p.Title().c_str());
                                    std::string pArtist = Utf16ToUtf8(p.Artist().c_str());
                                    TrimString(pTitle);
                                    TrimString(pArtist);
                                    std::string curTrackKey = pTitle + "_" + pArtist;
                                    if (curTrackKey != trackKey) return; // Switched track

                                    // After initial settling delay, accept any provided thumbnail as authentic
                                    bool allowStaleArt = (delay >= 500);
                                    std::string newArt = TryExtractThumbnail(p, "", allowStaleArt);
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
                                            if (newArt.size() > 10000) {
                                                return; // High-res poster acquired
                                            }
                                        }
                                    }
                                }
                            }
                        } catch (...) {}
                    }
                }).detach();
            }
        }
    } catch (...) {}

        try {
            auto info = session.GetPlaybackInfo();
            if (info) {
                isPlaying = (info.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing);
                g_isMediaPlaying.store(isPlaying);

                auto shuffleRef = info.IsShuffleActive();
                if (shuffleRef) {
                    isShuffleActive = shuffleRef.Value();
                }

                auto repeatRef = info.AutoRepeatMode();
                if (repeatRef) {
                    auto mode = repeatRef.Value();
                    if (mode == winrt::Windows::Media::MediaPlaybackAutoRepeatMode::Track) {
                        autoRepeatMode = 1;
                    } else if (mode == winrt::Windows::Media::MediaPlaybackAutoRepeatMode::List) {
                        autoRepeatMode = 2;
                    } else {
                        autoRepeatMode = 0;
                    }
                }

                auto controls = info.Controls();
                if (controls) {
                    canGoNext = controls.IsNextEnabled();
                    canGoPrev = controls.IsPreviousEnabled();
                    canPause = controls.IsPauseEnabled();
                    canPlay = controls.IsPlayEnabled();
                    if (controls.IsPlayPauseToggleEnabled()) {
                        if (isPlaying) {
                            canPause = true;
                        } else {
                            canPlay = true;
                        }
                    }
                    isShuffleEnabled = controls.IsShuffleEnabled();
                    isRepeatEnabled = controls.IsRepeatEnabled();
                }
            }
        } catch (...) {}

        if (title.empty() && artist.empty()) {
            g_isMediaPlaying.store(false);
            std::ostringstream json;
            json << "{\"type\":\"media\",\"isAvailable\":false}";
            g_mediaCb(json.str().c_str());
            return;
        }

        try {
            auto timeline = session.GetTimelineProperties();
            if (timeline) {
                double rawPosMs = std::chrono::duration<double, std::milli>(timeline.Position()).count();
                durMs = std::chrono::duration<double, std::milli>(timeline.EndTime()).count();

                if (isPlaying) {
                    auto lastUpdated = timeline.LastUpdatedTime();
                    auto now = winrt::clock::now();
                    if (now > lastUpdated) {
                        double elapsedMs = std::chrono::duration<double, std::milli>(now - lastUpdated).count();
                        // Media players report an anchor Position and LastUpdatedTime. While actively playing,
                        // extrapolate true position continuously from the anchor without arbitrary cutoff.
                        posMs = rawPosMs + elapsedMs;
                        if (durMs > 0.0 && posMs > durMs) {
                            posMs = durMs;
                        }
                    } else {
                        posMs = rawPosMs;
                    }
                } else {
                    posMs = rawPosMs;
                }
            }
        } catch (...) {}

        // Record complete valid track snapshot for session persistence and fallback
        if (!title.empty() && (!artist.empty() || !albumArt.empty())) {
            std::lock_guard<std::mutex> snapLock(g_validSnapshotMutex);
            g_lastValidTrackSnapshot.sourceApp = sourceApp;
            g_lastValidTrackSnapshot.title = title;
            g_lastValidTrackSnapshot.artist = artist;
            g_lastValidTrackSnapshot.album = album;
            g_lastValidTrackSnapshot.albumArt = albumArt;
            g_lastValidTrackSnapshot.posMs = posMs;
            g_lastValidTrackSnapshot.durMs = durMs;
            g_lastValidTrackSnapshot.isPlaying = isPlaying;
            g_lastValidTrackSnapshot.canGoNext = canGoNext;
            g_lastValidTrackSnapshot.canGoPrev = canGoPrev;
            g_lastValidTrackSnapshot.canPause = canPause;
            g_lastValidTrackSnapshot.canPlay = canPlay;
            g_lastValidTrackSnapshot.isShuffleActive = isShuffleActive;
            g_lastValidTrackSnapshot.isShuffleEnabled = isShuffleEnabled;
            g_lastValidTrackSnapshot.autoRepeatMode = autoRepeatMode;
            g_lastValidTrackSnapshot.isRepeatEnabled = isRepeatEnabled;
            g_lastValidTrackSnapshot.timestamp = std::chrono::steady_clock::now();
            g_lastValidTrackSnapshot.hasValue = true;
        }
    }

    std::string sessionsJson = "[]";
    std::string activeSessionId = ComputeSessionUniqueId(sourceApp, title, artist);
    try {
        if (g_sessionManager) {
            auto now = std::chrono::steady_clock::now();
            std::lock_guard<std::mutex> sLock(g_sessionsJsonMutex);
            auto elapsedMs = std::chrono::duration_cast<std::chrono::milliseconds>(now - g_lastSessionsFetchTime).count();

            if (elapsedMs < 1200 && g_cachedSessionsJson != "[]") {
                sessionsJson = g_cachedSessionsJson;
            } else {
                auto allSessions = g_sessionManager.GetSessions();
                std::ostringstream sJson;
                sJson << "[";
                bool first = true;
                std::unordered_set<std::string> seenUniqueIds;

                for (auto const& s : allSessions) {
                    if (!s) continue;
                    try {
                        auto pInfo = s.GetPlaybackInfo();
                        if (!pInfo) continue;
                        if (pInfo.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Closed) {
                            continue;
                        }

                        winrt_media::GlobalSystemMediaTransportControlsSessionMediaProperties props{nullptr};
                        try {
                            props = s.TryGetMediaPropertiesAsync().get();
                        } catch (...) {}
                        if (!props) continue;

                        if (!IsValidMediaSession(s, &pInfo, &props)) continue;

                        std::string sId = Utf16ToUtf8(s.SourceAppUserModelId().c_str());
                        std::string sApp = sId;
                        size_t slash = sApp.find_last_of("\\/");
                        if (slash != std::string::npos) sApp = sApp.substr(slash + 1);

                        std::string sTitle = Utf16ToUtf8(props.Title().c_str());
                        std::string sArtist = Utf16ToUtf8(props.Artist().c_str());
                        TrimString(sTitle);
                        TrimString(sArtist);

                        if (sTitle.empty() && sArtist.empty()) continue;

                        bool sPlaying = (pInfo.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing);

                        std::string uniqueId = ComputeSessionUniqueId(sId, sTitle, sArtist);
                        if (seenUniqueIds.find(uniqueId) != seenUniqueIds.end()) {
                            continue;
                        }
                        seenUniqueIds.insert(uniqueId);

                        if (s == session || (sId == sourceApp && sTitle == title && sArtist == artist)) {
                            activeSessionId = uniqueId;
                        }

                        if (!first) sJson << ",";
                        first = false;
                        sJson << "{"
                              << "\"id\":\"" << EscapeJsonString(uniqueId) << "\","
                              << "\"sourceApp\":\"" << EscapeJsonString(sApp) << "\","
                              << "\"title\":\"" << EscapeJsonString(sTitle) << "\","
                              << "\"artist\":\"" << EscapeJsonString(sArtist) << "\","
                              << "\"isPlaying\":" << (sPlaying ? "true" : "false")
                              << "}";
                    } catch (...) {}
                }
                sJson << "]";
                g_cachedSessionsJson = sJson.str();
                g_lastSessionsFetchTime = now;
                sessionsJson = g_cachedSessionsJson;
            }
        }
    } catch (...) {}

    auto [volume, isMuted] = GetMediaAudioVolumeAndMute(sourceApp);

    std::ostringstream json;
    json << "{\"type\":\"media\","
         << "\"isAvailable\":true,"
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

    if (g_mediaCb) {
        g_mediaCb(json.str().c_str());
    }
}

static float g_mediaVolumeStore = 100.0f;
static bool g_mediaMuteStore = false;
static std::mutex g_mediaVolStateMutex;
static std::chrono::steady_clock::time_point g_lastVolumeSetTime{};

static std::vector<std::string> ExtractAppTokens(const std::string& sourceApp) {
    std::vector<std::string> tokens;
    if (sourceApp.empty()) return tokens;

    std::string s = sourceApp;
    std::transform(s.begin(), s.end(), s.begin(), [](unsigned char c) { return static_cast<char>(std::tolower(c)); });

    // Common app aliases
    if (s.find("308046b0af4a39cb") != std::string::npos || s.find("firefox") != std::string::npos) {
        tokens.push_back("firefox");
    }
    if (s.find("spotify") != std::string::npos) {
        tokens.push_back("spotify");
    }
    if (s.find("apple") != std::string::npos || s.find("applemusic") != std::string::npos) {
        tokens.push_back("applemusic");
        tokens.push_back("apple");
    }
    if (s.find("chrome") != std::string::npos) {
        tokens.push_back("chrome");
    }
    if (s.find("edge") != std::string::npos || s.find("msedge") != std::string::npos) {
        tokens.push_back("msedge");
        tokens.push_back("edge");
    }
    if (s.find("vlc") != std::string::npos) {
        tokens.push_back("vlc");
    }
    if (s.find("zune") != std::string::npos || s.find("microsoft.zunemusic") != std::string::npos) {
        tokens.push_back("zune");
        tokens.push_back("wmplayer");
        tokens.push_back("media");
    }
    if (s.find("tidal") != std::string::npos) {
        tokens.push_back("tidal");
    }
    if (s.find("brave") != std::string::npos) {
        tokens.push_back("brave");
    }
    if (s.find("opera") != std::string::npos) {
        tokens.push_back("opera");
    }

    // Split on delimiters
    std::string cur;
    for (char c : s) {
        if (isalnum(static_cast<unsigned char>(c))) {
            cur += c;
        } else {
            if (cur.length() >= 3 && cur != "exe" && cur != "app" && cur != "win" && cur != "inc") {
                tokens.push_back(cur);
            }
            cur.clear();
        }
    }
    if (cur.length() >= 3 && cur != "exe" && cur != "app" && cur != "win" && cur != "inc") {
        tokens.push_back(cur);
    }

    size_t lastSlash = s.find_last_of("\\/");
    std::string base = (lastSlash != std::string::npos) ? s.substr(lastSlash + 1) : s;
    size_t pExe = base.find(".exe");
    if (pExe != std::string::npos) base = base.substr(0, pExe);
    if (!base.empty() && base.length() > 2) {
        tokens.push_back(base);
    }

    return tokens;
}

struct SessionMatch {
    ISimpleAudioVolume* pVol = nullptr;
    int score = 0;
    float peak = 0.0f;
    bool isActive = false;
};

static bool GetAllAppSessionVolumes(const std::string& sourceApp, std::vector<ISimpleAudioVolume*>& outVolumes, ISimpleAudioVolume** ppPrimaryVolume) {
    if (ppPrimaryVolume) *ppPrimaryVolume = nullptr;
    outVolumes.clear();

    IMMDeviceEnumerator* pEnumerator = nullptr;
    if (FAILED(CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, __uuidof(IMMDeviceEnumerator), (void**)&pEnumerator)) || !pEnumerator) return false;

    IMMDevice* pDevice = nullptr;
    HRESULT hr = pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDevice);
    pEnumerator->Release();
    if (FAILED(hr) || !pDevice) return false;

    IAudioSessionManager2* pSessionMgr = nullptr;
    hr = pDevice->Activate(__uuidof(IAudioSessionManager2), CLSCTX_ALL, NULL, (void**)&pSessionMgr);
    pDevice->Release();
    if (FAILED(hr) || !pSessionMgr) return false;

    IAudioSessionEnumerator* pSessionEnum = nullptr;
    hr = pSessionMgr->GetSessionEnumerator(&pSessionEnum);
    pSessionMgr->Release();
    if (FAILED(hr) || !pSessionEnum) return false;

    int count = 0;
    if (FAILED(pSessionEnum->GetCount(&count))) {
        pSessionEnum->Release();
        return false;
    }

    std::string appLower = sourceApp;
    std::transform(appLower.begin(), appLower.end(), appLower.begin(), [](unsigned char c) { return static_cast<char>(std::tolower(c)); });
    auto tokens = ExtractAppTokens(sourceApp);

    std::vector<SessionMatch> matches;
    SessionMatch bestFallback;

    for (int i = 0; i < count; ++i) {
        IAudioSessionControl* pSessionControl = nullptr;
        if (SUCCEEDED(pSessionEnum->GetSession(i, &pSessionControl)) && pSessionControl) {
            IAudioSessionControl2* pControl2 = nullptr;
            if (SUCCEEDED(pSessionControl->QueryInterface(__uuidof(IAudioSessionControl2), (void**)&pControl2)) && pControl2) {
                if (pControl2->IsSystemSoundsSession() == S_OK) {
                    pControl2->Release();
                    pSessionControl->Release();
                    continue;
                }

                int appScore = 0;
                float peak = 0.0f;
                AudioSessionState state = AudioSessionStateInactive;
                pControl2->GetState(&state);
                bool isActive = (state == AudioSessionStateActive);

                IAudioMeterInformation* pMeter = nullptr;
                if (SUCCEEDED(pSessionControl->QueryInterface(__uuidof(IAudioMeterInformation), (void**)&pMeter)) && pMeter) {
                    pMeter->GetPeakValue(&peak);
                    pMeter->Release();
                }

                // 1. Check Session Identifier & Display Name
                LPWSTR pswSessionId = nullptr;
                std::string sessId;
                if (SUCCEEDED(pControl2->GetSessionIdentifier(&pswSessionId)) && pswSessionId) {
                    sessId = Utf16ToUtf8(pswSessionId);
                    CoTaskMemFree(pswSessionId);
                    std::transform(sessId.begin(), sessId.end(), sessId.begin(), [](unsigned char c) { return static_cast<char>(std::tolower(c)); });
                }

                LPWSTR pswDisplayName = nullptr;
                std::string dispName;
                if (SUCCEEDED(pControl2->GetDisplayName(&pswDisplayName)) && pswDisplayName) {
                    dispName = Utf16ToUtf8(pswDisplayName);
                    CoTaskMemFree(pswDisplayName);
                    std::transform(dispName.begin(), dispName.end(), dispName.begin(), [](unsigned char c) { return static_cast<char>(std::tolower(c)); });
                }

                // 2. Check Process Executable Name
                DWORD pid = 0;
                pControl2->GetProcessId(&pid);
                std::string procName;
                std::string procPath;
                if (pid != 0) {
                    HANDLE hProcess = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, FALSE, pid);
                    if (hProcess) {
                        char szPath[MAX_PATH] = {0};
                        DWORD dwSize = MAX_PATH;
                        if (QueryFullProcessImageNameA(hProcess, 0, szPath, &dwSize)) {
                            procPath = szPath;
                            size_t lastSlash = procPath.find_last_of("\\/");
                            if (lastSlash != std::string::npos) procName = procPath.substr(lastSlash + 1);
                            else procName = procPath;
                            std::transform(procName.begin(), procName.end(), procName.begin(), ::tolower);
                            std::transform(procPath.begin(), procPath.end(), procPath.begin(), ::tolower);
                            size_t pExe = procName.find(".exe");
                            if (pExe != std::string::npos) procName = procName.substr(0, pExe);
                        }
                        CloseHandle(hProcess);
                    }
                }

                // Match against tokens & app name
                if (!appLower.empty()) {
                    if (!procName.empty() && (appLower.find(procName) != std::string::npos || procName.find(appLower) != std::string::npos)) {
                        appScore += 100;
                    }
                    if (!sessId.empty() && sessId.find(appLower) != std::string::npos) {
                        appScore += 80;
                    }
                    if (!dispName.empty() && (appLower.find(dispName) != std::string::npos || dispName.find(appLower) != std::string::npos)) {
                        appScore += 70;
                    }
                }

                for (const auto& token : tokens) {
                    if (token.empty()) continue;
                    if (!procName.empty() && procName.find(token) != std::string::npos) appScore += 60;
                    if (!procPath.empty() && procPath.find(token) != std::string::npos) appScore += 40;
                    if (!sessId.empty() && sessId.find(token) != std::string::npos) appScore += 40;
                    if (!dispName.empty() && dispName.find(token) != std::string::npos) appScore += 30;
                }

                int score = appScore + (isActive ? 40 : 0) + (peak > 0.001f ? 20 : 0);

                ISimpleAudioVolume* pVol = nullptr;
                pControl2->QueryInterface(__uuidof(ISimpleAudioVolume), (void**)&pVol);

                // A session is ONLY considered a match for the target app if appScore > 0!
                // Unrelated active apps (like browser, Discord, game) with appScore == 0 must NEVER be added to matches.
                bool isMatch = false;
                if (!appLower.empty() || !tokens.empty()) {
                    isMatch = (appScore > 0);
                } else {
                    isMatch = (score >= 50);
                }

                if (pVol) {
                    if (isMatch) {
                        matches.push_back({pVol, score, peak, isActive});
                    } else {
                        int fallbackScore = (isActive ? 20 : 0) + (peak > 0.001f ? 50 : 0);
                        if (fallbackScore > bestFallback.score || (!bestFallback.pVol && isActive)) {
                            if (bestFallback.pVol) bestFallback.pVol->Release();
                            bestFallback = {pVol, fallbackScore, peak, isActive};
                            pVol = nullptr; // ownership transferred
                        }
                    }
                }

                if (pVol && !isMatch) {
                    pVol->Release();
                }

                pControl2->Release();
            }
            pSessionControl->Release();
        }
    }

    pSessionEnum->Release();

    if (!matches.empty()) {
        if (bestFallback.pVol) bestFallback.pVol->Release();

        // Stable sort by score descending, then peak descending to prevent flapping
        std::stable_sort(matches.begin(), matches.end(), [](const SessionMatch& a, const SessionMatch& b) {
            if (a.score != b.score) return a.score > b.score;
            return a.peak > b.peak;
        });

        for (const auto& m : matches) {
            outVolumes.push_back(m.pVol);
        }

        if (ppPrimaryVolume && !outVolumes.empty()) {
            *ppPrimaryVolume = outVolumes[0];
        }
        return true;
    }

    if (bestFallback.pVol) {
        if (appLower.empty()) {
            outVolumes.push_back(bestFallback.pVol);
            if (ppPrimaryVolume) *ppPrimaryVolume = bestFallback.pVol;
            return true;
        }
        bestFallback.pVol->Release();
    }

    return false;
}

static std::pair<float, bool> GetMasterAudioVolumeAndMute() {
    IMMDeviceEnumerator* pEnumerator = nullptr;
    if (FAILED(CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, __uuidof(IMMDeviceEnumerator), (void**)&pEnumerator)) || !pEnumerator) return {100.0f, false};

    IMMDevice* pDevice = nullptr;
    HRESULT hr = pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDevice);
    pEnumerator->Release();
    if (FAILED(hr) || !pDevice) return {100.0f, false};

    IAudioEndpointVolume* pEndVol = nullptr;
    hr = pDevice->Activate(__uuidof(IAudioEndpointVolume), CLSCTX_ALL, NULL, (void**)&pEndVol);
    pDevice->Release();
    if (FAILED(hr) || !pEndVol) return {100.0f, false};

    float vol = 1.0f;
    BOOL muted = FALSE;
    pEndVol->GetMasterVolumeLevelScalar(&vol);
    pEndVol->GetMute(&muted);
    pEndVol->Release();

    return {vol * 100.0f, (muted == TRUE)};
}

static void SetMasterAudioVolume(float percent) {
    IMMDeviceEnumerator* pEnumerator = nullptr;
    if (FAILED(CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, __uuidof(IMMDeviceEnumerator), (void**)&pEnumerator)) || !pEnumerator) return;

    IMMDevice* pDevice = nullptr;
    HRESULT hr = pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDevice);
    pEnumerator->Release();
    if (FAILED(hr) || !pDevice) return;

    IAudioEndpointVolume* pEndVol = nullptr;
    hr = pDevice->Activate(__uuidof(IAudioEndpointVolume), CLSCTX_ALL, NULL, (void**)&pEndVol);
    pDevice->Release();
    if (FAILED(hr) || !pEndVol) return;

    float clamped = std::max(0.0f, std::min(1.0f, percent / 100.0f));
    pEndVol->SetMasterVolumeLevelScalar(clamped, NULL);
    pEndVol->Release();
}

static void ToggleMasterAudioMute() {
    IMMDeviceEnumerator* pEnumerator = nullptr;
    if (FAILED(CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, __uuidof(IMMDeviceEnumerator), (void**)&pEnumerator)) || !pEnumerator) return;

    IMMDevice* pDevice = nullptr;
    HRESULT hr = pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDevice);
    pEnumerator->Release();
    if (FAILED(hr) || !pDevice) return;

    IAudioEndpointVolume* pEndVol = nullptr;
    hr = pDevice->Activate(__uuidof(IAudioEndpointVolume), CLSCTX_ALL, NULL, (void**)&pEndVol);
    pDevice->Release();
    if (FAILED(hr) || !pEndVol) return;

    BOOL muted = FALSE;
    pEndVol->GetMute(&muted);
    pEndVol->SetMute(!muted, NULL);
    pEndVol->Release();
}

static std::pair<float, bool> GetMediaAudioVolumeAndMute(const std::string& sourceApp) {
    {
        std::lock_guard<std::mutex> stLock(g_mediaVolStateMutex);
        auto elapsedMs = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::steady_clock::now() - g_lastVolumeSetTime).count();
        if (elapsedMs < 700) {
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

        for (auto* v : volumes) {
            if (v) v->Release();
        }

        float volPercent = vol * 100.0f;
        {
            std::lock_guard<std::mutex> stLock(g_mediaVolStateMutex);
            g_mediaVolumeStore = volPercent;
            g_mediaMuteStore = (muted == TRUE);
        }
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
        g_lastVolumeSetTime = std::chrono::steady_clock::now();
    }

    std::lock_guard<std::mutex> lock(g_audioVolMutex);
    ComInitScope com;

    std::vector<ISimpleAudioVolume*> volumes;
    ISimpleAudioVolume* pPrimary = nullptr;

    if (GetAllAppSessionVolumes(sourceApp, volumes, &pPrimary) && !volumes.empty()) {
        float volScalar = clamped / 100.0f;
        for (auto* pVol : volumes) {
            if (pVol) {
                pVol->SetMasterVolume(volScalar, NULL);
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
    }

    std::lock_guard<std::mutex> lock(g_audioVolMutex);
    ComInitScope com;

    std::vector<ISimpleAudioVolume*> volumes;
    ISimpleAudioVolume* pPrimary = nullptr;

    if (GetAllAppSessionVolumes(sourceApp, volumes, &pPrimary) && !volumes.empty()) {
        BOOL bMute = nextMute ? TRUE : FALSE;
        for (auto* pVol : volumes) {
            if (pVol) {
                pVol->SetMute(bMute, NULL);
                pVol->Release();
            }
        }
    }
}

static void SendVolumeKey(BYTE vk) {
    UINT scanCode = MapVirtualKeyA(vk, 0);
    keybd_event(vk, static_cast<BYTE>(scanCode), 0x0001, 0);
    std::this_thread::sleep_for(std::chrono::milliseconds(15));
    keybd_event(vk, static_cast<BYTE>(scanCode), 0x0001 | 0x0002, 0);
}

void StartMediaService(NativeMediaCallback cb) {
    if (g_mediaRunning) return;
    g_mediaCb = cb;
    g_mediaRunning = true;

    g_mediaSessionPollThread = std::thread([]() {
        ApartmentScope apt;
        try {
            g_sessionManager = winrt_media::GlobalSystemMediaTransportControlsSessionManager::RequestAsync().get();
            if (g_sessionManager) {
                g_currentSessionChangedToken = g_sessionManager.CurrentSessionChanged([](auto const&, auto const&) {
                    ScheduleEmit();
                });
                g_sessionsChangedToken = g_sessionManager.SessionsChanged([](auto const&, auto const&) {
                    {
                        std::lock_guard<std::mutex> sLock(g_sessionsJsonMutex);
                        g_lastSessionsFetchTime = {};
                    }
                    ScheduleEmit();
                });
            }
        } catch (...) {}

        UpdateCurrentSession();

        while (g_mediaRunning) {
            // Heartbeat fallback: real-time updates are handled asynchronously by WinRT push callbacks
            std::this_thread::sleep_for(std::chrono::milliseconds(3000));
            if (g_mediaRunning) {
                UpdateCurrentSession();
            }
        }
    });
}

void StopMediaService() {
    if (!g_mediaRunning) return;
    g_mediaRunning = false;

    if (g_mediaSessionPollThread.joinable()) g_mediaSessionPollThread.join();

    {
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
}

int ExecuteMediaCommand(const char* command) {
    if (!command) return 0;
    ApartmentScope apt;
    std::string cmd(command);

    std::string sourceApp;
    {
        std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
        if (g_currentSession) {
            try { sourceApp = Utf16ToUtf8(g_currentSession.SourceAppUserModelId().c_str()); } catch (...) {}
        }
    }

    if (cmd.rfind("set_volume:", 0) == 0) {
        std::string payload = cmd.substr(11);
        size_t colon = payload.find(':');
        float vol = 0.0f;
        if (colon != std::string::npos) {
            vol = std::stof(payload.substr(0, colon));
            std::string explicitApp = payload.substr(colon + 1);
            if (!explicitApp.empty()) sourceApp = explicitApp;
        } else {
            vol = std::stof(payload);
        }
        SetMediaAudioVolume(sourceApp, vol);
        EmitMediaState(true);
        return 1;
    }
    if (cmd == "volume_up") {
        auto [curVol, _] = GetMediaAudioVolumeAndMute(sourceApp);
        SetMediaAudioVolume(sourceApp, curVol + 5.0f);
        EmitMediaState(true);
        return 1;
    }
    if (cmd == "volume_down") {
        auto [curVol, _] = GetMediaAudioVolumeAndMute(sourceApp);
        SetMediaAudioVolume(sourceApp, curVol - 5.0f);
        EmitMediaState(true);
        return 1;
    }
    if (cmd.rfind("toggle_mute:", 0) == 0) {
        std::string explicitApp = cmd.substr(12);
        if (!explicitApp.empty()) sourceApp = explicitApp;
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
        double posMs = 0.0;
        std::string explicitTarget;
        if (colon != std::string::npos) {
            posMs = std::stod(payload.substr(0, colon));
            explicitTarget = payload.substr(colon + 1);
        } else {
            posMs = std::stod(payload);
        }

        winrt_media::GlobalSystemMediaTransportControlsSession session{nullptr};
        if (!explicitTarget.empty() && g_sessionManager) {
            try {
                auto sessions = g_sessionManager.GetSessions();
                for (auto const& s : sessions) {
                    if (!s) continue;
                    try {
                        std::string sId = Utf16ToUtf8(s.SourceAppUserModelId().c_str());
                        std::string sTitle = "";
                        std::string sArtist = "";
                        auto props = s.TryGetMediaPropertiesAsync().get();
                        if (props) {
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
                    } catch (...) {}
                }
            } catch (...) {}
        }
        if (!session) {
            std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
            session = g_currentSession;
        }

        if (session) {
            try {
                auto ticks = static_cast<int64_t>(posMs * 10000.0);
                auto asyncOp = session.TryChangePlaybackPositionAsync(ticks);
                if (asyncOp) {
                    asyncOp.get();
                }
            } catch (...) {}

            std::thread([session, posMs]() {
                ApartmentScope apt;
                // Web browsers do not always trigger push events when scrubbing.
                // We poll briefly until the session reports near the target position.
                for (int i = 0; i < 8; ++i) {
                    std::this_thread::sleep_for(std::chrono::milliseconds(80));
                    try {
                        auto tl = session.GetTimelineProperties();
                        if (tl) {
                            double cur = std::chrono::duration<double, std::milli>(tl.Position()).count();
                            if (std::abs(cur - posMs) < 1500.0) {
                                break;
                            }
                        }
                    } catch (...) {}
                }
                UpdateCurrentSession();

                // Settle once more after audio and video buffers resume
                std::this_thread::sleep_for(std::chrono::milliseconds(500));
                UpdateCurrentSession();
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
            if (cmd == "play_pause") session.TryTogglePlayPauseAsync();
            else if (cmd == "play") session.TryPlayAsync();
            else if (cmd == "pause") session.TryPauseAsync();
            else if (cmd == "next") session.TrySkipNextAsync();
            else if (cmd == "previous" || cmd == "prev") session.TrySkipPreviousAsync();
            else if (cmd == "shuffle" || cmd == "toggle_shuffle") {
                auto info = session.GetPlaybackInfo();
                bool cur = false;
                if (info && info.IsShuffleActive()) cur = info.IsShuffleActive().Value();
                session.TryChangeShuffleActiveAsync(!cur);
            }
            else if (cmd == "repeat" || cmd == "toggle_repeat") {
                auto info = session.GetPlaybackInfo();
                int nextMode = 1;
                if (info && info.AutoRepeatMode()) {
                    auto cur = info.AutoRepeatMode().Value();
                    if (cur == winrt::Windows::Media::MediaPlaybackAutoRepeatMode::None) nextMode = 1;
                    else if (cur == winrt::Windows::Media::MediaPlaybackAutoRepeatMode::Track) nextMode = 2;
                    else nextMode = 0;
                }
                winrt::Windows::Media::MediaPlaybackAutoRepeatMode target = winrt::Windows::Media::MediaPlaybackAutoRepeatMode::None;
                if (nextMode == 1) target = winrt::Windows::Media::MediaPlaybackAutoRepeatMode::Track;
                else if (nextMode == 2) target = winrt::Windows::Media::MediaPlaybackAutoRepeatMode::List;
                session.TryChangeAutoRepeatModeAsync(target);
            }
        } catch (...) {}

        std::thread([]() {
            ApartmentScope apt;
            std::this_thread::sleep_for(std::chrono::milliseconds(30));
            UpdateCurrentSession();
            std::this_thread::sleep_for(std::chrono::milliseconds(120));
            UpdateCurrentSession();
            std::this_thread::sleep_for(std::chrono::milliseconds(300));
            UpdateCurrentSession();
        }).detach();
    }

    return 1;
}
