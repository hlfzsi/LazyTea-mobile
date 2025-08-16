package pages
import (
	"encoding/json"
	"fmt"
	"lazytea-mobile/internal/data"
	"lazytea-mobile/internal/network"
	"lazytea-mobile/internal/ui/components/bot"
	"lazytea-mobile/internal/utils"
	"lazytea-mobile/internal/utils/bottools"
	"sync"
	"time"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)
type BotInfoPage struct {
	*PageBase
	cardContainer *fyne.Container
	cardManager   *BotCardManager
	statusLabel   *widget.Label
	mainWindow    fyne.Window
	toolkit       *bottools.BotToolKit
	body          *fyne.Container  
}
func NewBotInfoPage(client *network.Client, storage *data.Storage, logger *utils.Logger, mainWindow fyne.Window) *BotInfoPage {
	page := &BotInfoPage{
		PageBase:   NewPageBase(client, storage, logger),
		mainWindow: mainWindow,
		toolkit:    bottools.GetDefaultToolKit(),
	}
	page.cardManager = NewBotCardManager(page.handleToggleStatus, page.handleShowDetails, page.handleShowRoster)
	page.setupUI()
	page.setupEventHandlers()
	return page
}
func (p *BotInfoPage) setupEventHandlers() {
	p.client.OnConnectionChanged(func(connected bool) {
		if connected {
			p.statusLabel.SetText("已连接")
		} else {
			p.statusLabel.SetText("未连接")
			p.cardManager.Clear()
			if bots, err := p.storage.GetBotInfoList(); err == nil {
				for _, bot := range bots {
					p.storage.UpdateBotOnlineStatus(bot.ID, false)
				}
			}
			p.refreshCardLayout()
			p.updateBotCount()
		}
	})
	p.client.OnMessage("bot_connect", func(header network.MessageHeader, payload interface{}) {
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			botID, _ := payloadMap["bot"].(string)
			adapter, _ := payloadMap["adapter"].(string)
			platform, _ := payloadMap["platform"].(string)
			p.logger.Info("Bot connected: %s (%s via %s)", botID, platform, adapter)
			p.toolkit.AddBot(botID)
			p.toolkit.SetBotOnline(botID, true)
			botInfo := data.BotInfo{
				ID:          botID,
				AdapterName: adapter,
				Platform:    platform,
				IsOnline:    true,
				LastSeen:    time.Now(),
			}
			if err := p.storage.SaveBotInfo(botInfo); err != nil {
				p.logger.Error("Failed to save bot info: %v", err)
			}
			p.cardManager.AddOrUpdate(botInfo)
			p.refreshCardLayout()
			p.updateBotCount()
		}
	})
	p.client.OnMessage("message", func(header network.MessageHeader, payload interface{}) {
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			if botID, ok := payloadMap["bot"].(string); ok {
				p.toolkit.IncrementMessage(botID)
				go func() {
				}()
			}
		}
	})
	p.client.OnMessage("bot_disconnect", func(header network.MessageHeader, payload interface{}) {
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			botID, _ := payloadMap["bot"].(string)
			p.logger.Info("Bot disconnected: %s", botID)
			p.toolkit.SetBotOnline(botID, false)
			if err := p.storage.UpdateBotOnlineStatus(botID, false); err != nil {
				p.logger.Error("Failed to update bot offline status: %v", err)
			}
			p.cardManager.SetOnlineStatus(botID, false)
			p.updateBotCount()
		}
	})
	p.startDataRefreshTimer()
}
func (p *BotInfoPage) setupUI() {
	titleLabel := widget.NewLabelWithStyle("Bot 管理", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	titleLabel.Importance = widget.HighImportance
	p.statusLabel = widget.NewLabel("正在连接...")
	p.statusLabel.Importance = widget.MediumImportance
	header := container.NewVBox(
		container.NewHBox(titleLabel, layout.NewSpacer(), p.statusLabel),
		widget.NewSeparator(),
	)
	p.cardContainer = container.NewVBox()
	paddedContainer := container.NewPadded(p.cardContainer)
	scroll := container.NewVScroll(paddedContainer)
	scroll.SetMinSize(fyne.NewSize(320, 480))
	p.body = container.NewStack(scroll)
	content := container.NewBorder(header, nil, nil, nil, p.body)
	p.SetContent(content)
}
func (p *BotInfoPage) refreshCardLayout() {
	cards := p.cardManager.GetCards()
	p.logger.Info("刷新卡片布局，共 %d 张卡片", len(cards))
	p.cardContainer.RemoveAll()
	for i, card := range cards {
		p.cardContainer.Add(card)
		if i < len(cards)-1 {
			spacer := container.NewWithoutLayout()
			spacer.Resize(fyne.NewSize(1, 12))
			p.cardContainer.Add(spacer)
		}
	}
	if len(cards) == 0 {
		emptyLabel := widget.NewLabelWithStyle("暂无Bot数据\n\n请确保已连接到LazyTea服务器", fyne.TextAlignCenter, fyne.TextStyle{})
		emptyLabel.Importance = widget.MediumImportance
		p.cardContainer.Add(emptyLabel)
	}
	p.cardContainer.Refresh()
}
func (p *BotInfoPage) updateBotCount() {
	online, total := p.cardManager.GetCounts()
	p.statusLabel.SetText(fmt.Sprintf("%d 在线 / %d 总数", online, total))
}
func (p *BotInfoPage) startDataRefreshTimer() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			p.refreshAllData()
		}
	}()
}
func (p *BotInfoPage) refreshAllData() {
	onlineCount := 0
	offlineCount := 0
	allBotInfos := p.cardManager.GetAllBotInfos()
	for botID, botInfo := range allBotInfos {
		if botInfo.IsOnline {
			onlineCount++
			onlineTime := p.toolkit.Timer.GetElapsedTime(botID)
			onlineMinutes := int(onlineTime.Minutes())
			if onlineMinutes == 0 {
				onlineMinutes = 1
			}
			periodMinutes := onlineMinutes
			if periodMinutes > 30 {
				periodMinutes = 30
			}
			rate := float64(p.toolkit.Counter.GetPeriodCount(botID, 1800)) / float64(periodMinutes)
			p.cardManager.UpdateData(
				botID,
				p.toolkit.Counter.GetTotalCount(botID),
				rate,
				int(onlineTime.Seconds()),
			)
		} else {
			offlineCount++
		}
	}
	if onlineCount > 0 || offlineCount > 0 {
		p.statusLabel.SetText(fmt.Sprintf("%d 在线 / %d 离线", onlineCount, offlineCount))
	}
}
func (p *BotInfoPage) handleToggleStatus(botID string, isOnlineNow bool) {
	botInfo, ok := p.cardManager.GetBotInfo(botID)
	if !ok {
		p.logger.Warn("Bot not found for status toggle: %s", botID)
		return
	}
	newState := !isOnlineNow
	p.logger.Info("Sending 'bot_switch' for bot %s, platform %s. New state: %v", botID, botInfo.Platform, newState)
	params := map[string]interface{}{
		"bot_id":        botID,
		"platform":      botInfo.Platform,
		"is_online_now": newState,
	}
	callback := &network.RequestCallback{
		Success: func(payload interface{}) {
			p.logger.Info("Successfully sent bot_switch request for bot %s", botID)
			p.cardManager.SetOnlineStatus(botID, newState)
			p.updateBotCount()
		},
		Error: func(err error) {
			p.logger.Error("Failed to send bot_switch request: %v", err)
			dialog.ShowError(err, p.mainWindow)
		},
	}
	if err := p.client.SendRequestWithCallback("bot_switch", params, callback); err != nil {
		p.logger.Error("Failed to send bot_switch request: %v", err)
		dialog.ShowError(err, p.mainWindow)
	}
}
func (p *BotInfoPage) handleShowDetails(botInfo data.BotInfo, botStats bot.BotStats) {
	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Bot ID: %s", botInfo.ID)),
		widget.NewLabel(fmt.Sprintf("Platform: %s", botInfo.Platform)),
		widget.NewLabel(fmt.Sprintf("Adapter: %s", botInfo.AdapterName)),
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("消息总量: %d", botStats.Total)),
		widget.NewLabel(fmt.Sprintf("处理速率: %.1f/min", botStats.Rate)),
		widget.NewLabel(fmt.Sprintf("在线时长: %s", formatUptime(botStats.Uptime))),
	)
	dialog.ShowCustom("Bot 详情", "关闭", content, p.mainWindow)
}
func formatUptime(seconds int) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
func (p *BotInfoPage) handleShowRoster(botID string) {
	p.logger.Info("开始获取Bot %s 的名单数据", botID)
	callback := &network.RequestCallback{
		Success: func(payload interface{}) {
			p.logger.Info("收到名单数据响应，payload类型: %T", payload)
			if jsonBytes, err := json.Marshal(payload); err == nil {
				p.logger.Info("原始响应内容: %s", string(jsonBytes))
			}
			var payloadMap map[string]interface{}
			if respPayload, ok := payload.(map[string]interface{}); ok {
				p.logger.Info("响应是map类型，包含以下键: %v", getMapKeys(respPayload))
				if data, hasData := respPayload["data"]; hasData {
					p.logger.Info("响应包含data字段，data类型: %T", data)
					if dataMap, ok := data.(map[string]interface{}); ok {
						p.logger.Info("data是map类型，包含以下键: %v", getMapKeys(dataMap))
						payloadMap = dataMap
					} else {
						p.logger.Warn("data字段不是map类型，使用原始响应")
						payloadMap = respPayload
					}
				} else {
					p.logger.Info("响应不包含data字段，使用原始响应")
					payloadMap = respPayload
				}
			} else {
				dialog.ShowError(fmt.Errorf("无效的payload类型: %T", payload), p.mainWindow)
				return
			}
			p.logger.Info("最终用于解析的payloadMap包含以下键: %v", getMapKeys(payloadMap))
			pageBase := NewPageBase(p.client, p.storage, p.logger)
			rosterPage := NewRosterPageForBot(payloadMap, func(data map[string]interface{}) {
				p.logger.Info("名单数据保存触发")
			}, p.mainWindow, pageBase, botID)  
			backBtn := widget.NewButton("返回", func() {
				padded := container.NewPadded(p.cardContainer)
				scroll := container.NewVScroll(padded)
				scroll.SetMinSize(fyne.NewSize(320, 480))
				p.body.Objects = []fyne.CanvasObject{scroll}
				p.body.Refresh()
			})
			backBtn.Importance = widget.MediumImportance
			inline := container.NewBorder(
				container.NewVBox(widget.NewSeparator(), backBtn, widget.NewSeparator()),
				nil, nil, nil,
				rosterPage.GetContent(),
			)
			p.body.Objects = []fyne.CanvasObject{inline}
			p.body.Refresh()
		},
		Error: func(err error) {
			dialog.ShowError(err, p.mainWindow)
		},
	}
	if err := p.client.SendRequestWithCallback("get_matchers", map[string]interface{}{}, callback); err != nil {
		dialog.ShowError(err, p.mainWindow)
	}
}
type BotCardManager struct {
	mu             sync.RWMutex
	cards          map[string]*bot.BotCard
	botInfos       map[string]data.BotInfo
	botStats       map[string]bot.BotStats
	onToggleStatus func(botID string, isOnline bool)
	onShowDetails  func(botInfo data.BotInfo, botStats bot.BotStats)
	onShowRoster   func(botID string)
}
func NewBotCardManager(onToggleStatus func(botID string, isOnline bool), onShowDetails func(botInfo data.BotInfo, botStats bot.BotStats), onShowRoster func(botID string)) *BotCardManager {
	return &BotCardManager{
		cards:          make(map[string]*bot.BotCard),
		botInfos:       make(map[string]data.BotInfo),
		botStats:       make(map[string]bot.BotStats),
		onToggleStatus: onToggleStatus,
		onShowDetails:  onShowDetails,
		onShowRoster:   onShowRoster,
	}
}
func (m *BotCardManager) AddOrUpdate(botInfo data.BotInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.botInfos[botInfo.ID] = botInfo
	if card, exists := m.cards[botInfo.ID]; exists {
		card.SetOnlineStatus(botInfo.IsOnline)
	} else {
		card := bot.NewBotCard(botInfo, m.onToggleStatus, m.onShowDetails, m.onShowRoster)
		m.cards[botInfo.ID] = card
	}
}
func (m *BotCardManager) SetOnlineStatus(botID string, isOnline bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if botInfo, ok := m.botInfos[botID]; ok {
		botInfo.IsOnline = isOnline
		m.botInfos[botID] = botInfo
	}
	if card, exists := m.cards[botID]; exists {
		card.SetOnlineStatus(isOnline)
	}
}
func (m *BotCardManager) UpdateData(botID string, total int, rate float64, uptime int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.botStats[botID] = bot.BotStats{
		Total:  total,
		Rate:   rate,
		Uptime: uptime,
	}
	if card, exists := m.cards[botID]; exists {
		card.UpdateData(total, rate, uptime)
	}
}
func (m *BotCardManager) GetBotInfo(botID string) (data.BotInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	botInfo, ok := m.botInfos[botID]
	return botInfo, ok
}
func (m *BotCardManager) GetBotStats(botID string) (bot.BotStats, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	stats, ok := m.botStats[botID]
	return stats, ok
}
func (m *BotCardManager) GetCards() []fyne.CanvasObject {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cards := make([]fyne.CanvasObject, 0, len(m.cards))
	for _, card := range m.cards {
		cards = append(cards, card)
	}
	return cards
}
func (m *BotCardManager) GetCounts() (online, total int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	total = len(m.botInfos)
	for _, botInfo := range m.botInfos {
		if botInfo.IsOnline {
			online++
		}
	}
	return
}
func (m *BotCardManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cards = make(map[string]*bot.BotCard)
	m.botInfos = make(map[string]data.BotInfo)
	m.botStats = make(map[string]bot.BotStats)
}
func (m *BotCardManager) GetAllBotInfos() map[string]data.BotInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]data.BotInfo)
	for k, v := range m.botInfos {
		result[k] = v
	}
	return result
}
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
