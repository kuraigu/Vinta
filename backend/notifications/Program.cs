using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading.Tasks;
using System.Xml.Linq;
using Microsoft.Data.Sqlite;
using Windows.UI.Notifications;
using Windows.UI.Notifications.Management;

var listener = UserNotificationListener.Current;
SQLitePCL.Batteries.Init();
var accessStatus = await listener.RequestAccessAsync();

if (accessStatus != UserNotificationListenerAccessStatus.Allowed)
{
    Console.WriteLine(JsonSerializer.Serialize(new ErrorPayload { error = "Access denied" }, AppJsonSerializerContext.Default.ErrorPayload));
    return;
}

var seenIds = new HashSet<uint>();

var initialNotifs = await listener.GetNotificationsAsync(NotificationKinds.Toast);
foreach (var notif in initialNotifs)
{
    seenIds.Add(notif.Id);
}

Console.WriteLine(JsonSerializer.Serialize(new StatusPayload { status = "Listening" }, AppJsonSerializerContext.Default.StatusPayload));

var retryCounts = new Dictionary<uint, int>();

while (true)
{
    try
    {
        var notifs = await listener.GetNotificationsAsync(NotificationKinds.Toast);
        foreach (var notif in notifs)
        {
            if (!seenIds.Contains(notif.Id))
            {
                if (ProcessNotification(notif))
                {
                    seenIds.Add(notif.Id);
                }
            }
        }
        
        if (seenIds.Count > notifs.Count + 100) 
        {
            var currentIds = new HashSet<uint>(notifs.Select(n => n.Id));
            seenIds.IntersectWith(currentIds);
        }
    }
    catch
    {
        // Ignored
    }

    await Task.Delay(200);
}

bool ProcessNotification(UserNotification notif)
{
    try 
    {
        string appName = "";
        try { appName = notif.AppInfo?.DisplayInfo?.DisplayName ?? ""; } catch {}
        if (string.IsNullOrEmpty(appName) || appName == "App")
        {
            try { appName = notif.AppInfo?.AppUserModelId ?? ""; } catch {}
        }

        string title = "";
        string body = "";

        if (notif.Notification?.Visual?.Bindings != null)
        {
            var bindings = notif.Notification.Visual.Bindings;
            foreach (var binding in bindings)
            {
                var texts = binding.GetTextElements();
                if (texts != null)
                {
                    if (texts.Count > 0) title = texts[0].Text;
                    if (texts.Count > 1) body = texts[1].Text;

                    for (int i = 2; i < texts.Count; i++) {
                        body += "\n" + texts[i].Text;
                    }
                }
            }
        }

        string heroImage = "";
        string icon = "";

        try 
        {
            string dbPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData), @"Microsoft\Windows\Notifications\wpndatabase.db");
            if (File.Exists(dbPath))
            {
                string tempPath = Path.GetTempFileName();
                try
                {
                    File.Copy(dbPath, tempPath, true);
                    
                    using (var connection = new SqliteConnection($"Data Source={tempPath}"))
                    {
                        connection.Open();
                        var command = connection.CreateCommand();
                        command.CommandText = "SELECT Payload FROM Notification WHERE Id = @Id";
                        command.Parameters.AddWithValue("@Id", notif.Id);
                        
                        using (var reader = command.ExecuteReader())
                        {
                            if (reader.Read())
                            {
                                byte[] payloadBytes = (byte[])reader["Payload"];
                                string xmlPayload = Encoding.UTF8.GetString(payloadBytes);
                                
                                var doc = XDocument.Parse(xmlPayload);
#if DEBUG
                                Console.Error.WriteLine($"[C# Interceptor] RAW XML: {xmlPayload}");
#endif
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
                                
                                heroImage = NotificationPayload.DownloadImageAsBase64(heroUrl);
                                icon = NotificationPayload.DownloadImageAsBase64(iconUrl);
#if DEBUG
                                Console.Error.WriteLine($"[C# Interceptor] Extracted iconUrl='{iconUrl}' (hasBase64={!string.IsNullOrEmpty(icon)}), heroUrl='{heroUrl}'");
#endif
                            }
                        }
                    }
                }
                finally
                {
                    try { if (File.Exists(tempPath)) File.Delete(tempPath); } catch {}
                }
            }
        }
        catch (Exception ex)
        {
            Console.Error.WriteLine($"[SQLite] Error extracting image: {ex.Message}");
        }

        // If SQLite image extraction failed or wpndatabase.db wasn't available,
        // we still proceed and emit the notification with its title, body, and app info!
        var payload = new NotificationPayload {
            id = notif.Id,
            appName = appName,
            title = title,
            message = body,
            heroImage = heroImage,
            icon = icon
        };

        string json = JsonSerializer.Serialize(payload, AppJsonSerializerContext.Default.NotificationPayload);
        Console.WriteLine(json);
        
        return true;
    }
    catch (Exception ex)
    {
        Console.Error.WriteLine($"[C# Interceptor] Error processing notification: {ex.Message}");
        return true;
    }
}

public class ErrorPayload { public string error { get; set; } = ""; }
public class StatusPayload { public string status { get; set; } = ""; }
public class NotificationPayload {
    public uint id { get; set; }
    public string appName { get; set; } = "";
    public string title { get; set; } = "";
    public string message { get; set; } = "";
    public string heroImage { get; set; } = "";
    public string icon { get; set; } = "";

    public static string DownloadImageAsBase64(string imgUrl)
    {
        if (string.IsNullOrEmpty(imgUrl)) return "";
        try {
            if (imgUrl.StartsWith("http")) {
                using (var client = new HttpClient())
                {
                    client.Timeout = TimeSpan.FromSeconds(3);
                    var response = client.GetAsync(imgUrl).Result;
                    if (!response.IsSuccessStatusCode) return "";
                    
                    // Cap maximum downloaded image size to 5 MB
                    if (response.Content.Headers.ContentLength > 5 * 1024 * 1024) return "";

                    var imgBytes = response.Content.ReadAsByteArrayAsync().Result;
                    if (imgBytes.Length > 5 * 1024 * 1024) return "";

                    return "data:image/jpeg;base64," + Convert.ToBase64String(imgBytes);
                }
            } else if (imgUrl.StartsWith("ms-appdata://") || imgUrl.StartsWith("ms-appx://")) {
                // Windows UWP app local asset scheme resolution
                string localAppData = Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData);
                string packagesPath = Path.Combine(localAppData, "Packages");
                string tempPath = Path.GetTempPath();
                string inetCachePath = Path.Combine(localAppData, @"Microsoft\Windows\INetCache");
                string fileName = Path.GetFileName(imgUrl);
                
                if (!string.IsNullOrEmpty(fileName)) {
                    // 1. Check Temp directory
                    string localTempFile = Path.Combine(tempPath, fileName);
                    if (File.Exists(localTempFile)) {
                        var imgBytes = File.ReadAllBytes(localTempFile);
                        return "data:image/png;base64," + Convert.ToBase64String(imgBytes);
                    }

                    // 2. Check INetCache
                    if (Directory.Exists(inetCachePath)) {
                        var matches = Directory.GetFiles(inetCachePath, fileName, SearchOption.AllDirectories);
                        if (matches.Length > 0 && File.Exists(matches[0])) {
                            var imgBytes = File.ReadAllBytes(matches[0]);
                            return "data:image/png;base64," + Convert.ToBase64String(imgBytes);
                        }
                    }

                    // 3. Search Packages directory
                    if (Directory.Exists(packagesPath))
                    {
                        var matches = Directory.GetFiles(packagesPath, fileName, SearchOption.AllDirectories);
                        if (matches.Length > 0 && File.Exists(matches[0]))
                        {
                            var fileInfo = new FileInfo(matches[0]);
                            if (fileInfo.Length <= 5 * 1024 * 1024)
                            {
                                var imgBytes = File.ReadAllBytes(matches[0]);
                                return "data:image/png;base64," + Convert.ToBase64String(imgBytes);
                            }
                        }
                    }
                }
                return "";
            } else if (imgUrl.StartsWith("file:///")) {
                string localPath = Uri.UnescapeDataString(imgUrl.Substring(8)).Replace('/', '\\');
                if (!File.Exists(localPath)) return "";
                var fileInfo = new FileInfo(localPath);
                if (fileInfo.Length > 5 * 1024 * 1024) return "";

                var imgBytes = File.ReadAllBytes(localPath);
                return "data:image/jpeg;base64," + Convert.ToBase64String(imgBytes);
            } else if (File.Exists(imgUrl)) {
                var fileInfo = new FileInfo(imgUrl);
                if (fileInfo.Length <= 5 * 1024 * 1024) {
                    var imgBytes = File.ReadAllBytes(imgUrl);
                    return "data:image/png;base64," + Convert.ToBase64String(imgBytes);
                }
            }
        } catch {}
        return "";
    }
}

[JsonSerializable(typeof(ErrorPayload))]
[JsonSerializable(typeof(StatusPayload))]
[JsonSerializable(typeof(NotificationPayload))]
internal partial class AppJsonSerializerContext : JsonSerializerContext
{
}
