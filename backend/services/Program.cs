using System.Net.Http;
using System.Runtime.InteropServices;
using System.Text;
using System.Text.Json;
using System.Xml.Linq;
using Microsoft.Data.Sqlite;
using Windows.Media.Control;

#nullable enable

Console.OutputEncoding = Encoding.UTF8;
SQLitePCL.Batteries.Init();

GlobalSystemMediaTransportControlsSessionManager? gsmtcManager = null;
GlobalSystemMediaTransportControlsSession? currentMediaSession = null;

var sharedHttpClient = new HttpClient { Timeout = TimeSpan.FromSeconds(3) };

int _emitPending = 0;
bool _isMediaPlaying = false;

float _cachedVolume = 50f;
bool _cachedMute = false;
DateTime _volumeCacheExpiry = DateTime.MinValue;
const int VolumeCacheTtlMs = 300;

var _iconCache = new Dictionary<string, string>();
var _artCache = new Dictionary<string, string>();

string _vintaCacheDir = Path.Combine(Path.GetTempPath(), "vinta_img_cache");

try {
    if (Directory.Exists(_vintaCacheDir)) Directory.Delete(_vintaCacheDir, true);
    Directory.CreateDirectory(_vintaCacheDir);
} catch { }

const byte VK_VOLUME_MUTE = 0xAD;
const byte VK_VOLUME_DOWN = 0xAE;
const byte VK_VOLUME_UP = 0xAF;

// Periodic memory optimization task: free LOH base64 strings & unmanaged COM wrappers
_ = Task.Run(async () =>
{
    while (true)
    {
        await Task.Delay(30000);
        try
        {
            if (_artCache.Count > 4) _artCache.Clear();
            if (_iconCache.Count > 15) _iconCache.Clear();
            GC.Collect(2, GCCollectionMode.Optimized, false);
        }
        catch { }
    }
});

_ = Task.Run(async () =>
{
    NativeMethods.CoInitializeEx(IntPtr.Zero, 0x2);
    var meters = new List<IAudioMeterInformation>();
    DateTime lastDeviceCheck = DateTime.MinValue;
    
    try
    {
        var enumerator = (IMMDeviceEnumerator)new MMDeviceEnumerator();
        while (true)
        {
            await Task.Delay(40); // 25 FPS
            try
            {
                if (_isMediaPlaying)
                {
                    if (DateTime.UtcNow - lastDeviceCheck > TimeSpan.FromSeconds(5))
                    {
                        foreach (var m in meters) { try { Marshal.ReleaseComObject(m); } catch { } }
                        meters.Clear();
                        
                        if (enumerator.EnumAudioEndpoints(0, 1, out object colObj) == 0 && colObj != null)
                        {
                            var collection = (IMMDeviceCollection)colObj;
                            collection.GetCount(out uint count);
                            for (uint i = 0; i < count; i++)
                            {
                                if (collection.Item(i, out object devObj) == 0 && devObj != null)
                                {
                                    var dev = (IMMDevice)devObj;
                                    Guid iid = typeof(IAudioMeterInformation).GUID;
                                    if (dev.Activate(ref iid, 23, IntPtr.Zero, out object meterObj) == 0 && meterObj != null)
                                    {
                                        meters.Add((IAudioMeterInformation)meterObj);
                                    }
                                    Marshal.ReleaseComObject(devObj);
                                }
                            }
                            Marshal.ReleaseComObject(colObj);
                        }
                        lastDeviceCheck = DateTime.UtcNow;
                    }
                    
                    float maxPeak = 0f;
                    foreach (var meter in meters)
                    {
                        try 
                        {
                            if (meter.GetPeakValue(out float peak) == 0)
                            {
                                if (peak > maxPeak) maxPeak = peak;
                            }
                        } 
                        catch { }
                    }
                    
                    try { System.IO.File.AppendAllText("audio_debug.log", $"Peak: {maxPeak}, Meters: {meters.Count}\n"); } catch { }
                    
                    Console.WriteLine($"{{\"type\":\"audio_peak\",\"peak\":{maxPeak.ToString(System.Globalization.CultureInfo.InvariantCulture)}}}");
                }
            }
            catch { }
        }
    }
    catch { }
});

_ = Task.Run(StartNotificationListener);
_ = Task.Run(StartMediaListenerAndCommandLoop);

await Task.Delay(-1);

#region Notification Listener
async Task StartNotificationListener()
{
    try
    {
        var listener = Windows.UI.Notifications.Management.UserNotificationListener.Current;
        var accessStatus = await listener.RequestAccessAsync();

        if (accessStatus != Windows.UI.Notifications.Management.UserNotificationListenerAccessStatus.Allowed)
        {
            Console.WriteLine(JsonSerializer.Serialize(new { type = "error", error = "Access denied" }));
            return;
        }

        var seenIds = new HashSet<uint>();
        var initialNotifs = await listener.GetNotificationsAsync(Windows.UI.Notifications.NotificationKinds.Toast);
        foreach (var notif in initialNotifs)
        {
            seenIds.Add(notif.Id);
        }

        Console.WriteLine(JsonSerializer.Serialize(new { type = "status", status = "Listening" }));

        while (true)
        {
            try
            {
                var notifs = await listener.GetNotificationsAsync(Windows.UI.Notifications.NotificationKinds.Toast);
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

                if (seenIds.Count > notifs.Count + 50)
                {
                    var currentIds = new HashSet<uint>(notifs.Select(n => n.Id));
                    seenIds.IntersectWith(currentIds);
                }
            }
            catch { }

            await Task.Delay(100);
        }
    }
    catch (Exception ex)
    {
        Console.WriteLine(JsonSerializer.Serialize(new { type = "error", error = ex.Message }));
    }
}

async Task<bool> ProcessNotificationAsync(Windows.UI.Notifications.UserNotification notif)
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
        var bodyBuilder = new StringBuilder();
        string heroImage = "";
        string icon = "";

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
                        if (texts.Count > 1 && bodyBuilder.Length == 0) bodyBuilder.Append(texts[1].Text ?? "");

                        for (int i = 2; i < texts.Count; i++)
                        {
                            if (!string.IsNullOrEmpty(texts[i].Text))
                            {
                                if (bodyBuilder.Length > 0) bodyBuilder.Append('\n');
                                bodyBuilder.Append(texts[i].Text);
                            }
                        }
                    }
                }
            }
        }
        catch { }

        try
        {
            if (notif.Notification?.Visual?.Bindings != null)
            {
                foreach (dynamic binding in notif.Notification.Visual.Bindings)
                {
                    try
                    {
                        var images = binding.GetImageElements();
                        if (images != null)
                        {
                            foreach (var img in images)
                            {
                                string src = img.Source ?? "";
                                File.AppendAllText(Path.Combine(Path.GetTempPath(), "vinta_img_cache_log.txt"), "RAM Extraction found image: " + src + "\n");
                                if (string.IsNullOrEmpty(icon)) icon = ProcessImageUrl(src);
                                else if (string.IsNullOrEmpty(heroImage)) heroImage = ProcessImageUrl(src);
                            }
                        }
                    }
                    catch (Exception ex) 
                    {
                        File.AppendAllText(Path.Combine(Path.GetTempPath(), "vinta_img_cache_log.txt"), "RAM Extraction Exception: " + ex.Message + "\n");
                    }
                }
            }
        }
        catch { }

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

        if (string.IsNullOrEmpty(icon))
        {
            try
            {
                var aumid = notif.AppInfo?.AppUserModelId ?? "";
                icon = ExtractAppIcon(aumid);
            }
            catch { }
        }

        try
        {
            var (extractedHero, extractedIcon) = ExtractNotificationImages(notif.Id);
            
            int retries = 10;
            while (retries > 0 && string.IsNullOrEmpty(extractedHero) && string.IsNullOrEmpty(extractedIcon) && (appName.Contains("Discord", StringComparison.OrdinalIgnoreCase) || appName.Contains("Slack", StringComparison.OrdinalIgnoreCase) || appName.Contains("Telegram", StringComparison.OrdinalIgnoreCase)))
            {
                await Task.Delay(100);
                (extractedHero, extractedIcon) = ExtractNotificationImages(notif.Id);
                retries--;
            }

            if (!string.IsNullOrEmpty(extractedIcon)) 
            {
                icon = extractedIcon;
            }
            else if (!string.IsNullOrEmpty(extractedHero) && (appName.Contains("Discord", StringComparison.OrdinalIgnoreCase) || appName.Contains("Slack", StringComparison.OrdinalIgnoreCase) || appName.Contains("Telegram", StringComparison.OrdinalIgnoreCase)))
            {
                icon = extractedHero;
            }
            else if (!string.IsNullOrEmpty(extractedHero))
            {
                heroImage = extractedHero;
            }
        }
        catch { }

        var payload = new
        {
            type = "notification",
            appName,
            icon,
            title,
            message = bodyBuilder.ToString(),
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

                heroImage = ProcessImageUrl(heroUrl);
                icon = ProcessImageUrl(iconUrl);
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
string GetCachedImagePath(string identifier)
{
    using (var sha = System.Security.Cryptography.SHA256.Create())
    {
        var hash = sha.ComputeHash(Encoding.UTF8.GetBytes(identifier));
        return Path.Combine(_vintaCacheDir, BitConverter.ToString(hash).Replace("-", "").ToLower() + ".img");
    }
}

string CacheLocalFile(string path)
{
    File.AppendAllText(Path.Combine(Path.GetTempPath(), "vinta_img_cache_log.txt"), "CacheLocalFile called for: " + path + "\n");
    if (!Directory.Exists(_vintaCacheDir)) Directory.CreateDirectory(_vintaCacheDir);
    string cachedPath = GetCachedImagePath(path);
    if (!File.Exists(cachedPath))
    {
        try
        {
            using (var fs = new FileStream(path, FileMode.Open, FileAccess.Read, FileShare.ReadWrite))
            using (var dest = new FileStream(cachedPath, FileMode.Create, FileAccess.Write, FileShare.None))
            {
                fs.CopyTo(dest);
            }
            File.AppendAllText(Path.Combine(Path.GetTempPath(), "vinta_img_cache_log.txt"), "CacheLocalFile successfully copied to: " + cachedPath + "\n");
        }
        catch (Exception ex)
        {
            File.AppendAllText(Path.Combine(Path.GetTempPath(), "vinta_img_cache_log.txt"), "CacheLocalFile Exception: " + ex.ToString() + "\n");
            throw; // Rethrow to be caught by ProcessImageUrl
        }
    }
    else
    {
        File.AppendAllText(Path.Combine(Path.GetTempPath(), "vinta_img_cache_log.txt"), "CacheLocalFile already exists at: " + cachedPath + "\n");
    }
    return "/local/" + Uri.EscapeDataString(cachedPath);
}

string ProcessImageUrl(string imgUrl)
{
    if (string.IsNullOrEmpty(imgUrl)) return "";
    try
    {
        if (imgUrl.StartsWith("http", StringComparison.OrdinalIgnoreCase))
        {
            if (!Directory.Exists(_vintaCacheDir)) Directory.CreateDirectory(_vintaCacheDir);
            string cachedPath = GetCachedImagePath(imgUrl);
            if (File.Exists(cachedPath)) return "/local/" + Uri.EscapeDataString(cachedPath);

            var response = sharedHttpClient.GetAsync(imgUrl).Result;
            if (!response.IsSuccessStatusCode) return "";
            if (response.Content.Headers.ContentLength > 2 * 1024 * 1024) return "";

            var imgBytes = response.Content.ReadAsByteArrayAsync().Result;
            if (imgBytes.Length > 2 * 1024 * 1024) return "";

            File.WriteAllBytes(cachedPath, imgBytes);
            return "/local/" + Uri.EscapeDataString(cachedPath);
        }
        else if (imgUrl.StartsWith("ms-appdata://", StringComparison.OrdinalIgnoreCase) || imgUrl.StartsWith("ms-appx://", StringComparison.OrdinalIgnoreCase))
        {
            string localAppData = Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData);
            string packagesPath = Path.Combine(localAppData, "Packages");
            string inetCachePath = Path.Combine(localAppData, @"Microsoft\Windows\INetCache");
            string fileName = Path.GetFileName(imgUrl);

            if (!string.IsNullOrEmpty(fileName))
            {
                if (Directory.Exists(inetCachePath))
                {
                    var matches = Directory.GetFiles(inetCachePath, fileName, SearchOption.AllDirectories);
                    if (matches.Length > 0 && File.Exists(matches[0]))
                    {
                        return CacheLocalFile(matches[0]);
                    }
                }

                if (Directory.Exists(packagesPath))
                {
                    try
                    {
                        var enumOptions = new EnumerationOptions { RecurseSubdirectories = true, MaxRecursionDepth = 5, IgnoreInaccessible = true };
                        var match = Directory.EnumerateFiles(packagesPath, fileName, enumOptions).FirstOrDefault();
                        if (match != null && File.Exists(match))
                        {
                            return CacheLocalFile(match);
                        }
                    }
                    catch { }
                }
            }
        }
        else if (imgUrl.StartsWith("file:///", StringComparison.OrdinalIgnoreCase))
        {
            string localPath = Uri.UnescapeDataString(imgUrl.Substring(8)).Replace('/', '\\');
            if (File.Exists(localPath))
            {
                return CacheLocalFile(localPath);
            }
        }
        else
        {
            string localPath = imgUrl.Replace('/', '\\').Trim();
            if (File.Exists(localPath))
            {
                return CacheLocalFile(localPath);
            }
        }
    }
    catch (Exception ex) 
    { 
        File.AppendAllText(Path.Combine(Path.GetTempPath(), "vinta_img_cache_log.txt"), "ProcessImageUrl Exception for " + imgUrl + ": " + ex.ToString() + "\n");
    }
    
    File.AppendAllText(Path.Combine(Path.GetTempPath(), "vinta_img_cache_log.txt"), "ProcessImageUrl failed/returned empty for: " + imgUrl + "\n");
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
        if (File.Exists(dbPath))
        {
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
                    var icon = "/local/" + Uri.EscapeDataString(logoPath);
                    _iconCache[aumid] = icon; 
                    return icon;
                }
            }
        }
    }
    catch { }

    _iconCache[aumid] = ""; 
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

    _ = Task.Run(async () =>
    {
        while (true)
        {
            await Task.Delay(3000);
            try
            {
                if (gsmtcManager != null) UpdateCurrentSession();
            }
            catch { }
        }
    });

    while (true)
    {
        try
        {
            var line = await Console.In.ReadLineAsync();
            if (line == null)
            {
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

void OnMediaPropertiesChanged(GlobalSystemMediaTransportControlsSession sender, MediaPropertiesChangedEventArgs args) => ScheduleEmit(includeVolume: false);
void OnPlaybackInfoChanged(GlobalSystemMediaTransportControlsSession sender, PlaybackInfoChangedEventArgs args) => ScheduleEmit(includeVolume: true);
void OnTimelinePropertiesChanged(GlobalSystemMediaTransportControlsSession sender, TimelinePropertiesChangedEventArgs args) => ScheduleEmit(includeVolume: false);

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
    if (currentMediaSession == null) UpdateCurrentSession();
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

        string trackKey = $"{mediaProps?.Title}_{mediaProps?.Artist}";
        string albumArtBase64 = "";

        if (!string.IsNullOrEmpty(trackKey) && _artCache.TryGetValue(trackKey, out var cachedArt))
        {
            albumArtBase64 = cachedArt;
        }

        if (string.IsNullOrEmpty(albumArtBase64) && mediaProps?.Thumbnail != null)
        {
            for (int retry = 0; retry < 3; retry++)
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
                        if (!string.IsNullOrEmpty(trackKey))
                        {
                            if (_artCache.Count > 3) _artCache.Clear();
                            _artCache[trackKey] = albumArtBase64;
                        }
                        break;
                    }
                }
                catch { await Task.Delay(80); }
            }
        }

        var isPlaying = playbackInfo?.PlaybackStatus == GlobalSystemMediaTransportControlsSessionPlaybackStatus.Playing;
        _isMediaPlaying = isPlaying;
        var controls = playbackInfo?.Controls;

        if (DateTime.UtcNow > _volumeCacheExpiry)
        {
            var (vol, muted) = GetSessionAudioVolumeAndMute();
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
            audioPeak = 0f,
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

// Robust session audio volume control with master device endpoint fallback
(float volume, bool isMuted) GetSessionAudioVolumeAndMute()
{
    object? enumeratorObj = null;
    object? devObj = null;
    object? epvObj = null;
    try
    {
        NativeMethods.CoInitializeEx(IntPtr.Zero, 0x2);
        enumeratorObj = new MMDeviceEnumerator();
        var enumerator = (IMMDeviceEnumerator)enumeratorObj;
        if (enumerator.GetDefaultAudioEndpoint(0, 0, out devObj) == 0 && devObj != null)
        {
            var dev = (IMMDevice)devObj;
            Guid epvIid = typeof(IAudioEndpointVolume).GUID;
            if (dev.Activate(ref epvIid, 23 /* CLSCTX_ALL */, IntPtr.Zero, out epvObj) == 0 && epvObj != null)
            {
                var epv = (IAudioEndpointVolume)epvObj;
                epv.GetMasterVolumeLevelScalar(out float vol);
                epv.GetMute(out bool mute);
                return (vol * 100f, mute);
            }
        }
    }
    catch { }
    finally
    {
        if (epvObj != null) try { Marshal.ReleaseComObject(epvObj); } catch { }
        if (devObj != null) try { Marshal.ReleaseComObject(devObj); } catch { }
        if (enumeratorObj != null) try { Marshal.ReleaseComObject(enumeratorObj); } catch { }
    }
    return (_cachedVolume, _cachedMute);
}

void SetSessionAudioVolume(float percent)
{
    var clamped = Math.Clamp(percent, 0f, 100f);
    _cachedVolume = clamped;
    _volumeCacheExpiry = DateTime.UtcNow.AddMilliseconds(VolumeCacheTtlMs);

    object? enumeratorObj = null;
    object? devObj = null;
    object? epvObj = null;
    try
    {
        NativeMethods.CoInitializeEx(IntPtr.Zero, 0x2);
        enumeratorObj = new MMDeviceEnumerator();
        var enumerator = (IMMDeviceEnumerator)enumeratorObj;
        if (enumerator.GetDefaultAudioEndpoint(0, 0, out devObj) == 0 && devObj != null)
        {
            var dev = (IMMDevice)devObj;
            Guid epvIid = typeof(IAudioEndpointVolume).GUID;
            if (dev.Activate(ref epvIid, 23 /* CLSCTX_ALL */, IntPtr.Zero, out epvObj) == 0 && epvObj != null)
            {
                var epv = (IAudioEndpointVolume)epvObj;
                epv.SetMasterVolumeLevelScalar(clamped / 100f, IntPtr.Zero);
            }
        }
    }
    catch { }
    finally
    {
        if (epvObj != null) try { Marshal.ReleaseComObject(epvObj); } catch { }
        if (devObj != null) try { Marshal.ReleaseComObject(devObj); } catch { }
        if (enumeratorObj != null) try { Marshal.ReleaseComObject(enumeratorObj); } catch { }
    }
}

void ToggleSessionAudioMute()
{
    _cachedMute = !_cachedMute;
    _volumeCacheExpiry = DateTime.UtcNow.AddMilliseconds(VolumeCacheTtlMs);

    object? enumeratorObj = null;
    object? devObj = null;
    object? epvObj = null;
    try
    {
        NativeMethods.CoInitializeEx(IntPtr.Zero, 0x2);
        enumeratorObj = new MMDeviceEnumerator();
        var enumerator = (IMMDeviceEnumerator)enumeratorObj;
        if (enumerator.GetDefaultAudioEndpoint(0, 0, out devObj) == 0 && devObj != null)
        {
            var dev = (IMMDevice)devObj;
            Guid epvIid = typeof(IAudioEndpointVolume).GUID;
            if (dev.Activate(ref epvIid, 23 /* CLSCTX_ALL */, IntPtr.Zero, out epvObj) == 0 && epvObj != null)
            {
                var epv = (IAudioEndpointVolume)epvObj;
                epv.SetMute(_cachedMute, IntPtr.Zero);
            }
        }
    }
    catch { }
    finally
    {
        if (epvObj != null) try { Marshal.ReleaseComObject(epvObj); } catch { }
        if (devObj != null) try { Marshal.ReleaseComObject(devObj); } catch { }
        if (enumeratorObj != null) try { Marshal.ReleaseComObject(enumeratorObj); } catch { }
    }
}


async Task HandleMediaCommandAsync(string action)
{
    if (action.StartsWith("set_volume:"))
    {
        string valStr = action.Substring("set_volume:".Length);
        if (float.TryParse(valStr, System.Globalization.NumberStyles.Float, System.Globalization.CultureInfo.InvariantCulture, out float volPercent))
        {
            SetSessionAudioVolume(volPercent);
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
        ToggleSessionAudioMute();
        SendVolumeKey(VK_VOLUME_MUTE);
        await EmitMediaStateAsync(includeVolume: true);
        return;
    }
    
    if (action.StartsWith("seek:"))
    {
        if (currentMediaSession == null) UpdateCurrentSession();
        if (currentMediaSession != null)
        {
            string seekStr = action.Substring("seek:".Length);
            if (double.TryParse(seekStr, System.Globalization.NumberStyles.Float, System.Globalization.CultureInfo.InvariantCulture, out double posMs))
            {
                try 
                { 
                    long ticks = (long)(posMs * 10000);
                    await currentMediaSession.TryChangePlaybackPositionAsync(ticks);
                    await Task.Delay(80);
                    await EmitMediaStateAsync(includeVolume: false);
                } 
                catch { }
            }
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
                case "play_pause": success = await currentMediaSession.TryTogglePlayPauseAsync(); break;
                case "play": success = await currentMediaSession.TryPlayAsync(); break;
                case "pause": success = await currentMediaSession.TryPauseAsync(); break;
                case "next": success = await currentMediaSession.TrySkipNextAsync(); break;
                case "previous":
                case "prev": success = await currentMediaSession.TrySkipPreviousAsync(); break;
            }
        }

        await Task.Delay(150);
        UpdateCurrentSession();
        await EmitMediaStateAsync(includeVolume: true);
    }
    catch { }
}
#endregion

void SendVolumeKey(byte vk)
{
    try
    {
        uint scanCode = NativeMethods.MapVirtualKey(vk, 0);
        NativeMethods.keybd_event(vk, (byte)scanCode, 0x0001, 0);
        Thread.Sleep(15);
        NativeMethods.keybd_event(vk, (byte)scanCode, 0x0001 | 0x0002, 0);
        _volumeCacheExpiry = DateTime.MinValue;
    }
    catch { }
}

#region COM Interface Declarations
internal static class ComGuids
{
    public const string MMDeviceEnumerator = "BCDE0395-E52F-467C-8E3D-C4579291692E";
    public const string IMMDeviceEnumerator = "A95664D2-9614-4F35-A746-DE8DB63617E6";
    public const string IMMDeviceCollection = "0BD7A1BE-7A1A-44DB-8397-CC5392387B5E";
    public const string IMMDevice = "D666063F-1587-4E43-81F1-B948E807363F";
    public const string IAudioEndpointVolume = "5CDF2C82-841E-4546-9722-0CF74078229A";
    public const string IAudioMeterInformation = "C02216F6-8C67-4B5B-9D00-D008E73E0064";
}

internal static class NativeMethods
{
    [DllImport("ole32.dll")]
    public static extern int CoInitializeEx(IntPtr pvReserved, uint dwCoInit);

    [DllImport("user32.dll")]
    public static extern void keybd_event(byte bVk, byte bScan, uint dwFlags, uint dwExtraInfo);

    [DllImport("user32.dll")]
    public static extern uint MapVirtualKey(uint uCode, uint uMapType);
}

[ComImport]
[Guid(ComGuids.MMDeviceEnumerator)]
class MMDeviceEnumerator { }

[Guid(ComGuids.IMMDeviceEnumerator), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IMMDeviceEnumerator
{
    [PreserveSig] int EnumAudioEndpoints(int dataFlow, int stateMask, [MarshalAs(UnmanagedType.IUnknown)] out object ppDevices);
    [PreserveSig] int GetDefaultAudioEndpoint(int dataFlow, int role, [MarshalAs(UnmanagedType.IUnknown)] out object ppEndpoint);
}

[Guid(ComGuids.IMMDeviceCollection), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IMMDeviceCollection
{
    [PreserveSig] int GetCount(out uint pcDevices);
    [PreserveSig] int Item(uint nDevice, [MarshalAs(UnmanagedType.IUnknown)] out object ppDevice);
}

[Guid(ComGuids.IMMDevice), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IMMDevice
{
    [PreserveSig] int Activate(ref Guid iid, int dwClsCtx, IntPtr pActivationParams, [MarshalAs(UnmanagedType.IUnknown)] out object? ppInterface);
}

[Guid("77AA4508-8699-4224-99CE-57163E9F0D6F"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IAudioSessionManager2
{
    [PreserveSig] int GetAudioSessionControl(IntPtr AudioSessionGuid, uint StreamFlags, out object SessionControl);
    [PreserveSig] int GetSimpleAudioVolume(IntPtr AudioSessionGuid, uint StreamFlags, out ISimpleAudioVolume AudioVolume);
    [PreserveSig] int GetSessionEnumerator(out IAudioSessionEnumerator SessionEnum);
}

[Guid("E2F56180-8904-43A0-801B-74A727366034"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IAudioSessionEnumerator
{
    [PreserveSig] int GetCount(out int SessionCount);
    [PreserveSig] int GetSession(int SessionIndex, [MarshalAs(UnmanagedType.IUnknown)] out object SessionControl);
}

[Guid("F4B89704-CA21-4361-A065-1804E92E68F3"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IAudioSessionControl
{
    [PreserveSig] int GetState(out int pRetVal);
    [PreserveSig] int GetDisplayName([MarshalAs(UnmanagedType.LPWStr)] out string pRetVal);
    [PreserveSig] int SetDisplayName([MarshalAs(UnmanagedType.LPWStr)] string Value, [In] ref Guid EventContext);
    [PreserveSig] int GetIconPath([MarshalAs(UnmanagedType.LPWStr)] out string pRetVal);
    [PreserveSig] int SetIconPath([MarshalAs(UnmanagedType.LPWStr)] string Value, [In] ref Guid EventContext);
    [PreserveSig] int GetGroupingParam(out Guid pRetVal);
    [PreserveSig] int SetGroupingParam([In] ref Guid Override, [In] ref Guid EventContext);
    [PreserveSig] int RegisterAudioSessionNotification(IntPtr NewNotifications);
    [PreserveSig] int UnregisterAudioSessionNotification(IntPtr NewNotifications);
}

[Guid("870149E1-9394-44C0-8584-82B7E7DC5B2B"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface ISimpleAudioVolume
{
    [PreserveSig] int SetMasterVolume(float fLevel, [In] ref Guid pguidEventContext);
    [PreserveSig] int GetMasterVolume(out float pfLevel);
    [PreserveSig] int SetMute([MarshalAs(UnmanagedType.Bool)] bool bMute, [In] ref Guid pguidEventContext);
    [PreserveSig] int GetMute([MarshalAs(UnmanagedType.Bool)] out bool pbMute);
}

[Guid(ComGuids.IAudioEndpointVolume), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IAudioEndpointVolume
{
    [PreserveSig] int RegisterControlChangeNotify(IntPtr pNotify);
    [PreserveSig] int UnregisterControlChangeNotify(IntPtr pNotify);
    [PreserveSig] int GetChannelCount(out uint pnChannelCount);
    [PreserveSig] int SetMasterVolumeLevel(float fLevelDB, IntPtr pguidEventContext);
    [PreserveSig] int SetMasterVolumeLevelScalar(float fLevel, IntPtr pguidEventContext);
    [PreserveSig] int GetMasterVolumeLevel(out float pfLevelDB);
    [PreserveSig] int GetMasterVolumeLevelScalar(out float pfLevel);
    [PreserveSig] int SetChannelVolumeLevel(uint nChannel, float fLevelDB, IntPtr pguidEventContext);
    [PreserveSig] int SetChannelVolumeLevelScalar(uint nChannel, float fLevel, IntPtr pguidEventContext);
    [PreserveSig] int GetChannelVolumeLevel(uint nChannel, out float pfLevelDB);
    [PreserveSig] int GetChannelVolumeLevelScalar(uint nChannel, out float pfLevel);
    [PreserveSig] int SetMute([MarshalAs(UnmanagedType.Bool)] bool bMute, IntPtr pguidEventContext);
    [PreserveSig] int GetMute([MarshalAs(UnmanagedType.Bool)] out bool pbMute);
}

[Guid(ComGuids.IAudioMeterInformation), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
interface IAudioMeterInformation
{
    [PreserveSig] int GetPeakValue(out float pfPeak);
    [PreserveSig] int GetMeteringChannelCount(out uint pnChannelCount);
    [PreserveSig] int GetChannelsPeakValues(uint u32ChannelCount, [Out] float[] afPeakValues);
    [PreserveSig] int QueryHardwareSupport(out uint pdwHardwareSupportMask);
}
#endregion