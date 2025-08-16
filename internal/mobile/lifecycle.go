package mobile
import (
	"log"
	"os"
	"time"
	"fyne.io/fyne/v2"
)
type MobileLifecycle struct {
	app            fyne.App
	onBackground   func()
	onForeground   func()
	onLowMemory    func()
	backgroundTime time.Time
	isInBackground bool
	reconnectTimer *time.Timer
}
func NewMobileLifecycle(fyneApp fyne.App) *MobileLifecycle {
	ml := &MobileLifecycle{
		app: fyneApp,
	}
	if IsMobile() {
		ml.setupMobileCallbacks()
	}
	return ml
}
func (ml *MobileLifecycle) SetBackgroundCallback(callback func()) {
	ml.onBackground = callback
}
func (ml *MobileLifecycle) SetForegroundCallback(callback func()) {
	ml.onForeground = callback
}
func (ml *MobileLifecycle) SetLowMemoryCallback(callback func()) {
	ml.onLowMemory = callback
}
func (ml *MobileLifecycle) setupMobileCallbacks() {
	log.Println("[Mobile] Mobile lifecycle callbacks configured")
}
func (ml *MobileLifecycle) HandleBackground() {
	if ml.isInBackground {
		return
	}
	ml.isInBackground = true
	ml.backgroundTime = time.Now()
	log.Println("[Mobile] App went to background")
	if ml.onBackground != nil {
		ml.onBackground()
	}
	ml.scheduleReconnection()
}
func (ml *MobileLifecycle) HandleForeground() {
	if !ml.isInBackground {
		return
	}
	ml.isInBackground = false
	backgroundDuration := time.Since(ml.backgroundTime)
	log.Printf("[Mobile] App returned to foreground after %v", backgroundDuration)
	if ml.reconnectTimer != nil {
		ml.reconnectTimer.Stop()
		ml.reconnectTimer = nil
	}
	if ml.onForeground != nil {
		ml.onForeground()
	}
}
func (ml *MobileLifecycle) HandleLowMemory() {
	log.Println("[Mobile] Low memory warning received")
	if ml.onLowMemory != nil {
		ml.onLowMemory()
	}
}
func (ml *MobileLifecycle) scheduleReconnection() {
	if ml.reconnectTimer != nil {
		ml.reconnectTimer.Stop()
	}
	ml.reconnectTimer = time.AfterFunc(30*time.Second, func() {
		if ml.isInBackground {
			log.Println("[Mobile] Triggering background reconnection")
		}
	})
}
func (ml *MobileLifecycle) IsInBackground() bool {
	return ml.isInBackground
}
func (ml *MobileLifecycle) GetBackgroundDuration() time.Duration {
	if !ml.isInBackground {
		return 0
	}
	return time.Since(ml.backgroundTime)
}
func IsMobile() bool {
	return os.Getenv("ANDROID_ROOT") != "" || os.Getenv("ANDROID_DATA") != ""
}
func GetDeviceInfo() map[string]interface{} {
	if IsMobile() {
		return map[string]interface{}{
			"is_mobile":    true,
			"platform":     "android",
			"has_keyboard": false,  
		}
	}
	device := fyne.CurrentDevice()
	return map[string]interface{}{
		"is_mobile":    device.IsMobile(),
		"has_keyboard": device.HasKeyboard(),
		"orientation":  device.Orientation(),
	}
}
func ConfigureForMobile(fyneApp fyne.App) {
	if !IsMobile() {
		return
	}
	log.Println("[Mobile] Configuring app for mobile device")
	metadata := fyneApp.Metadata()
	metadata.Custom = map[string]string{
		"mobile_optimized": "true",
		"touch_friendly":   "true",
	}
	prefs := fyneApp.Preferences()
	prefs.SetBool("mobile_mode", true)
	log.Println("[Mobile] Mobile configuration applied")
}
