package bottools
import (
	"sync"
	"time"
)
type BotToolKit struct {
	ColorMap *ColorMap
	Counter  *MsgCounter
	Timer    *BotTimer
	mu       sync.RWMutex
	initOnce sync.Once
}
var (
	defaultToolKit *BotToolKit
	toolkitOnce    sync.Once
)
func GetDefaultToolKit() *BotToolKit {
	toolkitOnce.Do(func() {
		defaultToolKit = NewBotToolKit()
	})
	return defaultToolKit
}
func NewBotToolKit() *BotToolKit {
	return &BotToolKit{
		ColorMap: NewColorMap(),
		Counter:  NewMsgCounter(time.Hour),  
		Timer:    NewBotTimer(),
	}
}
func (btk *BotToolKit) AddBot(botID string) {
	btk.mu.Lock()
	defer btk.mu.Unlock()
	btk.Timer.AddBot(botID)
	btk.ColorMap.Get(botID)  
}
func (btk *BotToolKit) RemoveBot(botID string) {
	btk.mu.Lock()
	defer btk.mu.Unlock()
	btk.Timer.RemoveBot(botID)
	btk.ColorMap.Remove(botID)
	btk.Counter.Reset(botID)
}
func (btk *BotToolKit) SetBotOnline(botID string, isOnline bool) {
	btk.mu.Lock()
	defer btk.mu.Unlock()
	if isOnline {
		btk.Timer.SetOnline(botID)
	} else {
		btk.Timer.SetOffline(botID)
	}
}
func (btk *BotToolKit) IncrementMessage(botID string) {
	btk.Counter.IncrementCount(botID)
}
func (btk *BotToolKit) GetBotColor(botID string) string {
	return btk.ColorMap.Get(botID)
}
func (btk *BotToolKit) GetBotStats(botID string) BotStats {
	uptime := btk.Timer.GetElapsedTime(botID)
	total := btk.Counter.GetTotalCount(botID)
	onlineMinutes := int(uptime.Minutes())
	if onlineMinutes == 0 {
		onlineMinutes = 1
	}
	periodMinutes := min(onlineMinutes, 30)
	rate := 0.0
	if periodMinutes > 0 {
		rate = float64(btk.Counter.GetPeriodCount(botID, periodMinutes*60)) / float64(periodMinutes)
	}
	return BotStats{
		Total:    total,
		Rate:     rate,
		Uptime:   int(uptime.Seconds()),
		IsOnline: btk.Timer.IsOnline(botID),
		Color:    btk.ColorMap.Get(botID),
	}
}
func (btk *BotToolKit) GetAllStats() map[string]BotStats {
	btk.mu.RLock()
	defer btk.mu.RUnlock()
	allCounts := btk.Counter.GetAllCounts()
	allTimers := btk.Timer.GetAllTimers()
	result := make(map[string]BotStats)
	allBots := make(map[string]bool)
	for botID := range allCounts {
		allBots[botID] = true
	}
	for botID := range allTimers {
		allBots[botID] = true
	}
	for botID := range allBots {
		result[botID] = btk.GetBotStats(botID)
	}
	return result
}
func (btk *BotToolKit) Cleanup() {
	btk.mu.Lock()
	defer btk.mu.Unlock()
	btk.Counter.CleanupOldData()
}
func (btk *BotToolKit) Reset() {
	btk.mu.Lock()
	defer btk.mu.Unlock()
	btk.ColorMap.Reset()
	btk.Counter.ResetAll()
	btk.Timer.ResetAll()
}
func (btk *BotToolKit) Stop() {
	btk.mu.Lock()
	defer btk.mu.Unlock()
	btk.Timer.Stop()
}
type BotStats struct {
	Total    int     `json:"total"`
	Rate     float64 `json:"rate"`
	Uptime   int     `json:"uptime"`
	IsOnline bool    `json:"is_online"`
	Color    string  `json:"color"`
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func AddBot(botID string) {
	GetDefaultToolKit().AddBot(botID)
}
func RemoveBot(botID string) {
	GetDefaultToolKit().RemoveBot(botID)
}
func SetBotOnline(botID string, isOnline bool) {
	GetDefaultToolKit().SetBotOnline(botID, isOnline)
}
func IncrementMessage(botID string) {
	GetDefaultToolKit().IncrementMessage(botID)
}
func GetBotColor(botID string) string {
	return GetDefaultToolKit().GetBotColor(botID)
}
func GetBotStats(botID string) BotStats {
	return GetDefaultToolKit().GetBotStats(botID)
}
func GetAllStats() map[string]BotStats {
	return GetDefaultToolKit().GetAllStats()
}
func Cleanup() {
	GetDefaultToolKit().Cleanup()
}
func Reset() {
	GetDefaultToolKit().Reset()
}
func Stop() {
	GetDefaultToolKit().Stop()
}
