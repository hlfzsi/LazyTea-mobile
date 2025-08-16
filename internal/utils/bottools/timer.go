package bottools
import (
	"sync"
	"time"
)
type BotStatus struct {
	lock          sync.Mutex
	startTime     time.Time
	offlineStart  *time.Time
	totalOffline  time.Duration
}
func NewBotStatus() *BotStatus {
	return &BotStatus{
		startTime:    time.Now(),
		totalOffline: 0,
	}
}
type BotTimer struct {
	bots map[string]*BotStatus
	lock sync.RWMutex
}
func NewBotTimer() *BotTimer {
	return &BotTimer{
		bots: make(map[string]*BotStatus),
	}
}
func (bt *BotTimer) AddBot(botID string) {
	bt.lock.Lock()
	defer bt.lock.Unlock()
	if _, exists := bt.bots[botID]; !exists {
		bt.bots[botID] = NewBotStatus()
	}
}
func (bt *BotTimer) RemoveBot(botID string) {
	bt.lock.Lock()
	defer bt.lock.Unlock()
	delete(bt.bots, botID)
}
func (bt *BotTimer) SetOffline(botID string) {
	bt.lock.RLock()
	status, exists := bt.bots[botID]
	bt.lock.RUnlock()
	if exists {
		status.lock.Lock()
		defer status.lock.Unlock()
		if status.offlineStart == nil {
			now := time.Now()
			status.offlineStart = &now
		}
	}
}
func (bt *BotTimer) SetOnline(botID string) {
	bt.lock.RLock()
	status, exists := bt.bots[botID]
	bt.lock.RUnlock()
	if exists {
		status.lock.Lock()
		defer status.lock.Unlock()
		if status.offlineStart != nil {
			status.totalOffline += time.Since(*status.offlineStart)
			status.offlineStart = nil
		}
	}
}
func (bt *BotTimer) GetElapsedTime(botID string) time.Duration {
	bt.lock.RLock()
	status, exists := bt.bots[botID]
	bt.lock.RUnlock()
	if !exists {
		return 0
	}
	status.lock.Lock()
	defer status.lock.Unlock()
	var currentOffline time.Duration
	if status.offlineStart != nil {
		currentOffline = time.Since(*status.offlineStart)
	}
	totalOfflineDuration := status.totalOffline + currentOffline
	totalElapsed := time.Since(status.startTime)
	return totalElapsed - totalOfflineDuration
}
func (bt *BotTimer) IsOnline(botID string) bool {
	bt.lock.RLock()
	status, exists := bt.bots[botID]
	bt.lock.RUnlock()
	if !exists {
		return false
	}
	status.lock.Lock()
	defer status.lock.Unlock()
	return status.offlineStart == nil
}
func (bt *BotTimer) GetAllTimers() map[string]*BotStatus {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	result := make(map[string]*BotStatus, len(bt.bots))
	for k, v := range bt.bots {
		result[k] = v
	}
	return result
}
func (bt *BotTimer) ResetAll() {
	bt.lock.Lock()
	defer bt.lock.Unlock()
	bt.bots = make(map[string]*BotStatus)
}
func (bt *BotTimer) Stop() {
}