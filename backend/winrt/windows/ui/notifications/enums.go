//go:build windows

package notifications

type NotificationKinds int32

const (
	NotificationKindsUnknown NotificationKinds = 0
	NotificationKindsToast   NotificationKinds = 1
)
