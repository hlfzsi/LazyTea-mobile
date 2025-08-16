package mobile
import (
	"log"
	"sync"
	"time"
)
type NetworkState int
const (
	NetworkStateUnknown NetworkState = iota
	NetworkStateConnected
	NetworkStateDisconnected
	NetworkStateConnecting
)
type MobileNetworkManager struct {
	mutex               sync.RWMutex
	state               NetworkState
	lastConnectionTime  time.Time
	reconnectAttempts   int
	maxReconnectAttempts int
	reconnectDelay      time.Duration
	onStateChange       func(NetworkState)
	onReconnectNeeded   func()
	backgroundReconnect bool
}
func NewMobileNetworkManager() *MobileNetworkManager {
	return &MobileNetworkManager{
		state:                NetworkStateUnknown,
		maxReconnectAttempts: 5,
		reconnectDelay:       5 * time.Second,
		backgroundReconnect:  true,
	}
}
func (nm *MobileNetworkManager) SetStateChangeCallback(callback func(NetworkState)) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.onStateChange = callback
}
func (nm *MobileNetworkManager) SetReconnectCallback(callback func()) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.onReconnectNeeded = callback
}
func (nm *MobileNetworkManager) GetState() NetworkState {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	return nm.state
}
func (nm *MobileNetworkManager) SetState(state NetworkState) {
	nm.mutex.Lock()
	oldState := nm.state
	nm.state = state
	if state == NetworkStateConnected {
		nm.lastConnectionTime = time.Now()
		nm.reconnectAttempts = 0
	}
	nm.mutex.Unlock()
	if oldState != state {
		log.Printf("[Mobile] Network state changed: %v -> %v", 
			nm.stateString(oldState), nm.stateString(state))
		if nm.onStateChange != nil {
			go nm.onStateChange(state)
		}
	}
}
func (nm *MobileNetworkManager) OnConnectionLost() {
	nm.SetState(NetworkStateDisconnected)
	if IsMobile() {
		go nm.handleReconnection()
	}
}
func (nm *MobileNetworkManager) OnConnectionEstablished() {
	nm.SetState(NetworkStateConnected)
}
func (nm *MobileNetworkManager) OnConnectionAttempt() {
	nm.SetState(NetworkStateConnecting)
}
func (nm *MobileNetworkManager) handleReconnection() {
	nm.mutex.Lock()
	if nm.reconnectAttempts >= nm.maxReconnectAttempts {
		nm.mutex.Unlock()
		log.Printf("[Mobile] Max reconnection attempts (%d) reached", nm.maxReconnectAttempts)
		return
	}
	nm.reconnectAttempts++
	attempts := nm.reconnectAttempts
	delay := nm.getReconnectDelay()
	nm.mutex.Unlock()
	log.Printf("[Mobile] Scheduling reconnection attempt %d/%d in %v", 
		attempts, nm.maxReconnectAttempts, delay)
	time.Sleep(delay)
	nm.mutex.RLock()
	shouldReconnect := nm.state == NetworkStateDisconnected && 
		nm.reconnectAttempts <= nm.maxReconnectAttempts
	nm.mutex.RUnlock()
	if shouldReconnect && nm.onReconnectNeeded != nil {
		log.Printf("[Mobile] Triggering reconnection attempt %d", attempts)
		nm.onReconnectNeeded()
	}
}
func (nm *MobileNetworkManager) getReconnectDelay() time.Duration {
	baseDelay := nm.reconnectDelay
	multiplier := time.Duration(1 << uint(nm.reconnectAttempts-1))
	delay := baseDelay * multiplier
	if delay > 60*time.Second {
		delay = 60 * time.Second
	}
	return delay
}
func (nm *MobileNetworkManager) ResetReconnectionAttempts() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.reconnectAttempts = 0
}
func (nm *MobileNetworkManager) ShouldReconnectInBackground() bool {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	return nm.backgroundReconnect
}
func (nm *MobileNetworkManager) SetBackgroundReconnect(enabled bool) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.backgroundReconnect = enabled
}
func (nm *MobileNetworkManager) GetConnectionInfo() map[string]interface{} {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	info := map[string]interface{}{
		"state":             nm.stateString(nm.state),
		"reconnect_attempts": nm.reconnectAttempts,
		"max_attempts":      nm.maxReconnectAttempts,
		"background_reconnect": nm.backgroundReconnect,
	}
	if !nm.lastConnectionTime.IsZero() {
		info["last_connection"] = nm.lastConnectionTime.Format(time.RFC3339)
		info["time_since_connection"] = time.Since(nm.lastConnectionTime).String()
	}
	return info
}
func (nm *MobileNetworkManager) stateString(state NetworkState) string {
	switch state {
	case NetworkStateConnected:
		return "Connected"
	case NetworkStateDisconnected:
		return "Disconnected"
	case NetworkStateConnecting:
		return "Connecting"
	default:
		return "Unknown"
	}
}
func (nm *MobileNetworkManager) HandleAppBackground() {
	if !IsMobile() {
		return
	}
	log.Println("[Mobile] App went to background, adjusting network behavior")
	nm.mutex.Lock()
	nm.reconnectDelay = 30 * time.Second
	nm.mutex.Unlock()
}
func (nm *MobileNetworkManager) HandleAppForeground() {
	if !IsMobile() {
		return
	}
	log.Println("[Mobile] App returned to foreground, adjusting network behavior")
	nm.mutex.Lock()
	nm.reconnectDelay = 5 * time.Second
	nm.mutex.Unlock()
	if nm.state == NetworkStateDisconnected {
		log.Println("[Mobile] Triggering immediate reconnection check")
		if nm.onReconnectNeeded != nil {
			go nm.onReconnectNeeded()
		}
	}
}