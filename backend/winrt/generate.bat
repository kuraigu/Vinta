@echo off
set GEN=go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen@v0.0.0-20260513072510-45f10383b2b8

%GEN% -class Windows.UI.Notifications.Management.UserNotificationListener
%GEN% -class Windows.UI.Notifications.UserNotification
%GEN% -class Windows.UI.Notifications.Notification
%GEN% -class Windows.UI.Notifications.NotificationVisual
%GEN% -class Windows.UI.Notifications.NotificationBinding
%GEN% -class Windows.UI.Notifications.AdaptiveNotificationText
%GEN% -class Windows.Foundation.IAsyncOperation`1
%GEN% -class Windows.Foundation.Collections.IVectorView`1
