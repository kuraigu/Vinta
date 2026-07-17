//go:build windows

package management

type UserNotificationListenerAccessStatus int32

const (
	UserNotificationListenerAccessStatusUnspecified UserNotificationListenerAccessStatus = 0
	UserNotificationListenerAccessStatusAllowed     UserNotificationListenerAccessStatus = 1
	UserNotificationListenerAccessStatusDenied      UserNotificationListenerAccessStatus = 2
)
