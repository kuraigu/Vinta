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
static std::string g_currentTrackKey;
static std::string g_previousTrackArt;
static std::mutex g_artCacheMutex;

static std::mutex g_audioVolMutex;

static std::string g_selectedSessionId;
static std::chrono::steady_clock::time_point g_selectedSessionTime;
static std::mutex g_selectedSessionMutex;

static std::atomic<uint64_t> g_artCorrectionGeneration{0};
static std::chrono::steady_clock::time_point g_lastArtResyncTime{};
static std::mutex g_artResyncMutex;

static std::string g_cachedSessionsJson = "[]";
static std::chrono::steady_clock::time_point g_lastSessionsFetchTime{};
static std::mutex g_sessionsJsonMutex;

static std::string TryExtractThumbnail(winrt_media::GlobalSystemMediaTransportControlsSessionMediaProperties const& props, const std::string& staleArt = "") {
    if (!props) return "";
    try {
        auto thumbRef = props.Thumbnail();
        if (thumbRef) {
            auto stream = thumbRef.OpenReadAsync().get();
            if (stream) {
                std::string fetched = StreamToBase64(stream);
                if (!fetched.empty() && (staleArt.empty() || fetched != staleArt)) {
                    return fetched;
                }
            }
        }
    } catch (...) {}
    return "";
}

static void UpdateCurrentSession();
static void EmitMediaState(bool includeVolume = true);
static void ScheduleEmit();
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
    ScheduleEmit();
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

static bool IsValidMediaSession(winrt_media::GlobalSystemMediaTransportControlsSession const& session) {
    if (!session) return false;
    try {
        auto info = session.GetPlaybackInfo();
        if (!info) return false;

        auto status = info.PlaybackStatus();
        if (status == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Closed ||
            status == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Stopped) {
            return false;
        }

        // Must support at least one interactive media playback control
        auto controls = info.Controls();
        if (!controls) return false;
        bool hasControls = controls.IsPlayEnabled() || controls.IsPauseEnabled() ||
                           controls.IsStopEnabled() || controls.IsNextEnabled() ||
                           controls.IsPreviousEnabled();
        if (!hasControls) return false;

        // Must have valid track metadata
        auto props = session.TryGetMediaPropertiesAsync().get();
        if (!props) return false;

        std::string title = Utf16ToUtf8(props.Title().c_str());
        std::string artist = Utf16ToUtf8(props.Artist().c_str());
        std::string album = Utf16ToUtf8(props.AlbumTitle().c_str());

        auto trim = [](std::string& s) {
            size_t start = s.find_first_not_of(" \t\r\n");
            if (start == std::string::npos) {
                s.clear();
                return;
            }
            size_t end = s.find_last_not_of(" \t\r\n");
            s = s.substr(start, end - start + 1);
        };
        trim(title);
        trim(artist);
        trim(album);

        if (title.empty() && artist.empty()) {
            return false;
        }

        // Non-media applications (e.g. game launchers, screen or audio recorders) may register
        // a dummy audio session where the title is just the application or window name.
        std::string sourceApp;
        try {
            sourceApp = Utf16ToUtf8(session.SourceAppUserModelId().c_str());
            trim(sourceApp);
        } catch (...) {}

        if (artist.empty() && album.empty()) {
            std::string appName = sourceApp;
            size_t slash = appName.find_last_of("\\/");
            if (slash != std::string::npos) appName = appName.substr(slash + 1);
            size_t dot = appName.rfind('.');
            if (dot != std::string::npos) appName = appName.substr(0, dot);

            auto toLower = [](std::string s) {
                std::transform(s.begin(), s.end(), s.begin(), [](unsigned char c) {
                    return static_cast<char>(std::tolower(c));
                });
                return s;
            };

            std::string lowerTitle = toLower(title);
            std::string lowerAppName = toLower(appName);

            // If the title matches the process name and there are no playlist navigation controls,
            // this is an application audio stream rather than a media player.
            if (!lowerAppName.empty() && (lowerTitle == lowerAppName || lowerTitle.find(lowerAppName) != std::string::npos)) {
                if (!controls.IsNextEnabled() && !controls.IsPreviousEnabled()) {
                    return false;
                }
            }
        }

        return true;
    } catch (...) {
        return false;
    }
}

static void UpdateCurrentSession() {
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
                int sIndex = 0;
                for (auto const& s : sessions) {
                    try {
                        std::string sId = Utf16ToUtf8(s.SourceAppUserModelId().c_str());
                        std::string sTitle = "";
                        try {
                            auto props = s.TryGetMediaPropertiesAsync().get();
                            if (props) {
                                sTitle = Utf16ToUtf8(props.Title().c_str());
                            }
                        } catch (...) {}

                        std::string compositeId = sId + "::" + std::to_string(sIndex) + "::" + sTitle;
                        if (selectedId == compositeId || (!sTitle.empty() && selectedId.find(sTitle) != std::string::npos) || selectedId == sId) {
                            chosenSession = s;
                            break;
                        }
                    } catch (...) {}
                    sIndex++;
                }
            } catch (...) {}
        }

        // Check if chosen session is actively playing or within recent selection window
        bool chosenIsPlaying = false;
        if (chosenSession && IsValidMediaSession(chosenSession)) {
            try {
                auto info = chosenSession.GetPlaybackInfo();
                if (info && info.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing) {
                    chosenIsPlaying = true;
                }
            } catch (...) {}
        }

        auto now = std::chrono::steady_clock::now();
        bool isRecentSelection = (std::chrono::duration_cast<std::chrono::milliseconds>(now - selectedTime).count() < 4000);

        if (chosenSession && (chosenIsPlaying || isRecentSelection)) {
            // User's chosen session has priority if playing or freshly selected
            targetSession = chosenSession;
        } else {
            // Check if Windows's own focused session is actively playing and valid
            try {
                winrt_media::GlobalSystemMediaTransportControlsSession focusedSession = g_sessionManager.GetCurrentSession();
                if (focusedSession && IsValidMediaSession(focusedSession)) {
                    auto info = focusedSession.GetPlaybackInfo();
                    if (info && info.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing) {
                        targetSession = focusedSession;
                    }
                }
            } catch (...) {}

            // If focused session is not playing, scan other sessions for an actively playing session
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

            // Fallback if none are actively playing
            if (!targetSession) {
                if (chosenSession && IsValidMediaSession(chosenSession)) {
                    targetSession = chosenSession;
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
    bool isShuffleActive = false;
    int autoRepeatMode = 0;
    bool isShuffleEnabled = true;
    bool isRepeatEnabled = true;

    winrt_media::GlobalSystemMediaTransportControlsSession session{nullptr};
    {
        std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
        session = g_currentSession;
    }

    if (!session || !IsValidMediaSession(session)) {
        g_isMediaPlaying.store(false);
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
            std::string staleArt;
            bool isNewTrack = false;
            {
                std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                if (trackKey != g_currentTrackKey) {
                    isNewTrack = true;
                    auto prevIt = g_trackArtMap.find(g_currentTrackKey);
                    if (prevIt != g_trackArtMap.end() && !prevIt->second.empty()) {
                        g_previousTrackArt = prevIt->second;
                    }
                    g_currentTrackKey = trackKey;
                }
                staleArt = g_previousTrackArt;

                auto it = g_trackArtMap.find(trackKey);
                if (it != g_trackArtMap.end() && !it->second.empty()) {
                    albumArt = it->second;
                }
            }

            // Attempt immediate extraction from current props if art is missing
            if (albumArt.empty() && props.Thumbnail() != nullptr) {
                std::string fetchedArt = TryExtractThumbnail(props, staleArt);
                if (!fetchedArt.empty()) {
                    std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                    g_trackArtMap[trackKey] = fetchedArt;
                    albumArt = fetchedArt;
                }
            }

            // Continuous 10-second auto-resync:
            // If the song is active and albumArt is still missing, retry at least every 10 seconds
            if (albumArt.empty() && !trackKey.empty() && !isNewTrack) {
                auto now = std::chrono::steady_clock::now();
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
                            std::string freshArt = TryExtractThumbnail(freshProps, staleArt);
                            if (!freshArt.empty()) {
                                {
                                    std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                                    g_trackArtMap[trackKey] = freshArt;
                                }
                                albumArt = freshArt;
                            }
                        }
                    } catch (...) {}
                }
            }

            // Rapid initial retry worker: spawned strictly when a new track is detected without full-res art.
            // Generation counter only increments on actual track change to avoid routine progress cancels.
            bool needsWorker = isNewTrack && (albumArt.empty() || albumArt.size() < 10000);
            if (needsWorker && !trackKey.empty()) {
                {
                    std::lock_guard<std::mutex> resyncLock(g_artResyncMutex);
                    g_lastArtResyncTime = std::chrono::steady_clock::now();
                }
                uint64_t gen = ++g_artCorrectionGeneration;
                std::thread([gen, trackKey, staleArt]() {
                    ApartmentScope apt;
                    const int delaysMs[] = { 300, 500, 800, 1200, 1800, 2500, 3500 };

                    for (int delay : delaysMs) {
                        std::this_thread::sleep_for(std::chrono::milliseconds(delay));
                        if (g_artCorrectionGeneration.load() != gen) return;

                        try {
                            winrt_media::GlobalSystemMediaTransportControlsSession curSession{nullptr};
                            {
                                std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
                                curSession = g_currentSession;
                            }

                            if (curSession) {
                                auto p = curSession.TryGetMediaPropertiesAsync().get();
                                if (p) {
                                    std::string curTrackKey = Utf16ToUtf8(p.Title().c_str()) + "_" + Utf16ToUtf8(p.Artist().c_str());
                                    if (curTrackKey != trackKey) return; // Switched track

                                    std::string newArt = TryExtractThumbnail(p, staleArt);
                                    if (!newArt.empty()) {
                                        bool changed = false;
                                        {
                                            std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                                            if (g_trackArtMap[trackKey] != newArt) {
                                                g_trackArtMap[trackKey] = newArt;
                                                changed = true;
                                            }
                                        }
                                        if (changed) {
                                            ScheduleEmit();
                                            if (newArt.size() > 10000) {
                                                return; // Acquired high-res poster
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
                isShuffleEnabled = controls.IsShuffleEnabled();
                isRepeatEnabled = controls.IsRepeatEnabled();
            }
        }
    } catch (...) {}

    if ((title.empty() && artist.empty()) || (!canPlay && !canPause && !canGoNext && !canGoPrev)) {
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
                    // Chrome / Spotify update LastUpdatedTime every 1-5s while actively playing.
                    // Extrapolate position smoothly while fresh (< 10 seconds) to prevent stale backwards snapshots.
                    if (elapsedMs > 0.0 && elapsedMs < 10000.0) {
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
            } else {
                posMs = rawPosMs;
            }
        }
    } catch (...) {}

    std::string sessionsJson = "[]";
    std::string activeSessionId = sourceApp;
    try {
        if (g_sessionManager) {
            auto now = std::chrono::steady_clock::now();
            std::lock_guard<std::mutex> sLock(g_sessionsJsonMutex);
            auto elapsedMs = std::chrono::duration_cast<std::chrono::milliseconds>(now - g_lastSessionsFetchTime).count();

            if (elapsedMs < 2500 && g_cachedSessionsJson != "[]") {
                sessionsJson = g_cachedSessionsJson;
            } else {
                auto allSessions = g_sessionManager.GetSessions();
                std::ostringstream sJson;
                sJson << "[";
                bool first = true;
                int sIndex = 0;
                for (auto const& s : allSessions) {
                    try {
                        if (!IsValidMediaSession(s)) continue;

                        std::string sId = Utf16ToUtf8(s.SourceAppUserModelId().c_str());
                        std::string sApp = sId;
                        size_t slash = sApp.find_last_of("\\/");
                        if (slash != std::string::npos) sApp = sApp.substr(slash + 1);

                        std::string sTitle = "";
                        std::string sArtist = "";
                        try {
                            auto props = s.TryGetMediaPropertiesAsync().get();
                            if (props) {
                                sTitle = Utf16ToUtf8(props.Title().c_str());
                                sArtist = Utf16ToUtf8(props.Artist().c_str());
                            }
                        } catch (...) {}

                        if (sTitle.empty() && sArtist.empty()) continue;

                        bool sPlaying = false;
                        try {
                            auto pInfo = s.GetPlaybackInfo();
                            if (pInfo) {
                                sPlaying = (pInfo.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing);
                            }
                        } catch (...) {}

                        std::string uniqueId = sId + "::" + std::to_string(sIndex) + "::" + sTitle;
                        if (s == session) {
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
                    sIndex++;
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
