using System;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading.Tasks;
using Windows.UI.Notifications;
using Windows.UI.Notifications.Management;
using System.Xml.Linq;
using System.Linq;
using System.IO;
using Microsoft.Data.Sqlite;

var listener = UserNotificationListener.Current;
SQLitePCL.Batteries.Init();
var accessStatus = await listener.RequestAccessAsync();

if (accessStatus != UserNotificationListenerAccessStatus.Allowed)
{
    Console.WriteLine(JsonSerializer.Serialize(new ErrorPayload { error = "Access denied" }, AppJsonSerializerContext.Default.ErrorPayload));
    return;
}

var seenIds = new System.Collections.Generic.HashSet<uint>();

var initialNotifs = await listener.GetNotificationsAsync(NotificationKinds.Toast);
foreach (var notif in initialNotifs)
{
    seenIds.Add(notif.Id);
}

Console.WriteLine(JsonSerializer.Serialize(new StatusPayload { status = "Listening" }, AppJsonSerializerContext.Default.StatusPayload));

while (true)
{
    try
    {
        var notifs = await listener.GetNotificationsAsync(NotificationKinds.Toast);
        var retryCounts = new System.Collections.Generic.Dictionary<uint, int>();
        foreach (var notif in notifs)
        {
            if (!seenIds.Contains(notif.Id))
            {
                if (ProcessNotification(notif, retryCounts))
                {
                    seenIds.Add(notif.Id);
                    retryCounts.Remove(notif.Id);
                }
            }
        }
        
        if (seenIds.Count > notifs.Count + 100) 
        {
            var currentIds = new System.Collections.Generic.HashSet<uint>();
            foreach (var notif in notifs) currentIds.Add(notif.Id);
            seenIds.IntersectWith(currentIds);
        }
    }
    catch (Exception)
    {
    }

    await Task.Delay(200);
}

bool ProcessNotification(UserNotification notif, System.Collections.Generic.Dictionary<uint, int> retryCounts)
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

        if (notif.Notification != null && notif.Notification.Visual != null && notif.Notification.Visual.Bindings != null)
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

        bool dbSuccess = false;
        try 
        {
            string dbPath = Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData) + @"\Microsoft\Windows\Notifications\wpndatabase.db";
            if (File.Exists(dbPath))
            {
                string tempPath = Path.GetTempFileName();
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
                            dbSuccess = true;
                            byte[] payloadBytes = (byte[])reader["Payload"];
                            string xmlPayload = System.Text.Encoding.UTF8.GetString(payloadBytes);
                            
                            var doc = XDocument.Parse(xmlPayload);
                            Console.Error.WriteLine($"[C# Interceptor] RAW XML: {xmlPayload}");
                            string heroUrl = "";
                            string iconUrl = "";
                            foreach (var img in doc.Descendants().Where(x => x.Name.LocalName == "image"))
                            {
                                string placement = img.Attribute("placement")?.Value ?? "";
                                string src = img.Attribute("src")?.Value ?? "";
                                if (placement == "appLogoOverride") iconUrl = src;
                                else if (string.IsNullOrEmpty(heroUrl)) heroUrl = src;
                            }
                            
                            heroImage = NotificationPayload.DownloadImageAsBase64(heroUrl);
                            icon = NotificationPayload.DownloadImageAsBase64(iconUrl);
                        }
                    }
                }
                
                try { File.Delete(tempPath); } catch {}
            }
        }
        catch (Exception ex)
        {
            Console.Error.WriteLine($"[SQLite] Error extracting image: {ex.Message}");
        }

        if (!dbSuccess)
        {
            if (!retryCounts.ContainsKey(notif.Id)) retryCounts[notif.Id] = 0;
            retryCounts[notif.Id]++;
            if (retryCounts[notif.Id] < 10) return false;
        }
        
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

public class ErrorPayload { public string error { get; set; } }
public class StatusPayload { public string status { get; set; } }
public class NotificationPayload {
    public uint id { get; set; }
    public string appName { get; set; }
    public string title { get; set; }
    public string message { get; set; }
    public string heroImage { get; set; }
    public string icon { get; set; }

    public static string DownloadImageAsBase64(string imgUrl)
    {
        if (string.IsNullOrEmpty(imgUrl)) return "";
        try {
            if (imgUrl.StartsWith("http")) {
                using (var client = new System.Net.Http.HttpClient())
                {
                    var imgBytes = client.GetByteArrayAsync(imgUrl).Result;
                    return "data:image/jpeg;base64," + Convert.ToBase64String(imgBytes);
                }
            } else if (imgUrl.StartsWith("file:///")) {
                string localPath = Uri.UnescapeDataString(imgUrl.Substring(8)).Replace('/', '\\');
                var imgBytes = System.IO.File.ReadAllBytes(localPath);
                return "data:image/jpeg;base64," + Convert.ToBase64String(imgBytes);
            } else if (imgUrl.StartsWith("C:\\") || imgUrl.StartsWith("D:\\")) {
                var imgBytes = System.IO.File.ReadAllBytes(imgUrl);
                return "data:image/jpeg;base64," + Convert.ToBase64String(imgBytes);
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
