package pages
import (
	"fmt"
	"lazytea-mobile/internal/config"
	"lazytea-mobile/internal/data"
	"lazytea-mobile/internal/network"
	"lazytea-mobile/internal/utils"
	"os"
	"time"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)
type OverviewPage struct {
	*PageBase
	config                *config.Config
	connectionStatusLabel *widget.Label
	onlineBotCountLabel   *widget.Label
	messageCountLabel     *widget.Label
	versionLabel   *widget.Label
	statsContainer *fyne.Container
}
func NewOverviewPage(client *network.Client, storage *data.Storage, logger *utils.Logger, cfg *config.Config) *OverviewPage {
	page := &OverviewPage{
		PageBase: NewPageBase(client, storage, logger),
		config:   cfg,
	}
	page.setupUI()
	page.setupEventHandlers()
	go func() {
		time.Sleep(100 * time.Millisecond)
		page.refreshData()
	}()
	return page
}
func (p *OverviewPage) setupUI() {
	titleLabel := widget.NewLabelWithStyle("🌟 LazyTea Mobile", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	subtitleLabel := widget.NewLabelWithStyle("实时监控您的聊天机器人", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	subtitleLabel.Importance = widget.MediumImportance
	p.versionLabel = widget.NewLabel(fmt.Sprintf("v%s", os.Getenv("UIVERSION")))
	versionCard := p.createModernInfoCard("📱 版本信息", p.versionLabel, theme.InfoIcon(), "success")
	p.connectionStatusLabel = widget.NewLabel("未连接")
	p.connectionStatusLabel.Importance = widget.DangerImportance
	connectionCard := p.createModernInfoCard("🔌 连接状态", p.connectionStatusLabel, theme.ComputerIcon(), "danger")
	p.onlineBotCountLabel = widget.NewLabel("0")
	botCard := p.createModernInfoCard("🤖 在线机器人", p.onlineBotCountLabel, theme.ComputerIcon(), "primary")
	p.messageCountLabel = widget.NewLabel("0")
	messageCard := p.createModernInfoCard("💬 消息总数", p.messageCountLabel, theme.MailComposeIcon(), "info")
	statsGrid := container.NewGridWithColumns(2,
		versionCard,
		connectionCard,
		botCard,
		messageCard,
	)
	p.statsContainer = container.NewVBox(
		titleLabel,
		subtitleLabel,
		widget.NewSeparator(),
		statsGrid,
	)
	refreshBtn := widget.NewButtonWithIcon("🔄 刷新数据", theme.ViewRefreshIcon(), func() {
		p.refreshData()
	})
	refreshBtn.Importance = widget.MediumImportance
	connectBtn := widget.NewButtonWithIcon("🔗 连接服务器", theme.ComputerIcon(), func() {
		p.toggleConnection()
	})
	connectBtn.Importance = widget.HighImportance
	buttonContainer := container.NewVBox(
		refreshBtn,
		connectBtn,
	)
	authorLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("开发者: %s", os.Getenv("UIAUTHOR")),
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	content := container.NewVBox(
		p.statsContainer,
		widget.NewSeparator(),
		buttonContainer,
		widget.NewSeparator(),
		authorLabel,
	)
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(300, 500))
	p.SetContent(scroll)
}
func (p *OverviewPage) createModernInfoCard(title string, valueLabel *widget.Label, icon fyne.Resource, cardType string) *widget.Card {
	iconWidget := widget.NewIcon(icon)
	iconWidget.Resize(fyne.NewSize(24, 24))  
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	switch cardType {
	case "success":
		titleLabel.Importance = widget.SuccessImportance
		valueLabel.Importance = widget.SuccessImportance
	case "danger":
		titleLabel.Importance = widget.DangerImportance
		valueLabel.Importance = widget.DangerImportance
	case "warning":
		titleLabel.Importance = widget.WarningImportance
		valueLabel.Importance = widget.WarningImportance
	case "info":
		titleLabel.Importance = widget.MediumImportance
		valueLabel.Importance = widget.MediumImportance
	case "primary":
		titleLabel.Importance = widget.HighImportance
		valueLabel.Importance = widget.HighImportance
	}
	valueLabel.Alignment = fyne.TextAlignCenter
	valueLabel.TextStyle = fyne.TextStyle{Bold: true}
	headerContainer := container.NewHBox(
		iconWidget,
		titleLabel,
	)
	spacer := container.NewWithoutLayout()
	spacer.Resize(fyne.NewSize(1, 8))
	content := container.NewVBox(
		headerContainer,
		spacer,
		valueLabel,
	)
	card := widget.NewCard("", "", content)
	return card
}
func (p *OverviewPage) setupEventHandlers() {
	p.client.OnConnectionChanged(func(connected bool) {
		if connected {
			p.connectionStatusLabel.SetText("✓ 已连接")
			p.connectionStatusLabel.Importance = widget.SuccessImportance
		} else {
			p.connectionStatusLabel.SetText("✗ 未连接")
			p.connectionStatusLabel.Importance = widget.DangerImportance
			p.onlineBotCountLabel.SetText("0 / 0")
		}
		p.refreshData()
	})
	p.client.OnMessage("bot_connect", func(header network.MessageHeader, payload interface{}) {
		go p.refreshData()  
	})
	p.client.OnMessage("bot_disconnect", func(header network.MessageHeader, payload interface{}) {
		go p.refreshData()  
	})
	p.client.OnMessage("message", func(header network.MessageHeader, payload interface{}) {
		go func() {
			totalMsgs, err := p.storage.GetTotalMessageCount()
			if err == nil {
				p.messageCountLabel.SetText(fmt.Sprintf("%d", totalMsgs))
			}
		}()
	})
}
func (p *OverviewPage) refreshData() {
	go func() {
		bots, err := p.storage.GetBotInfoList()
		if err != nil {
			p.logger.Error("Failed to get bot info: %v", err)
			p.onlineBotCountLabel.SetText("获取失败")
		} else {
			onlineCount := 0
			for _, bot := range bots {
				if bot.IsOnline {
					onlineCount++
				}
			}
			p.onlineBotCountLabel.SetText(fmt.Sprintf("%d / %d", onlineCount, len(bots)))
			if onlineCount > 0 {
				p.onlineBotCountLabel.Importance = widget.SuccessImportance
			} else if len(bots) > 0 {
				p.onlineBotCountLabel.Importance = widget.WarningImportance
			} else {
				p.onlineBotCountLabel.Importance = widget.MediumImportance
			}
		}
		totalMsgs, err := p.storage.GetTotalMessageCount()
		if err != nil {
			p.logger.Error("Failed to get message count: %v", err)
			p.messageCountLabel.SetText("获取失败")
		} else {
			if totalMsgs >= 1000000 {
				p.messageCountLabel.SetText(fmt.Sprintf("%.1fM", float64(totalMsgs)/1000000))
			} else if totalMsgs >= 1000 {
				p.messageCountLabel.SetText(fmt.Sprintf("%.1fK", float64(totalMsgs)/1000))
			} else {
				p.messageCountLabel.SetText(fmt.Sprintf("%d", totalMsgs))
			}
		}
	}()
}
func (p *OverviewPage) toggleConnection() {
	if p.client.IsConnected() {
		p.client.Disconnect()
		p.logger.Info("断开连接")
		return
	}
	if p.config.Network.Host == "" || p.config.Network.Port == 0 {
		p.logger.Error("连接配置不完整，请在设置页面配置服务器信息")
		return
	}
	p.logger.Info("正在连接到服务器...")
	go func() {
		if err := p.client.Connect(
			p.config.Network.Host,
			p.config.Network.Port,
			p.config.Network.Token,
		); err != nil {
			p.logger.Error("连接失败: %v", err)
		}
	}()
}
