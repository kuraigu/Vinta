using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net.Http;
using System.Runtime.InteropServices;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading;
using System.Threading.Tasks;
using System.Xml.Linq;
using Microsoft.Data.Sqlite;
using Windows.Media.Control;
using Windows.Storage.Streams;
using Windows.UI.Notifications;
using Windows.UI.Notifications.Management;

Console.OutputEncoding = Encoding.UTF8;
SQLitePCL.Batteries.Init();

GlobalSystemMediaTransportControlsSessionManager? gsmtcManager = null;
GlobalSystemMediaTransportControlsSession? currentMediaSession = null;

// Shared HttpClient (singleton - avoids socket exhaustion from repeated creation/disposal)
var sharedHttpClient = new HttpClient { Timeout = TimeSpan.FromSeconds(3) };

// Debounce: prevent EmitMediaStateAsync from firing more than once per 150ms
// (TimelinePropertiesChanged can fire many times per second during playback)
int _emitPending = 0;

// Volume cache: only re-query COM interfaces when actually needed, not on every position tick
float _cachedVolume = 50f;
bool _cachedMute = false;
DateTime _volumeCacheExpiry = DateTime.MinValue;
const int VolumeCacheTtlMs = 500;

// App icon cache: avoid repeated SQLite queries for the same app
var _iconCache = new Dictionary<string, string>();

// Start Notification Listener Task
_ = Task.Run(StartNotificationListener);

// Start GSMTC Media Listener & Command Loop Task
_ = Task.Run(StartMediaListenerAndCommandLoop);

// Keep main thread alive
await Task.Delay(-1);

#region Notification Listener
async Task StartNotificationListener()
{
    try
    {
        var listener = UserNotificationListener.Current;
        var accessStatus = await listener.RequestAccessAsync();

        if (accessStatus != UserNotificationListenerAccessStatus.Allowed)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { type = "error", error = "Access denied" }));
            return;
        }

        var seenIds = new HashSet<uint>();
        var initialNotifs = await listener.GetNotificationsAsync(NotificationKinds.Toast);
        foreach (var notif in initialNotifs)
        {
            seenIds.Add(notif.Id);
        }

        Console.WriteLine(JsonSerializer.Serialize(new { type = "status", status = "Listening" }));

        while (true)
        {
            try
            {
                var notifs = await listener.GetNotificationsAsync(NotificationKinds.Toast);
                foreach (var notif in notifs)
                {
                    if (!seenIds.Contains(notif.Id))
                    {
                        if (await ProcessNotificationAsync(notif))
                        {
                            seenIds.Add(notif.Id);
                        }
                    }
                }

                // Prune seenIds when it grows stale (keeps RAM tight)
                if (seenIds.Count > notifs.Count + 50)
                {
                    var currentIds = new HashSet<uint>(notifs.Select(n => n.Id));
                    seenIds.IntersectWith(currentIds);
                }
            }
            catch
            {
                // Ignore transient listener errors
            }

            // 500ms poll — notifications are not time-critical to the millisecond
            await Task.Delay(500);
        }
    }
    catch (Exception ex)
    {
        Console.WriteLine(JsonSerializer.Serialize(new { type = "error", error = ex.Message }));
    }
}

async Task<bool> ProcessNotificationAsync(UserNotification notif)
{
    try
    {
        string appName = "";
        try { appName = notif.AppInfo?.DisplayInfo?.DisplayName ?? ""; } catch { }
        if (string.IsNullOrEmpty(appName) || appName == "App")
        {
            try { appName = notif.AppInfo?.AppUserModelId ?? ""; } catch { }
        }

        string title = "";
        string body = "";
        string heroImage = "";
        string icon = "";

        // Parse notification text elements
        try
        {
            if (notif.Notification?.Visual?.Bindings != null)
            {
                foreach (var binding in notif.Notification.Visual.Bindings)
                {
                    var texts = binding.GetTextElements();
                    if (texts != null && texts.Count > 0)
                    {
                        if (string.IsNullOrEmpty(title)) title = texts[0].Text ?? "";
                        if (texts.Count > 1 && string.IsNullOrEmpty(body)) body = texts[1].Text ?? "";

                        for (int i = 2; i < texts.Count; i++)
                        {
                            if (!string.IsNullOrEmpty(texts[i].Text))
                            {
                                body += (string.IsNullOrEmpty(body) ? "" : "\n") + texts[i].Text;
                            }
                        }
                    }
                }
            }
        }
        catch { }

        // Attempt SQLite wpndatabase.db extraction for hero images & notification appLogoOverrides
        try
        {
            var (extractedHero, extractedIcon) = ExtractNotificationImages(notif.Id);
            if (!string.IsNullOrEmpty(extractedHero)) heroImage = extractedHero;
            if (!string.IsNullOrEmpty(extractedIcon)) icon = extractedIcon;
        }
        catch { }

        // Attempt direct WinRT logo stream extraction if icon was not found in wpndatabase.db
        if (string.IsNullOrEmpty(icon))
        {
            try
            {
                if (notif.AppInfo?.DisplayInfo != null)
                {
                    var logoStreamRef = notif.AppInfo.DisplayInfo.GetLogo(new Windows.Foundation.Size(48, 48));
                    if (logoStreamRef != null)
                    {
                        using var stream = await logoStreamRef.OpenReadAsync();
                        using var netStream = stream.AsStreamForRead();
                        using var ms = new MemoryStream();
                        await netStream.CopyToAsync(ms);
                        var bytes = ms.ToArray();
                        if (bytes.Length > 0)
                        {
                            icon = "data:image/png;base64," + Convert.ToBase64String(bytes);
                        }
                    }
                }
            }
            catch { }
        }

        // Fallback to SQLite AppCache lookup if icon was still empty
        if (string.IsNullOrEmpty(icon))
        {
            try
            {
                var aumid = notif.AppInfo?.AppUserModelId ?? "";
                icon = ExtractAppIcon(aumid);
            }
            catch { }
        }

        var payload = new
        {
            type = "notification",
            appName,
            icon,
            title,
            message = body,
            heroImage
        };

        Console.WriteLine(JsonSerializer.Serialize(payload));
        return true;
    }
    catch
    {
        return false;
    }
}

(string heroImage, string icon) ExtractNotificationImages(uint notifId)
{
    string heroImage = "";
    string icon = "";

    try
    {
        var localAppData = Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData);
        var dbPath = Path.Combine(localAppData, @"Microsoft\Windows\Notifications\wpndatabase.db");
        if (!File.Exists(dbPath)) return ("", "");

        var tempPath = Path.Combine(Path.GetTempPath(), $"vinta_wpn_{Guid.NewGuid():N}.db");
        File.Copy(dbPath, tempPath, true);

        try
        {
            using var connection = new SqliteConnection($"Data Source={tempPath};Mode=ReadOnly;Cache=Shared;");
            connection.Open();

            using var command = connection.CreateCommand();
            command.CommandText = "SELECT Payload FROM Notification WHERE Id = @Id";
            command.Parameters.AddWithValue("@Id", notifId);

            using var reader = command.ExecuteReader();
            if (reader.Read())
            {
                byte[] payloadBytes = (byte[])reader["Payload"];
                string xmlPayload = Encoding.UTF8.GetString(payloadBytes);

                var doc = XDocument.Parse(xmlPayload);
                string heroUrl = "";
                string iconUrl = "";

                foreach (var img in doc.Descendants().Where(x => x.Name.LocalName == "image"))
                {
                    string placement = img.Attribute("placement")?.Value ?? "";
                    string src = img.Attribute("src")?.Value ?? "";
                    if (placement == "appLogoOverride")
                    {
                        iconUrl = src;
                    }
                    else if (placement == "hero" || string.IsNullOrEmpty(placement))
                    {
                        if (string.IsNullOrEmpty(heroUrl)) heroUrl = src;
                    }
                }

                heroImage = DownloadImageAsBase64(heroUrl);
                icon = DownloadImageAsBase64(iconUrl);
            }
        }
        finally
        {
            try { if (File.Exists(tempPath)) File.Delete(tempPath); } catch { }
        }
    }
    catch { }

    return (heroImage, icon);
}

string DownloadImageAsBase64(string imgUrl)
{
    if (string.IsNullOrEmpty(imgUrl)) return "";
    try
    {
        if (imgUrl.StartsWith("http", StringComparison.OrdinalIgnoreCase))
        {
            // Reuse shared HttpClient — no socket exhaustion
            var response = sharedHttpClient.GetAsync(imgUrl).Result;
            if (!response.IsSuccessStatusCode) return "";

            if (response.Content.Headers.ContentLength > 2 * 1024 * 1024) return "";

            var imgBytes = response.Content.ReadAsByteArrayAsync().Result;
            if (imgBytes.Length > 2 * 1024 * 1024) return "";

            return "data:image/jpeg;base64," + Convert.ToBase64String(imgBytes);
        }
        else if (imgUrl.StartsWith("ms-appdata://", StringComparison.OrdinalIgnoreCase) || imgUrl.StartsWith("ms-appx://", StringComparison.OrdinalIgnoreCase))
        {
            string localAppData = Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData);
            string packagesPath = Path.Combine(localAppData, "Packages");
            string tempPath = Path.GetTempPath();
            string inetCachePath = Path.Combine(localAppData, @"Microsoft\Windows\INetCache");
            string fileName = Path.GetFileName(imgUrl);

            if (!string.IsNullOrEmpty(fileName))
            {
                string localTempFile = Path.Combine(tempPath, fileName);
                if (File.Exists(localTempFile))
                {
                    var imgBytes = File.ReadAllBytes(localTempFile);
                    return "data:image/png;base64," + Convert.ToBase64String(imgBytes);
                }

                if (Directory.Exists(inetCachePath))
                {
                    // Limit search to top-level + 1 depth to avoid full recursive scan
                    var matches = Directory.GetFiles(inetCachePath, fileName, SearchOption.AllDirectories);
                    if (matches.Length > 0 && File.Exists(matches[0]))
                    {
                        var imgBytes = File.ReadAllBytes(matches[0]);
                        return "data:image/png;base64," + Convert.ToBase64String(imgBytes);
                    }
                }

                // Constrain Packages search: limit recursion depth to avoid scanning gigabytes of package data
                if (Directory.Exists(packagesPath))
                {
                    try
                    {
                        var enumOptions = new EnumerationOptions
                        {
                            RecurseSubdirectories = true,
                            MaxRecursionDepth = 5,
                            IgnoreInaccessible = true
                        };
                        var match = Directory
                            .EnumerateFiles(packagesPath, fileName, enumOptions)
                            .FirstOrDefault();
                        if (match != null && File.Exists(match))
                        {
                            var fileInfo = new FileInfo(match);
                            if (fileInfo.Length <= 2 * 1024 * 1024)
                            {
                                var imgBytes = File.ReadAllBytes(match);
                                return "data:image/png;base64," + Convert.ToBase64String(imgBytes);
                            }
                        }
                    }
                    catch { }
                }
            }
            return "";
        }
        else if (imgUrl.StartsWith("file:///", StringComparison.OrdinalIgnoreCase))
        {
            string localPath = Uri.UnescapeDataString(imgUrl.Substring(8)).Replace('/', '\\');
            if (!File.Exists(localPath)) return "";
            var fileInfo = new FileInfo(localPath);
            if (fileInfo.Length > 2 * 1024 * 1024) return "";

            var imgBytes = File.ReadAllBytes(localPath);
            return "data:image/jpeg;base64," + Convert.ToBase64String(imgBytes);
        }
        else if (File.Exists(imgUrl))
        {
            var fileInfo = new FileInfo(imgUrl);
            if (fileInfo.Length <= 2 * 1024 * 1024)
            {
                var imgBytes = File.ReadAllBytes(imgUrl);
                return "data:image/png;base64," + Convert.ToBase64String(imgBytes);
            }
        }
    }
    catch { }
    return "";
}


string ExtractAppIcon(string aumid)
{
    if (string.IsNullOrEmpty(aumid)) return "";

    if (_iconCache.TryGetValue(aumid, out var cached)) return cached;

    try
    {
        var localAppData = Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData);
        var dbPath = Path.Combine(localAppData, @"Packages\Microsoft.Windows.Search_cw5n1h2txyewy\LocalState\DeviceSearchCache\AppCache.db");
        if (!File.Exists(dbPath)) return "";

        using var conn = new SqliteConnection($"Data Source={dbPath};Mode=ReadOnly;Cache=Shared;");
        conn.Open();

        using var cmd = conn.CreateCommand();
        cmd.CommandText = "SELECT Logo FROM Apps WHERE AUMID = @aumid LIMIT 1";
        cmd.Parameters.AddWithValue("@aumid", aumid);

        var result = cmd.ExecuteScalar();
        if (result != null && result != DBNull.Value)
        {
            var logoPath = result.ToString() ?? "";
            if (!string.IsNullOrEmpty(logoPath) && File.Exists(logoPath))
            {
                var bytes = File.ReadAllBytes(logoPath);
                var icon = "data:image/png;base64," + Convert.ToBase64String(bytes);
                _iconCache[aumid] = icon; // Cache hit for future notifications from same app
                return icon;
            }
        }
    }
    catch { }

    _iconCache[aumid] = ""; // Cache miss to avoid repeated DB hits
    return "";
}
#endregion

#region GSMTC Media Listener & Command Loop

async Task StartMediaListenerAndCommandLoop()
{
    try
    {
        gsmtcManager = await GlobalSystemMediaTransportControlsSessionManager.RequestAsync();
        if (gsmtcManager != null)
        {
            gsmtcManager.CurrentSessionChanged += (s, e) => UpdateCurrentSession();
            gsmtcManager.SessionsChanged += (s, e) => UpdateCurrentSession();
            UpdateCurrentSession();
        }
    }
    catch { }

    // Reduced polling: GSMTC events cover most cases.
    // 3000ms fallback poll is sufficient for edge cases (session not firing events).
    _ = Task.Run(async () =>
    {
        while (true)
        {
            await Task.Delay(3000);
            try
            {
                if (gsmtcManager != null)
                {
                    UpdateCurrentSession();
                }
            }
            catch { }
        }
    });

    // IPC Command Reader Loop reading stdin
    while (true)
    {
        try
        {
            var line = await Console.In.ReadLineAsync();
            if (line == null)
            {
                // Stdin pipe closed by parent Vinta.exe -> terminate immediately!
                Environment.Exit(0);
                break;
            }
            line = line.Trim();

            if (line.StartsWith("COMMAND:"))
            {
                var action = line.Substring("COMMAND:".Length).Trim().ToLowerInvariant();
                await HandleMediaCommandAsync(action);
            }
        }
        catch { }
    }
}

void UpdateCurrentSession()
{
    if (gsmtcManager == null) return;
    try
    {
        var targetSession = gsmtcManager.GetCurrentSession();

        // Fallback: If GetCurrentSession() returns null, search GetSessions() for an active playing session
        if (targetSession == null)
        {
            var sessions = gsmtcManager.GetSessions();
            if (sessions != null && sessions.Count > 0)
            {
                targetSession = sessions.FirstOrDefault(s => s.GetPlaybackInfo()?.PlaybackStatus == GlobalSystemMediaTransportControlsSessionPlaybackStatus.Playing)
                                ?? sessions.FirstOrDefault();
            }
        }

        if (currentMediaSession != targetSession)
        {
            if (currentMediaSession != null)
            {
                try { currentMediaSession.MediaPropertiesChanged -= OnMediaPropertiesChanged; } catch { }
                try { currentMediaSession.PlaybackInfoChanged -= OnPlaybackInfoChanged; } catch { }
                try { currentMediaSession.TimelinePropertiesChanged -= OnTimelinePropertiesChanged; } catch { }
            }

            currentMediaSession = targetSession;

            if (currentMediaSession != null)
            {
                currentMediaSession.MediaPropertiesChanged += OnMediaPropertiesChanged;
                currentMediaSession.PlaybackInfoChanged += OnPlaybackInfoChanged;
                currentMediaSession.TimelinePropertiesChanged += OnTimelinePropertiesChanged;
            }
        }

        _ = EmitMediaStateAsync(includeVolume: true);
    }
    catch { }
}

void OnMediaPropertiesChanged(GlobalSystemMediaTransportControlsSession sender, MediaPropertiesChangedEventArgs args)
    => ScheduleEmit(includeVolume: false);

void OnPlaybackInfoChanged(GlobalSystemMediaTransportControlsSession sender, PlaybackInfoChangedEventArgs args)
    => ScheduleEmit(includeVolume: true);

void OnTimelinePropertiesChanged(GlobalSystemMediaTransportControlsSession sender, TimelinePropertiesChangedEventArgs args)
    => ScheduleEmit(includeVolume: false); // Position ticks don't need volume re-query

/// <summary>
/// Debounced emit: coalesces rapid-fire events (e.g. many TimelinePropertiesChanged per second)
/// into a single actual emission after 150ms, preventing CPU/RAM spike during playback.
/// </summary>
void ScheduleEmit(bool includeVolume)
{
    if (Interlocked.CompareExchange(ref _emitPending, 1, 0) == 0)
    {
        _ = Task.Run(async () =>
        {
            await Task.Delay(150);
            Interlocked.Exchange(ref _emitPending, 0);
            await EmitMediaStateAsync(includeVolume);
        });
    }
}

async Task EmitMediaStateAsync(bool includeVolume = true)
{
    if (currentMediaSession == null)
    {
        Console.WriteLine(JsonSerializer.Serialize(new { type = "media", isAvailable = false }));
        return;
    }

    try
    {
        GlobalSystemMediaTransportControlsSessionMediaProperties? mediaProps = null;
        GlobalSystemMediaTransportControlsSessionPlaybackInfo? playbackInfo = null;
        GlobalSystemMediaTransportControlsSessionTimelineProperties? timelineProps = null;

        try { mediaProps = await currentMediaSession.TryGetMediaPropertiesAsync(); } catch { }
        try { playbackInfo = currentMediaSession.GetPlaybackInfo(); } catch { }
        try { timelineProps = currentMediaSession.GetTimelineProperties(); } catch { }

        string albumArtBase64 = "";
        if (mediaProps?.Thumbnail != null)
        {
            try
            {
                using var stream = await mediaProps.Thumbnail.OpenReadAsync();
                using var netStream = stream.AsStreamForRead();
                using var ms = new MemoryStream();
                await netStream.CopyToAsync(ms);
                var bytes = ms.ToArray();
                if (bytes.Length > 0)
                {
                    albumArtBase64 = "data:image/png;base64," + Convert.ToBase64String(bytes);
                }
            }
            catch { }
        }

        var isPlaying = playbackInfo?.PlaybackStatus == GlobalSystemMediaTransportControlsSessionPlaybackStatus.Playing;
        var controls = playbackInfo?.Controls;

        // Only query COM volume interfaces when needed (not on every position tick)
        if (includeVolume || DateTime.UtcNow > _volumeCacheExpiry)
        {
            var (vol, muted) = GetMasterVolume();
            _cachedVolume = vol;
            _cachedMute = muted;
            _volumeCacheExpiry = DateTime.UtcNow.AddMilliseconds(VolumeCacheTtlMs);
        }

        var payload = new
        {
            type = "media",
            isAvailable = true,
            sourceApp = currentMediaSession?.SourceAppUserModelId ?? "",
            title = mediaProps?.Title ?? "",
            artist = mediaProps?.Artist ?? "",
            album = mediaProps?.AlbumTitle ?? "",
            albumArt = albumArtBase64,
            isPlaying,
            positionMs = timelineProps?.Position.TotalMilliseconds ?? 0,
            durationMs = timelineProps?.EndTime.TotalMilliseconds ?? 0,
            volume = (int)Math.Round(_cachedVolume),
            isMuted = _cachedMute,
            canGoNext = controls?.IsNextEnabled ?? true,
            canGoPrev = controls?.IsPreviousEnabled ?? true,
            canPause = controls?.IsPauseEnabled ?? true,
            canPlay = controls?.IsPlayEnabled ?? true
        };

        Console.WriteLine(JsonSerializer.Serialize(payload));
    }
    catch { }
}

(float volume, bool isMuted) GetMasterVolume()
{
    try
    {
        var enumerator = (IMMDeviceEnumerator)new MMDeviceEnumerator();
        IMMDevice? dev = null;
        int hr = enumerator.GetDefaultAudioEndpoint(0, 1, out dev);
        if (hr != 0 || dev == null)
        {
            enumerator.GetDefaultAudioEndpoint(0, 0, out dev);
        }

        if (dev != null)
        {
            Guid iid = typeof(IAudioEndpointVolume).GUID;
            dev.Activate(ref iid, 23, IntPtr.Zero, out object epvObj);
            var epv = (IAudioEndpointVolume)epvObj;
            epv.GetMasterVolumeLevelScalar(out float vol);
            epv.GetMute(out bool mute);
            return (vol * 100f, mute);
        }
    }
    catch { }
    return (50f, false);
}

void SetMasterVolume(float percent)
{
    try
    {
        var clamped = Math.Clamp(percent, 0f, 100f);
        _cachedVolume = clamped;
        _volumeCacheExpiry = DateTime.UtcNow.AddMilliseconds(VolumeCacheTtlMs);

        var enumerator = (IMMDeviceEnumerator)new MMDeviceEnumerator();
        IMMDevice? dev = null;
        int hr = enumerator.GetDefaultAudioEndpoint(0, 1, out dev);
        if (hr != 0 || dev == null)
        {
            enumerator.GetDefaultAudioEndpoint(0, 0, out dev);
        }

        if (dev != null)
        {
            Guid iid = typeof(IAudioEndpointVolume).GUID;
            dev.Activate(ref iid, 23, IntPtr.Zero, out object epvObj);
            var epv = (IAudioEndpointVolume)epvObj;
            Guid zero = Guid.Empty;
            epv.SetMasterVolumeLevelScalar(clamped / 100f, ref zero);
        }
    }
    catch { }
}

const byte VK_VOLUME_MUTE = 0xAD;
const byte VK_VOLUME_DOWN = 0xAE;
const byte VK_VOLUME_UP = 0xAF;
const byte VK_MEDIA_NEXT = 0xB0;
const byte VK_MEDIA_PREV = 0xB1;
const byte VK_MEDIA_STOP = 0xB2;
const byte VK_MEDIA_PLAY_PAUSE = 0xB3;
const uint KEYEVENTF_KEYUP = 0x0002;

void SendVolumeKey(byte vk)
{
    try
    {
        keybd_event(vk, 0, 0, 0);
        keybd_event(vk, 0, KEYEVENTF_KEYUP, 0);
        // Invalidate cache so next emit picks up new value
        _volumeCacheExpiry = DateTime.MinValue;
    }
    catch { }
}

async Task HandleMediaCommandAsync(string action)
{
    if (action.StartsWith("set_volume:"))
    {
        string valStr = action.Substring("set_volume:".Length);
        if (float.TryParse(valStr, System.Globalization.NumberStyles.Float, System.Globalization.CultureInfo.InvariantCulture, out float volPercent))
        {
            SetMasterVolume(volPercent);
            await EmitMediaStateAsync(includeVolume: true);
        }
        return;
    }
    if (action == "volume_up")
    {
        SendVolumeKey(VK_VOLUME_UP);
        await EmitMediaStateAsync(includeVolume: true);
        return;
    }
    if (action == "volume_down")
    {
        SendVolumeKey(VK_VOLUME_DOWN);
        await EmitMediaStateAsync(includeVolume: true);
        return;
    }
    if (action == "mute")
    {
        SendVolumeKey(VK_VOLUME_MUTE);
        await EmitMediaStateAsync(includeVolume: true);
        return;
    }
    if (action.StartsWith("seek:") && currentMediaSession != null)
    {
        if (double.TryParse(action.Substring("seek:".Length), System.Globalization.NumberStyles.Float, System.Globalization.CultureInfo.InvariantCulture, out double posMs))
        {
            try { await currentMediaSession.TryChangePlaybackPositionAsync((long)(posMs * 10000)); } catch { }
        }
        return;
    }

    try
    {
        bool success = false;
        if (currentMediaSession != null)
        {
            switch (action)
            {
                case "play_pause":
                    success = await currentMediaSession.TryTogglePlayPauseAsync();
                    break;
                case "play":
                    success = await currentMediaSession.TryPlayAsync();
                    break;
                case "pause":
                    success = await currentMediaSession.TryPauseAsync();
                    break;
                case "next":
                    success = await currentMediaSession.TrySkipNextAsync();
                    break;
                case "previous":
                case "prev":
                    success = await currentMediaSession.TrySkipPreviousAsync();
                    break;
            }
        }

        // Fallback: If GSMTC command was not handled or failed, send universal Win32 media key
        if (!success)
        {
            switch (action)
            {
                case "play_pause":
                case "play":
                case "pause":
                    SendVolumeKey(VK_MEDIA_PLAY_PAUSE);
                    break;
                case "next":
                    SendVolumeKey(VK_MEDIA_NEXT);
                    break;
                case "previous":
                case "prev":
                    SendVolumeKey(VK_MEDIA_PREV);
                    break;
            }
        }
    }
    catch { }
}
#endregion

#region COM Interface Declarations
[DllImport("user32.dll")]
static extern void keybd_event(byte bVk, byte bScan, uint dwFlags, uint dwExtraInfo);

[ComImport]
[Guid("BCDE0395-E52F-467C-8E3D-C4579291692E")]
class MMDeviceEnumerator { }

[Guid("A95664D2-9614-4F35-A746-DE8DB63617E6"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IMMDeviceEnumerator
{
    int EnumAudioEndpoints(int dataFlow, int stateMask, out object devices);
    int GetDefaultAudioEndpoint(int dataFlow, int role, out IMMDevice device);
}

[Guid("D6660630-1587-4E43-81F1-B948E807363F"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IMMDevice
{
    int Activate(ref Guid iid, int dwClsCtx, IntPtr pActivationParams, [MarshalAs(UnmanagedType.IUnknown)] out object ppInterface);
}

[Guid("5CDF2C82-841E-4546-9722-0CF74078229A"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IAudioEndpointVolume
{
    int RegisterControlChangeNotify(IntPtr pNotify);
    int UnregisterControlChangeNotify(IntPtr pNotify);
    int GetChannelCount(out uint pnChannelCount);
    int SetMasterVolumeLevel(float fLevelDB, ref Guid pguidEventContext);
    int SetMasterVolumeLevelScalar(float fLevel, ref Guid pguidEventContext);
    int GetMasterVolumeLevel(out float pfLevelDB);
    int GetMasterVolumeLevelScalar(out float pfLevel);
    int SetMute([MarshalAs(UnmanagedType.Bool)] bool bMute, ref Guid pguidEventContext);
    int GetMute([MarshalAs(UnmanagedType.Bool)] out bool pbMute);
}
#endregion
