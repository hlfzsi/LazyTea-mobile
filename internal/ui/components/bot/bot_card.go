package bot
import (
	"fmt"
	"image/color"
	"lazytea-mobile/internal/data"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)
type BotStats struct {
	Total  int     `json:"total"`
	Rate   float64 `json:"rate"`
	Uptime int     `json:"uptime"`
}
type BotCard struct {
	widget.BaseWidget
	botInfo  data.BotInfo
	stats    BotStats
	isOnline bool
	onToggleStatus func(botID string, isOnline bool)
	onShowDetails  func(botInfo data.BotInfo, botStats BotStats)
	onShowRoster   func(botID string)
	statusIcon    *canvas.Circle
	statusLabel   *widget.Label
	statsLabel    *widget.Label
	actionButtons *fyne.Container
	cardContainer *fyne.Container
}
func NewBotCard(botInfo data.BotInfo, onToggleStatus func(botID string, isOnline bool), onShowDetails func(botInfo data.BotInfo, botStats BotStats), onShowRoster func(botID string)) *BotCard {
	card := &BotCard{
		botInfo:        botInfo,
		isOnline:       botInfo.IsOnline,
		onToggleStatus: onToggleStatus,
		onShowDetails:  onShowDetails,
		onShowRoster:   onShowRoster,
		stats:          BotStats{},
	}
	card.ExtendBaseWidget(card)
	card.setupUI()
	return card
}
func (c *BotCard) setupUI() {
	c.statusIcon = canvas.NewCircle(color.RGBA{76, 175, 80, 255})
	c.statusIcon.Resize(fyne.NewSize(16, 16))  
	c.statusLabel = widget.NewLabel("在线")
	c.statusLabel.TextStyle.Bold = true
	c.statusLabel.Importance = widget.SuccessImportance
	c.statsLabel = widget.NewLabel("消息: 0 | 速率: 0/min")
	c.statsLabel.Importance = widget.MediumImportance
	c.statsLabel.Wrapping = fyne.TextWrapWord  
	c.actionButtons = c.createActionButtons()
	c.cardContainer = container.NewVBox()
	botIDLabel := widget.NewLabelWithStyle(c.botInfo.ID, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	botIDLabel.Importance = widget.HighImportance
	header := container.NewBorder(
		nil, nil,
		container.NewHBox(c.statusIcon, botIDLabel),
		c.statusLabel,
		nil,
	)
	platformLabel := widget.NewLabel(fmt.Sprintf("%s · %s", c.botInfo.Platform, c.botInfo.AdapterName))
	platformLabel.Importance = widget.MediumImportance
	c.cardContainer.Add(header)
	c.cardContainer.Add(platformLabel)
	c.cardContainer.Add(widget.NewSeparator())
	c.cardContainer.Add(c.statsLabel)
	c.cardContainer.Add(c.actionButtons)
	c.updateStatusDisplay()
}
func (c *BotCard) createActionButtons() *fyne.Container {
	toggleBtn := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		if c.onToggleStatus != nil {
			c.onToggleStatus(c.botInfo.ID, c.isOnline)
		}
	})
	toggleBtn.Importance = widget.HighImportance
	toggleBtn.Resize(fyne.NewSize(48, 48))  
	detailsBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		if c.onShowDetails != nil {
			c.onShowDetails(c.botInfo, c.stats)
		}
	})
	detailsBtn.Importance = widget.MediumImportance
	detailsBtn.Resize(fyne.NewSize(40, 40))
	rosterBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		if c.onShowRoster != nil {
			c.onShowRoster(c.botInfo.ID)
		}
	})
	rosterBtn.Importance = widget.MediumImportance
	rosterBtn.Resize(fyne.NewSize(40, 40))
	return container.NewHBox(
		toggleBtn,
		layout.NewSpacer(),
		detailsBtn,
		rosterBtn,
	)
}
func (c *BotCard) updateStatusDisplay() {
	if c.isOnline {
		c.statusIcon.FillColor = color.RGBA{76, 175, 80, 255}  
		c.statusLabel.SetText("在线")
		c.statusLabel.Importance = widget.SuccessImportance
	} else {
		c.statusIcon.FillColor = color.RGBA{244, 67, 54, 255}  
		c.statusLabel.SetText("离线")
		c.statusLabel.Importance = widget.DangerImportance
	}
	if toggleBtn := c.actionButtons.Objects[0].(*widget.Button); toggleBtn != nil {
		if c.isOnline {
			toggleBtn.SetIcon(theme.MediaStopIcon())
		} else {
			toggleBtn.SetIcon(theme.MediaPlayIcon())
		}
	}
	c.statusIcon.Refresh()
	c.statusLabel.Refresh()
}
func (c *BotCard) SetOnlineStatus(isOnline bool) {
	c.isOnline = isOnline
	c.updateStatusDisplay()
}
func (c *BotCard) UpdateData(total int, rate float64, uptime int) {
	c.stats = BotStats{
		Total:  total,
		Rate:   rate,
		Uptime: uptime,
	}
	uptimeStr := formatUptime(uptime)
	if rate > 0 || total > 0 {
		c.statsLabel.SetText(fmt.Sprintf("💬 %d  •  🚀 %.1f/min  •  ⏱️ %s", total, rate, uptimeStr))
	} else {
		c.statsLabel.SetText("暂无数据")
	}
	c.statsLabel.Refresh()
}
func (c *BotCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.cardContainer)
}
func formatUptime(seconds int) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}
func (c *BotCard) Tapped(*fyne.PointEvent) {
	if c.onShowDetails != nil {
		c.onShowDetails(c.botInfo, c.stats)
	}
}
