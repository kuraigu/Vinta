// Package devtest provides development tools for triggering mock notifications.
package devtest

import (
	"fmt"
	"os/exec"
	"syscall"
)

// DevTest provides testing functionality for the application.
type DevTest struct{}

// NewDevTest creates a new instance of DevTest.
func NewDevTest() *DevTest {
	return &DevTest{}
}

// SendTestNotification triggers a native Windows Toast notification via PowerShell.
// Uses PowerShell's own AUMID so the toast is delivered to the notification center
// and intercepted by the WinRT UserNotificationListener.
func (d *DevTest) SendTestNotification() {
	script := `
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$appId = '{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe'
$toastXml = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText01)
$toastXml.LoadXml('<toast><visual><binding template="ToastGeneric"><text>Hello Vinta</text><text>This is a test notification.</text><image placement="appLogoOverride" src="https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png" /><image placement="hero" src="https://upload.wikimedia.org/wikipedia/commons/4/47/PNG_transparency_demonstration_1.png" /></binding></visual></toast>')
$toast = [Windows.UI.Notifications.ToastNotification]::new($toastXml)
$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($appId)
$notifier.Show($toast)
`
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Start()
}

// SendTestLongMessageToast triggers a toast with a very long message
func (d *DevTest) SendTestLongMessageToast() {
	script := `
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$appId = '{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe'
$toastXml = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText01)
$toastXml.LoadXml('<toast><visual><binding template="ToastGeneric"><text>Extremely Long Title Truncation Test</text><text>This is an incredibly long notification message specifically designed to test whether the Vinta Notification Pill correctly handles CSS text-overflow and white-space wrap mechanics. It should truncate nicely in the minimized pill view, but properly expand to show this entire paragraph inside the full list view!</text></binding></visual></toast>')
$toast = [Windows.UI.Notifications.ToastNotification]::new($toastXml)
$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($appId)
$notifier.Show($toast)
`
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Start()
}

// SendTestNoTitleToast triggers a toast with only a message, mimicking NVIDIA
func (d *DevTest) SendTestNoTitleToast() {
	script := `
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$appId = '{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe'
$toastXml = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText01)
$toastXml.LoadXml('<toast><visual><binding template="ToastGeneric"><text></text><text>NVIDIA GeForce NOW: Gameplay recording has started.</text></binding></visual></toast>')
$toast = [Windows.UI.Notifications.ToastNotification]::new($toastXml)
$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($appId)
$notifier.Show($toast)
`
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Start()
}

// SendTestBadge triggers a Windows Badge notification via PowerShell
func (d *DevTest) SendTestBadge(value string) {
	script := fmt.Sprintf(`
[Windows.UI.Notifications.BadgeUpdateManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$badgeXml = [Windows.UI.Notifications.BadgeUpdateManager]::GetTemplateContent([Windows.UI.Notifications.BadgeTemplateType]::BadgeNumber)
$badgeElement = $badgeXml.SelectSingleNode('/badge')
$badgeElement.SetAttribute('value', '%s')
$badge = [Windows.UI.Notifications.BadgeNotification]::new($badgeXml)
$notifier = [Windows.UI.Notifications.BadgeUpdateManager]::CreateBadgeUpdaterForApplication('Microsoft.WindowsStore_8wekyb3d8bbwe!App')
$notifier.Update($badge)
`, value)
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Start()
}
