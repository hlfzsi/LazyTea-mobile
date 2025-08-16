package pages

import (
	"fmt"
	"lazytea-mobile/internal/config"
	"lazytea-mobile/internal/data"
	"lazytea-mobile/internal/network"
	"lazytea-mobile/internal/utils"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	fyneTheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type SettingsPage struct {
	*PageBase
	config            *config.Config
	window            fyne.Window
	hostEntry         *widget.Entry
	portEntry         *widget.Entry
	tokenEntry        *widget.Entry
	autoConnectCheck  *widget.Check
	rememberAuthCheck *widget.Check
	autoScrollCheck   *widget.Check
	statusLabel       *widget.Label
	connectBtn        *widget.Button
	saveBtn           *widget.Button
	resetBtn          *widget.Button
}

func NewSettingsPage(client *network.Client, storage *data.Storage, logger *utils.Logger, window fyne.Window, cfg *config.Config) *SettingsPage {
	page := &SettingsPage{
		PageBase: NewPageBase(client, storage, logger),
		config:   cfg,
		window:   window,
	}
	page.setupUI()
	page.setupEventHandlers()
	page.loadSettings()
	return page
}
func (p *SettingsPage) setupUI() {
	titleLabel := widget.NewLabelWithStyle("⚙️ 应用设置", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	titleLabel.Importance = widget.HighImportance
	p.statusLabel = widget.NewLabel("设置已加载")
	p.statusLabel.Alignment = fyne.TextAlignCenter
	p.statusLabel.Importance = widget.SuccessImportance
	connectionCard := p.createConnectionCard()
	databaseCard := p.createDatabaseCard()
	aboutCard := p.createAboutCard()
	buttonContainer := p.createActionButtons()
	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		p.statusLabel,
		widget.NewSeparator(),
		connectionCard,
		databaseCard,
		aboutCard,
		widget.NewSeparator(),
		buttonContainer,
	)
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(350, 550))
	p.SetContent(scroll)
}
func (p *SettingsPage) createConnectionCard() *widget.Card {
	p.hostEntry = widget.NewEntry()
	p.hostEntry.SetPlaceHolder("127.0.0.1")
	p.portEntry = widget.NewEntry()
	p.portEntry.SetPlaceHolder("8080")
	p.tokenEntry = widget.NewPasswordEntry()
	p.tokenEntry.SetPlaceHolder("请输入访问令牌")
	p.autoConnectCheck = widget.NewCheck("启动时自动连接", nil)
	p.rememberAuthCheck = widget.NewCheck("记住认证信息", nil)
	p.connectBtn = widget.NewButtonWithIcon("测试连接", fyneTheme.ComputerIcon(), func() {
		p.testConnection()
	})
	p.connectBtn.Importance = widget.MediumImportance
	content := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("服务器地址:"),
			p.hostEntry,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel("端口:"),
			p.portEntry,
		),
		container.NewVBox(
			widget.NewLabel("访问令牌:"),
			p.tokenEntry,
		),
		widget.NewSeparator(),
		p.autoConnectCheck,
		p.rememberAuthCheck,
		widget.NewSeparator(),
		p.connectBtn,
	)
	return widget.NewCard("连接设置", "", content)
}
func (p *SettingsPage) createDatabaseCard() *widget.Card {
	uri, err := config.GetDatabaseURI()
	var dbPathText string
	if err != nil {
		dbPathText = fmt.Sprintf("数据库路径获取失败: %v", err)
	} else {
		dbPathText = uri.String()
	}
	dbPathLabel := widget.NewLabel(dbPathText)
	dbPathLabel.Importance = widget.MediumImportance
	cleanBtn := widget.NewButtonWithIcon("清理数据", fyneTheme.DeleteIcon(), func() {
		p.confirmCleanDatabase()
	})
	cleanBtn.Importance = widget.DangerImportance
	content := container.NewVBox(
		container.NewVBox(
			widget.NewLabel("数据库路径:"),
			dbPathLabel,
		),
		widget.NewSeparator(),
		cleanBtn,
	)
	return widget.NewCard("数据设置", "", content)
}
func (p *SettingsPage) createAboutCard() *widget.Card {
	versionLabel := widget.NewLabel(fmt.Sprintf("版本: %s", os.Getenv("UIVERSION")))
	authorLabel := widget.NewLabel(fmt.Sprintf("开发者: %s", os.Getenv("UIAUTHOR")))
	githubBtn := widget.NewButtonWithIcon("GitHub", fyneTheme.ComputerIcon(), func() {
		dialog.ShowInformation("GitHub", "项目主页: https://github.com/hlfzsi/LazyTea-mobile", p.window)
	})
	updateBtn := widget.NewButtonWithIcon("检查更新", fyneTheme.ViewRefreshIcon(), func() {
		p.checkForUpdates()
	})
	content := container.NewVBox(
		versionLabel,
		authorLabel,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, githubBtn, updateBtn),
	)
	return widget.NewCard("关于", "", content)
}
func (p *SettingsPage) createActionButtons() *fyne.Container {
	p.saveBtn = widget.NewButtonWithIcon("保存设置", fyneTheme.DocumentSaveIcon(), func() {
		p.saveSettings()
	})
	p.saveBtn.Importance = widget.HighImportance
	p.resetBtn = widget.NewButtonWithIcon("重置默认", fyneTheme.ViewRefreshIcon(), func() {
		p.resetToDefaults()
	})
	p.resetBtn.Importance = widget.MediumImportance
	return container.NewGridWithColumns(2, p.saveBtn, p.resetBtn)
}
func (p *SettingsPage) setupEventHandlers() {
	p.client.OnConnectionChanged(func(connected bool) {
		if connected {
			p.connectBtn.SetText("连接成功")
			p.connectBtn.SetIcon(fyneTheme.ConfirmIcon())
			p.connectBtn.Importance = widget.SuccessImportance
		} else {
			p.connectBtn.SetText("测试连接")
			p.connectBtn.SetIcon(fyneTheme.ComputerIcon())
			p.connectBtn.Importance = widget.MediumImportance
		}
	})
}
func (p *SettingsPage) loadSettings() {
	p.hostEntry.SetText(p.config.Network.Host)
	p.portEntry.SetText(strconv.Itoa(p.config.Network.Port))
	p.tokenEntry.SetText(p.config.Network.Token)
	p.autoConnectCheck.SetChecked(p.config.Network.AutoConnect)
	p.rememberAuthCheck.SetChecked(p.config.Network.RememberAuth)
	p.statusLabel.SetText("设置已加载")
	p.statusLabel.Importance = widget.SuccessImportance
}
func (p *SettingsPage) saveSettings() {
	if strings.TrimSpace(p.hostEntry.Text) == "" {
		dialog.ShowError(fmt.Errorf("服务器地址不能为空"), p.window)
		return
	}
	port, err := strconv.Atoi(p.portEntry.Text)
	if err != nil || port <= 0 || port > 65535 {
		dialog.ShowError(fmt.Errorf("端口必须是1-65535之间的数字"), p.window)
		return
	}
	p.config.Network.Host = strings.TrimSpace(p.hostEntry.Text)
	p.config.Network.Port = port
	p.config.Network.Token = p.tokenEntry.Text
	p.config.Network.AutoConnect = p.autoConnectCheck.Checked
	p.config.Network.RememberAuth = p.rememberAuthCheck.Checked
	if err := config.Save(p.config); err != nil {
		dialog.ShowError(fmt.Errorf("保存设置失败: %v", err), p.window)
		return
	}
	p.statusLabel.SetText("设置已保存")
	p.statusLabel.Importance = widget.SuccessImportance
	dialog.ShowInformation("保存成功", "设置已保存，部分设置需要重启应用后生效", p.window)
}
func (p *SettingsPage) testConnection() {
	host := strings.TrimSpace(p.hostEntry.Text)
	portStr := strings.TrimSpace(p.portEntry.Text)
	token := p.tokenEntry.Text
	if host == "" {
		dialog.ShowError(fmt.Errorf("请输入服务器地址"), p.window)
		return
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		dialog.ShowError(fmt.Errorf("端口必须是1-65535之间的数字"), p.window)
		return
	}
	p.connectBtn.SetText("连接中...")
	p.connectBtn.Disable()
	go func() {
		if p.client.IsConnected() {
			p.client.Disconnect()
		}
		if err := p.client.Connect(host, port, token); err != nil {
			p.connectBtn.SetText("连接失败")
			p.connectBtn.SetIcon(fyneTheme.CancelIcon())
			p.connectBtn.Importance = widget.DangerImportance
			p.connectBtn.Enable()
			dialog.ShowError(fmt.Errorf("连接失败: %v", err), p.window)
		} else {
			p.connectBtn.SetText("连接成功")
			p.connectBtn.SetIcon(fyneTheme.ConfirmIcon())
			p.connectBtn.Importance = widget.SuccessImportance
			p.connectBtn.Enable()
		}
	}()
}
func (p *SettingsPage) resetToDefaults() {
	dialog.ShowConfirm(
		"重置设置",
		"确定要将所有设置重置为默认值吗？此操作不可撤销。",
		func(confirmed bool) {
			if confirmed {
				p.hostEntry.SetText("127.0.0.1")
				p.portEntry.SetText("8080")
				p.tokenEntry.SetText("疯狂星期四V我50")
				p.autoConnectCheck.SetChecked(false)
				p.rememberAuthCheck.SetChecked(false)
				p.autoScrollCheck.SetChecked(true)
				p.statusLabel.SetText("已重置为默认设置")
				p.statusLabel.Importance = widget.MediumImportance
			}
		},
		p.window,
	)
}
func (p *SettingsPage) confirmCleanDatabase() {
	dialog.ShowConfirm(
		"清理数据",
		"确定要清理所有数据吗？这将删除所有消息记录、Bot信息和插件数据。此操作不可撤销！",
		func(confirmed bool) {
			if confirmed {
				p.cleanDatabase()
			}
		},
		p.window,
	)
}
func (p *SettingsPage) cleanDatabase() {
	p.statusLabel.SetText("正在清理数据...")
	p.statusLabel.Importance = widget.MediumImportance
	go func() {
		if err := p.storage.ClearAllData(); err != nil {
			p.statusLabel.SetText(fmt.Sprintf("清理失败: %v", err))
			p.statusLabel.Importance = widget.DangerImportance
			dialog.ShowError(fmt.Errorf("数据库清理失败: %v", err), p.window)
			return
		}
		p.statusLabel.SetText("数据清理成功")
		p.statusLabel.Importance = widget.SuccessImportance
		dialog.ShowInformation("清理完成", "所有数据已成功清理，包括消息记录、插件调用记录和连接配置。", p.window)
	}()
}
func (p *SettingsPage) checkForUpdates() {
	p.statusLabel.SetText("正在检查更新...")
	p.statusLabel.Importance = widget.MediumImportance
	go func() {
		updateChecker := utils.NewUpdateChecker()
		release, hasUpdate, err := updateChecker.CheckForUpdates("hlfzsi", "LazyTea-mobile", os.Getenv("UIVERSION"))
		if err != nil {
			p.statusLabel.SetText("检查更新失败")
			p.statusLabel.Importance = widget.DangerImportance
			dialog.ShowError(fmt.Errorf("检查更新失败: %v", err), p.window)
			return
		}
		if hasUpdate {
			p.statusLabel.SetText("发现新版本")
			p.statusLabel.Importance = widget.HighImportance
			updateContent := container.NewVBox(
				widget.NewLabel(fmt.Sprintf("当前版本: %s", os.Getenv("UIVERSION"))),
				widget.NewLabel(fmt.Sprintf("最新版本: %s", release.TagName)),
				widget.NewSeparator(),
				widget.NewLabel("更新说明:"),
			)
			if release.Body != "" {
				releaseNotes := widget.NewRichTextFromMarkdown(release.Body)
				releaseNotes.Wrapping = fyne.TextWrapWord
				scrollContainer := container.NewScroll(releaseNotes)
				scrollContainer.SetMinSize(fyne.NewSize(300, 150))
				updateContent.Add(scrollContainer)
			}
			updateBtn := widget.NewButtonWithIcon("下载更新", fyneTheme.DownloadIcon(), func() {
				githubURL := fmt.Sprintf("https://github.com/hlfzsi/LazyTea-mobile/releases/tag/%s", release.TagName)
				dialog.ShowInformation("前往下载", fmt.Sprintf("请在浏览器中访问:\n%s", githubURL), p.window)
			})
			updateBtn.Importance = widget.HighImportance
			laterBtn := widget.NewButton("稍后提醒", func() {
			})
			buttonContainer := container.NewGridWithColumns(2, laterBtn, updateBtn)
			updateContent.Add(widget.NewSeparator())
			updateContent.Add(buttonContainer)
			updateDialog := dialog.NewCustom("发现新版本", "关闭", updateContent, p.window)
			updateDialog.Resize(fyne.NewSize(400, 300))
			updateDialog.Show()
		} else {
			p.statusLabel.SetText("已是最新版本")
			p.statusLabel.Importance = widget.SuccessImportance
			dialog.ShowInformation("检查更新", "当前已是最新版本", p.window)
		}
	}()
}
