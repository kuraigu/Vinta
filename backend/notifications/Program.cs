// ============================================================================
// Vinta Notification Interceptor (C#)
// ============================================================================
// This standalone console application uses the native WinRT UserNotificationListener 
// API to securely intercept Windows Action Center notifications. 
//
// It is spawned as a background subprocess by the Vinta Go backend. When a 
// notification arrives, it extracts the details (Title, Message, AppUserModelID) 
// and streams them back to the Go backend via standard output (stdout) as JSON.
// ============================================================================

using System;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading.Tasks;
using Windows.UI.Notifications;
using Windows.UI.Notifications.Management;

var listener = UserNotificationListener.Current;
var accessStatus = await listener.RequestAccessAsync();

if (accessStatus != UserNotificationListenerAccessStatus.Allowed)
{
    Console.WriteLine(JsonSerializer.Serialize(new ErrorPayload { error = "Access denied" }, AppJsonSerializerContext.Default.ErrorPayload));
    return;
}

// Track seen notifications to avoid duplicates in polling
var seenIds = new System.Collections.Generic.HashSet<uint>();

var initialNotifs = await listener.GetNotificationsAsync(NotificationKinds.Toast);
foreach (var notif in initialNotifs)
{
    seenIds.Add(notif.Id);
}

Console.WriteLine(JsonSerializer.Serialize(new StatusPayload { status = "Listening" }, AppJsonSerializerContext.Default.StatusPayload));

// Poll for new notifications
while (true)
{
    try
    {
        var notifs = await listener.GetNotificationsAsync(NotificationKinds.Toast);
        foreach (var notif in notifs)
        {
            if (seenIds.Add(notif.Id))
            {
                ProcessNotification(notif);
            }
        }
        
        // Remove old IDs that are no longer in the action center
        if (seenIds.Count > notifs.Count + 100) 
        {
            var currentIds = new System.Collections.Generic.HashSet<uint>();
            foreach (var notif in notifs) currentIds.Add(notif.Id);
            seenIds.IntersectWith(currentIds);
        }
    }
    catch (Exception)
    {
        // Ignore COM exceptions during polling
    }

    await Task.Delay(200);
}

void ProcessNotification(UserNotification notif)
{
    try 
    {
        string appName = notif.AppInfo?.DisplayInfo?.DisplayName ?? "";
        if (string.IsNullOrEmpty(appName) || appName == "App")
        {
            appName = notif.AppInfo?.AppUserModelId ?? "";
        }

        string title = "";
        string body = "";
        string heroImage = "";

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

        var payload = new NotificationPayload {
            id = notif.Id,
            appName = appName,
            title = title,
            message = body,
            heroImage = heroImage
        };

        Console.WriteLine(JsonSerializer.Serialize(payload, AppJsonSerializerContext.Default.NotificationPayload));
    }
    catch (Exception ex)
    {
        Console.Error.WriteLine($"Error processing notification {notif.Id}: {ex.Message}");
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
}

[JsonSerializable(typeof(ErrorPayload))]
[JsonSerializable(typeof(StatusPayload))]
[JsonSerializable(typeof(NotificationPayload))]
internal partial class AppJsonSerializerContext : JsonSerializerContext
{
}
