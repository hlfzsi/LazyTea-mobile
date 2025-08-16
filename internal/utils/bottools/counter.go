package bottools
import (
	"sync"
	"time"
)
type timePoint struct {
	time  time.Time
	count int
}
type counter struct {
	lock        sync.Mutex
	timeWindows []timePoint
	maxPeriod   time.Duration
}
func newCounter(maxPeriod time.Duration) *counter {
	return &counter{
		timeWindows: make([]timePoint, 0),
		maxPeriod:   maxPeriod,
	}
}
func (c *counter) addEvent(eventTime time.Time, count int) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cleanup()
	c.timeWindows = append(c.timeWindows, timePoint{time: eventTime, count: count})
}
func (c *counter) cleanup() {
	cutoff := time.Now().Add(-c.maxPeriod)
	firstKeeper := 0
	for i, p := range c.timeWindows {
		if p.time.After(cutoff) {
			firstKeeper = i
			break
		}
		firstKeeper = i + 1
	}
	if firstKeeper > 0 {
		c.timeWindows = c.timeWindows[firstKeeper:]
	}
}
func (c *counter) getTotalCount() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cleanup()
	total := 0
	for _, p := range c.timeWindows {
		total += p.count
	}
	return total
}
func (c *counter) getPeriodCount(period time.Duration) int {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cleanup()
	cutoff := time.Now().Add(-period)
	total := 0
	for _, p := range c.timeWindows {
		if p.time.After(cutoff) {
			total += p.count
		}
	}
	return total
}
type MsgCounter struct {
	lock      sync.RWMutex
	counters  map[string]map[string]*counter  
	maxPeriod time.Duration
}
func NewMsgCounter(maxPeriod time.Duration) *MsgCounter {
	return &MsgCounter{
		counters:  make(map[string]map[string]*counter),
		maxPeriod: maxPeriod,
	}
}
func (mc *MsgCounter) getCounter(botID, eventType string) *counter {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	if _, ok := mc.counters[botID]; !ok {
		mc.counters[botID] = make(map[string]*counter)
	}
	if _, ok := mc.counters[botID][eventType]; !ok {
		mc.counters[botID][eventType] = newCounter(mc.maxPeriod)
	}
	return mc.counters[botID][eventType]
}
func (mc *MsgCounter) IncrementCount(botID string) {
	counter := mc.getCounter(botID, "all")
	counter.addEvent(time.Now(), 1)
}
func (mc *MsgCounter) GetTotalCount(botID string) int {
	mc.lock.RLock()
	botCounters, ok := mc.counters[botID]
	mc.lock.RUnlock()
	if !ok {
		return 0
	}
	if c, ok := botCounters["all"]; ok {
		return c.getTotalCount()
	}
	return 0
}
func (mc *MsgCounter) GetPeriodCount(botID string, periodSeconds int) int {
	mc.lock.RLock()
	botCounters, ok := mc.counters[botID]
	mc.lock.RUnlock()
	if !ok {
		return 0
	}
	if c, ok := botCounters["all"]; ok {
		return c.getPeriodCount(time.Duration(periodSeconds) * time.Second)
	}
	return 0
}
func (mc *MsgCounter) Reset(botID string) {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	delete(mc.counters, botID)
}
func (mc *MsgCounter) ResetAll() {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	mc.counters = make(map[string]map[string]*counter)
}
func (mc *MsgCounter) GetAllCounts() map[string]int {
	mc.lock.RLock()
	defer mc.lock.RUnlock()
	counts := make(map[string]int)
	for botID := range mc.counters {
		counts[botID] = mc.GetTotalCount(botID)
	}
	return counts
}
func (mc *MsgCounter) CleanupOldData() {
	mc.lock.RLock()
	defer mc.lock.RUnlock()
	for _, botCounters := range mc.counters {
		for _, c := range botCounters {
			c.lock.Lock()
			c.cleanup()
			c.lock.Unlock()
		}
	}
}