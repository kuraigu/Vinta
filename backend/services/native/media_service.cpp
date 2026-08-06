#include "media_service.h"
#include "winrt_common.h"

#include <mmdeviceapi.h>
#include <endpointvolume.h>
#include <winrt/Windows.Media.Control.h>

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
static std::mutex g_artCacheMutex;

// Cached Audio Endpoint Volume COM interfaces to avoid audiosrv RPC thrashing
static IAudioEndpointVolume* g_pCachedEndpointVolume = nullptr;
static IMMDevice* g_pCachedAudioDevice = nullptr;
static IMMDeviceEnumerator* g_pCachedEnumerator = nullptr;
static std::mutex g_audioVolMutex;

static void EnsureAudioEndpointVolume() {
    if (g_pCachedEndpointVolume) return;
    if (!g_pCachedEnumerator) {
        HRESULT hr = CoCreateInstance(__uuidof(MMDeviceEnumerator), NULL, CLSCTX_ALL, __uuidof(IMMDeviceEnumerator), (void**)&g_pCachedEnumerator);
        if (FAILED(hr) || !g_pCachedEnumerator) return;
    }
    if (!g_pCachedAudioDevice) {
        HRESULT hr = g_pCachedEnumerator->GetDefaultAudioEndpoint(eRender, eMultimedia, &g_pCachedAudioDevice);
        if (FAILED(hr) || !g_pCachedAudioDevice) return;
    }
    g_pCachedAudioDevice->Activate(__uuidof(IAudioEndpointVolume), CLSCTX_ALL, NULL, (void**)&g_pCachedEndpointVolume);
}

static void ReleaseAudioEndpointVolume() {
    if (g_pCachedEndpointVolume) { g_pCachedEndpointVolume->Release(); g_pCachedEndpointVolume = nullptr; }
    if (g_pCachedAudioDevice) { g_pCachedAudioDevice->Release(); g_pCachedAudioDevice = nullptr; }
    if (g_pCachedEnumerator) { g_pCachedEnumerator->Release(); g_pCachedEnumerator = nullptr; }
}

static void UpdateCurrentSession();
static void EmitMediaState(bool includeVolume = true);
static void ScheduleEmit();
static std::pair<float, bool> GetMasterAudioVolumeAndMute();
static void SetMasterAudioVolume(float percent);
static void ToggleMasterAudioMute();
static void SendVolumeKey(BYTE vk);

bool IsMediaPlayingState() {
    return g_isMediaPlaying.load();
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

static std::pair<float, bool> GetMasterAudioVolumeAndMute() {
    float vol = 50.0f;
    bool muted = false;

    std::lock_guard<std::mutex> lock(g_audioVolMutex);
    ComInitScope com;
    EnsureAudioEndpointVolume();

    if (g_pCachedEndpointVolume) {
        float scalar = 0.5f;
        BOOL isMuted = FALSE;
        if (FAILED(g_pCachedEndpointVolume->GetMasterVolumeLevelScalar(&scalar))) {
            ReleaseAudioEndpointVolume();
            EnsureAudioEndpointVolume();
            if (g_pCachedEndpointVolume) {
                g_pCachedEndpointVolume->GetMasterVolumeLevelScalar(&scalar);
            }
        }
        vol = scalar * 100.0f;
        if (g_pCachedEndpointVolume && SUCCEEDED(g_pCachedEndpointVolume->GetMute(&isMuted))) {
            muted = (isMuted == TRUE);
        }
    }

    return {vol, muted};
}

static void SetMasterAudioVolume(float percent) {
    float clamped = std::max(0.0f, std::min(100.0f, percent));
    std::lock_guard<std::mutex> lock(g_audioVolMutex);
    ComInitScope com;
    EnsureAudioEndpointVolume();
    if (g_pCachedEndpointVolume) {
        g_pCachedEndpointVolume->SetMasterVolumeLevelScalar(clamped / 100.0f, NULL);
    }
}

static void ToggleMasterAudioMute() {
    std::lock_guard<std::mutex> lock(g_audioVolMutex);
    ComInitScope com;
    EnsureAudioEndpointVolume();
    if (g_pCachedEndpointVolume) {
        BOOL isMuted = FALSE;
        if (SUCCEEDED(g_pCachedEndpointVolume->GetMute(&isMuted))) {
            g_pCachedEndpointVolume->SetMute(!isMuted, NULL);
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

    {
        std::lock_guard<std::mutex> lock(g_audioVolMutex);
        ReleaseAudioEndpointVolume();
    }
}

int ExecuteMediaCommand(const char* command) {
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
