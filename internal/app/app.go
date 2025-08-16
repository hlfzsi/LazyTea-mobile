package app
import (
	"lazytea-mobile/internal/config"
	"lazytea-mobile/internal/data"
	"lazytea-mobile/internal/mobile"
	"lazytea-mobile/internal/network"
	"lazytea-mobile/internal/ui/pages"
	"lazytea-mobile/internal/utils"
	"log"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fyneTheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)
type App struct {
	fyneApp fyne.App
	window  fyne.Window
	client  *network.Client
	storage *data.Storage
	tabs    *container.AppTabs
	logger  *utils.Logger
	config  *config.Config
	mobileLifecycle   *mobile.MobileLifecycle
	permissionManager *mobile.PermissionManager
	networkManager    *mobile.MobileNetworkManager
	overviewPage *pages.OverviewPage
	botInfoPage  *pages.BotInfoPage
	messagePage  *pages.MessagePage
	pluginPage   *pages.PluginPage
	settingsPage *pages.SettingsPage
}
func NewApp(fyneApp fyne.App) *App {
	app := &App{
		fyneApp: fyneApp,
		logger:  utils.NewLogger(),
	}
	config.SetApp(fyneApp)
	var err error
	app.config, err = config.Load()
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		app.config = config.GetGlobal()
	}
	app.logger.Info("Requesting permissions at startup...")
	app.permissionManager = mobile.CheckAndRequestPermissions()
	status := app.permissionManager.GetPermissionStatus()
	for perm, granted := range status {
		if granted {
			app.logger.Info("Permission %s: GRANTED", perm)
		} else {
			app.logger.Warn("Permission %s: DENIED", perm)
		}
	}
	app.networkManager = mobile.NewMobileNetworkManager()
	if mobile.IsMobile() {
		app.logger.Info("Mobile environment detected, configuring mobile features...")
		defer func() {
			if r := recover(); r != nil {
				app.logger.Error("Mobile configuration panic recovered: %v", r)
			}
		}()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					app.logger.Error("Delayed mobile UI configuration panic recovered: %v", r)
				}
			}()
			mobile.ConfigureForMobile(fyneApp)
			mobile.ApplyMobileTheme(fyneApp)
			app.mobileLifecycle = mobile.NewMobileLifecycle(fyneApp)
			app.setupMobileNetworking()
			app.logger.Info("Mobile UI configuration completed")
		}()
	}
	var dbPath string
	uri, err := config.GetDatabaseURI()
	if err != nil {
		app.logger.Error("Failed to get database URI: %v", err)
		dbPath = "data.db"
	} else {
		dbPath = uri.Path()
	}
	app.storage, err = data.NewStorage(dbPath)
	if err != nil {
		app.logger.Error("Failed to initialize storage: %v", err)
	}
	app.client = network.NewClient(app.logger)
	return app
}
func (a *App) Run() {
	a.setupWindow()
	a.setupPages()
	a.setupLayout()
	a.tryAutoConnect()
	a.window.ShowAndRun()
}
func (a *App) setupWindow() {
	a.window = a.fyneApp.NewWindow("LazyTea Mobile")
	a.window.SetIcon(fyne.NewStaticResource("icon", []byte{}))  
	if mobile.IsMobile() {
		a.window.SetFullScreen(true)
	} else {
		a.window.Resize(fyne.NewSize(390, 680))
		a.window.SetFixedSize(false)
	}
	a.window.SetContent(widget.NewLabel("Loading..."))
}
func (a *App) setupPages() {
	a.overviewPage = pages.NewOverviewPage(a.client, a.storage, a.logger, a.config)
	a.botInfoPage = pages.NewBotInfoPage(a.client, a.storage, a.logger, a.window)
	a.messagePage = pages.NewMessagePage(a.client, a.storage, a.logger)
	a.pluginPage = pages.NewPluginPage(a.client, a.storage, a.logger)
	a.settingsPage = pages.NewSettingsPage(a.client, a.storage, a.logger, a.window, a.config)
}
func (a *App) setupLayout() {
	a.tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("概览", fyneTheme.HomeIcon(), a.overviewPage.GetContent()),
		container.NewTabItemWithIcon("Bot", fyneTheme.ComputerIcon(), a.botInfoPage.GetContent()),
		container.NewTabItemWithIcon("消息", fyneTheme.MailComposeIcon(), a.messagePage.GetContent()),
		container.NewTabItemWithIcon("插件", fyneTheme.FolderIcon(), a.pluginPage.GetContent()),
		container.NewTabItemWithIcon("设置", fyneTheme.SettingsIcon(), a.settingsPage.GetContent()),
	)
	a.tabs.SetTabLocation(container.TabLocationBottom)
	a.setupTabChangeHandlers()
	statusBar := a.createStatusBar()
	content := container.NewBorder(
		statusBar,  
		nil,        
		nil,        
		nil,        
		a.tabs,     
	)
	a.window.SetContent(content)
}
func (a *App) createStatusBar() *fyne.Container {
	statusLabel := widget.NewLabel("未连接")
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	statusLabel.Importance = widget.DangerImportance
	a.client.OnConnectionChanged(func(connected bool) {
		if connected {
			statusLabel.SetText("✓ 已连接")
			statusLabel.Importance = widget.SuccessImportance
		} else {
			statusLabel.SetText("✗ 未连接")
			statusLabel.Importance = widget.DangerImportance
		}
	})
	statusContainer := container.NewHBox(
		widget.NewIcon(fyneTheme.InfoIcon()),
		statusLabel,
	)
	return container.NewBorder(
		nil,                                   
		widget.NewSeparator(),                 
		nil,                                   
		nil,                                   
		container.NewPadded(statusContainer),  
	)
}
func (a *App) tryAutoConnect() {
	if !a.config.Network.AutoConnect || a.config.Network.Host == "" {
		a.logger.Info("自动连接未启用或无连接配置")
		return
	}
	a.logger.Info("尝试自动连接到 %s:%d", a.config.Network.Host, a.config.Network.Port)
	go func() {
		if err := a.client.Connect(a.config.Network.Host, a.config.Network.Port, a.config.Network.Token); err != nil {
			a.logger.Error("自动连接失败: %v", err)
		}
	}()
}
func (a *App) setupTabChangeHandlers() {
	var currentTabIndex int = 0
	a.tabs.OnSelected = func(tab *container.TabItem) {
		newIndex := -1
		for i, item := range a.tabs.Items {
			if item == tab {
				newIndex = i
				break
			}
		}
		if newIndex == -1 || newIndex == currentTabIndex {
			return
		}
		a.callPageLeave(currentTabIndex)
		a.callPageEnter(newIndex)
		currentTabIndex = newIndex
	}
}
func (a *App) callPageEnter(pageIndex int) {
	switch pageIndex {
	case 0:  
	case 1:  
	case 2:  
	case 3:  
		a.pluginPage.OnEnter()
	}
}
func (a *App) callPageLeave(pageIndex int) {
	switch pageIndex {
	case 0:  
	case 1:  
	case 2:  
	case 3:  
	}
}
func (a *App) setupMobileNetworking() {
	if !mobile.IsMobile() {
		return
	}
	a.networkManager.SetStateChangeCallback(func(state mobile.NetworkState) {
		switch state {
		case mobile.NetworkStateConnected:
			a.logger.Info("[Mobile] Network connected")
		case mobile.NetworkStateDisconnected:
			a.logger.Info("[Mobile] Network disconnected")
		case mobile.NetworkStateConnecting:
			a.logger.Info("[Mobile] Network connecting")
		}
	})
	a.networkManager.SetReconnectCallback(func() {
		a.logger.Info("[Mobile] Attempting network reconnection")
		go func() {
			if err := a.client.Connect(a.config.Network.Host, a.config.Network.Port, a.config.Network.Token); err != nil {
				a.logger.Error("[Mobile] Reconnection failed: %v", err)
				a.networkManager.OnConnectionLost()
			} else {
				a.logger.Info("[Mobile] Reconnection successful")
				a.networkManager.OnConnectionEstablished()
			}
		}()
	})
	a.mobileLifecycle.SetBackgroundCallback(func() {
		a.logger.Info("[Mobile] App going to background")
		a.networkManager.HandleAppBackground()
	})
	a.mobileLifecycle.SetForegroundCallback(func() {
		a.logger.Info("[Mobile] App coming to foreground")
		a.networkManager.HandleAppForeground()
	})
	a.mobileLifecycle.SetLowMemoryCallback(func() {
		a.logger.Warn("[Mobile] Low memory warning")
	})
	a.client.OnConnectionChanged(func(connected bool) {
		if connected {
			a.networkManager.OnConnectionEstablished()
		} else {
			a.networkManager.OnConnectionLost()
		}
	})
	a.logger.Info("[Mobile] Mobile networking configured")
}
