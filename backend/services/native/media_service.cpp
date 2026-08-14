#include "media_service.h"
#include "winrt_common.h"

#include <mmdeviceapi.h>
#include <endpointvolume.h>
#include <audiopolicy.h>
#include <psapi.h>
#include <winrt/Windows.Media.Control.h>
#include <algorithm>

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

static std::string g_cachedTrackKey;
static std::string g_cachedAlbumArt;
static int g_artFetchCount = 0;
static std::mutex g_artCacheMutex;

static std::mutex g_audioVolMutex;

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

static void UpdateCurrentSession() {
    std::lock_guard<std::recursive_mutex> lock(g_sessionMutex);
    if (!g_sessionManager) return;

    try {
        winrt_media::GlobalSystemMediaTransportControlsSession targetSession{nullptr};

        // 1. Get focused session
        winrt_media::GlobalSystemMediaTransportControlsSession focusedSession{nullptr};
        try {
            focusedSession = g_sessionManager.GetCurrentSession();
        } catch (...) {}

        // 2. Check if focused session is actively playing
        bool focusedIsPlaying = false;
        if (focusedSession) {
            try {
                auto info = focusedSession.GetPlaybackInfo();
                if (info && info.PlaybackStatus() == winrt_media::GlobalSystemMediaTransportControlsSessionPlaybackStatus::Playing) {
                    focusedIsPlaying = true;
                }
            } catch (...) {}
        }

        if (focusedIsPlaying) {
            targetSession = focusedSession;
        } else {
            // Focused session is not playing (or null). Check if ANY other session is actively Playing
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
            } catch (...) {}

            // If no session is actively playing, fallback to focusedSession (paused), or first available session
            if (!targetSession) {
                if (focusedSession) {
                    targetSession = focusedSession;
                } else {
                    try {
                        auto sessions = g_sessionManager.GetSessions();
                        for (auto const& s : sessions) {
                            targetSession = s;
                            break;
                        }
                    } catch (...) {}
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

    if (!session) {
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
            bool shouldFetchArt = false;
            {
                std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                if (trackKey != g_cachedTrackKey) {
                    g_cachedTrackKey = trackKey;
                    g_cachedAlbumArt = "";
                    g_artFetchCount = 0;
                }

                if (props.Thumbnail() != nullptr && (g_cachedAlbumArt.empty() || g_artFetchCount < 6)) {
                    shouldFetchArt = true;
                    g_artFetchCount++;
                } else if (!g_cachedAlbumArt.empty()) {
                    albumArt = g_cachedAlbumArt;
                }
            }

            if (shouldFetchArt && props.Thumbnail()) {
                try {
                    auto stream = props.Thumbnail().OpenReadAsync().get();
                    if (stream) {
                        std::string fetchedArt = StreamToBase64(stream);
                        if (!fetchedArt.empty()) {
                            std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                            if (g_cachedTrackKey == trackKey) {
                                g_cachedAlbumArt = fetchedArt;
                                albumArt = fetchedArt;
                            }
                        }
                    }
                } catch (...) {}
            }

            if (albumArt.empty()) {
                std::lock_guard<std::mutex> artLock(g_artCacheMutex);
                if (g_cachedTrackKey == trackKey && !g_cachedAlbumArt.empty()) {
                    albumArt = g_cachedAlbumArt;
                }
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

    try {
        auto timeline = session.GetTimelineProperties();
        if (timeline) {
            double rawPosMs = std::chrono::duration<double, std::milli>(timeline.Position()).count();
            durMs = std::chrono::duration<double, std::milli>(timeline.EndTime()).count();
            posMs = rawPosMs;

            if (isPlaying && durMs > 0) {
                auto lastUpdated = timeline.LastUpdatedTime();
                auto now = winrt::clock::now();
                if (now > lastUpdated) {
                    double elapsedSinceUpdate = std::chrono::duration<double, std::milli>(now - lastUpdated).count();
                    if (elapsedSinceUpdate > 0 && elapsedSinceUpdate < 600000) {
                        posMs = std::min(durMs, rawPosMs + elapsedSinceUpdate);
                    }
                }
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
         << "\"canPlay\":" << (canPlay ? "true" : "false") << "}";

    g_mediaCb(json.str().c_str());
}

static float g_mediaVolumeStore = 100.0f;
static bool g_mediaMuteStore = false;
static std::mutex g_mediaVolStateMutex;

static bool GetAppSessionSimpleAudioVolume(const std::string& sourceApp, ISimpleAudioVolume** ppSessionVolume) {
    if (!ppSessionVolume) return false;
    *ppSessionVolume = nullptr;

    IMMDeviceEnumerator* pEnumerator = nullptr;
    if (FAILED(CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, __uuidof(IMMDeviceEnumerator), (void**)&pEnumerator)) || !pEnumerator) return false;

    IMMDevice* pDevice = nullptr;
    HRESULT hr = pEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &pDevice);
    pEnumerator->Release();
    if (FAILED(hr) || !pDevice) return false;

    IAudioSessionManager2* pSessionManager2 = nullptr;
    hr = pDevice->Activate(__uuidof(IAudioSessionManager2), CLSCTX_ALL, NULL, (void**)&pSessionManager2);
    pDevice->Release();
    if (FAILED(hr) || !pSessionManager2) return false;

    IAudioSessionEnumerator* pSessionEnum = nullptr;
    hr = pSessionManager2->GetSessionEnumerator(&pSessionEnum);
    pSessionManager2->Release();
    if (FAILED(hr) || !pSessionEnum) return false;

    int count = 0;
    pSessionEnum->GetCount(&count);

    ISimpleAudioVolume* pMatchedVolume = nullptr;
    ISimpleAudioVolume* pFallbackActiveVolume = nullptr;

    std::string appLower = sourceApp;
    std::transform(appLower.begin(), appLower.end(), appLower.begin(), ::tolower);

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

                bool isMatch = false;

                // 1. Match Session Identifier & Display Name
                LPWSTR pswSessionId = nullptr;
                if (SUCCEEDED(pControl2->GetSessionIdentifier(&pswSessionId)) && pswSessionId) {
                    std::string sessId = Utf16ToUtf8(pswSessionId);
                    CoTaskMemFree(pswSessionId);
                    std::transform(sessId.begin(), sessId.end(), sessId.begin(), ::tolower);
                    if (!appLower.empty() && sessId.find(appLower) != std::string::npos) {
                        isMatch = true;
                    }
                }

                if (!isMatch) {
                    LPWSTR pswDisplayName = nullptr;
                    if (SUCCEEDED(pControl2->GetDisplayName(&pswDisplayName)) && pswDisplayName) {
                        std::string dispName = Utf16ToUtf8(pswDisplayName);
                        CoTaskMemFree(pswDisplayName);
                        std::transform(dispName.begin(), dispName.end(), dispName.begin(), ::tolower);
                        if (!appLower.empty() && !dispName.empty() && (appLower.find(dispName) != std::string::npos || dispName.find(appLower) != std::string::npos)) {
                            isMatch = true;
                        }
                    }
                }

                // 2. Match Process Executable Name
                if (!isMatch) {
                    DWORD pid = 0;
                    pControl2->GetProcessId(&pid);
                    if (pid != 0) {
                        HANDLE hProcess = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, FALSE, pid);
                        if (hProcess) {
                            char procPath[MAX_PATH] = {0};
                            DWORD dwSize = MAX_PATH;
                            if (QueryFullProcessImageNameA(hProcess, 0, procPath, &dwSize)) {
                                std::string procName = procPath;
                                size_t lastSlash = procName.find_last_of("\\/");
                                if (lastSlash != std::string::npos) procName = procName.substr(lastSlash + 1);
                                std::transform(procName.begin(), procName.end(), procName.begin(), ::tolower);
                                size_t pExe = procName.find(".exe");
                                if (pExe != std::string::npos) procName = procName.substr(0, pExe);

                                if (!appLower.empty() && !procName.empty()) {
                                    if (appLower.find(procName) != std::string::npos || procName.find(appLower) != std::string::npos) {
                                        isMatch = true;
                                    }
                                }
                            }
                            CloseHandle(hProcess);
                        }
                    }
                }

                if (isMatch) {
                    pControl2->QueryInterface(__uuidof(ISimpleAudioVolume), (void**)&pMatchedVolume);
                    pControl2->Release();
                    pSessionControl->Release();
                    break;
                }

                if (!pFallbackActiveVolume) {
                    AudioSessionState state;
                    if (SUCCEEDED(pControl2->GetState(&state)) && state == AudioSessionStateActive) {
                        pControl2->QueryInterface(__uuidof(ISimpleAudioVolume), (void**)&pFallbackActiveVolume);
                    }
                }

                pControl2->Release();
            }
            pSessionControl->Release();
        }
    }

    pSessionEnum->Release();

    if (pMatchedVolume) {
        if (pFallbackActiveVolume) pFallbackActiveVolume->Release();
        *ppSessionVolume = pMatchedVolume;
        return true;
    }

    if (pFallbackActiveVolume) {
        *ppSessionVolume = pFallbackActiveVolume;
        return true;
    }

    return false;
}

static std::pair<float, bool> GetMediaAudioVolumeAndMute(const std::string& sourceApp) {
    std::lock_guard<std::mutex> lock(g_audioVolMutex);
    ComInitScope com;

    ISimpleAudioVolume* pAppVolume = nullptr;
    if (GetAppSessionSimpleAudioVolume(sourceApp, &pAppVolume) && pAppVolume) {
        float vol = 1.0f;
        BOOL muted = FALSE;
        pAppVolume->GetMasterVolume(&vol);
        pAppVolume->GetMute(&muted);
        pAppVolume->Release();
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
    }

    std::lock_guard<std::mutex> lock(g_audioVolMutex);
    ComInitScope com;

    ISimpleAudioVolume* pAppVolume = nullptr;
    if (GetAppSessionSimpleAudioVolume(sourceApp, &pAppVolume) && pAppVolume) {
        pAppVolume->SetMasterVolume(clamped / 100.0f, NULL);
        pAppVolume->Release();
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

    ISimpleAudioVolume* pAppVolume = nullptr;
    if (GetAppSessionSimpleAudioVolume(sourceApp, &pAppVolume) && pAppVolume) {
        pAppVolume->SetMute(nextMute ? TRUE : FALSE, NULL);
        pAppVolume->Release();
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
                    ScheduleEmit();
                });
            }
        } catch (...) {}

        UpdateCurrentSession();

        while (g_mediaRunning) {
            std::this_thread::sleep_for(std::chrono::milliseconds(500));
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
        float vol = std::stof(cmd.substr(11));
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
    if (cmd == "mute") {
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
