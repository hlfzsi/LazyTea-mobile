package pages
import (
	"encoding/json"
	"fmt"
	"lazytea-mobile/internal/data"
	"lazytea-mobile/internal/network"
	"lazytea-mobile/internal/ui/components/roster"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)
type RosterPage struct {
	*PageBase
	data       map[string]interface{}
	config     *data.FullConfigModel
	onSave     func(data map[string]interface{})
	mainWindow fyne.Window
	targetBotID string  
	configTree       *roster.ConfigTree
	permissionPanel  *roster.PermissionPanel
	contentContainer *fyne.Container
	currentNode      *roster.TreeNode
	tabs             *container.AppTabs  
}
func NewRosterPage(initialData map[string]interface{}, onSave func(data map[string]interface{}), mainWindow fyne.Window, pageBase *PageBase) *RosterPage {
	var cfg *data.FullConfigModel
	if pageBase != nil && pageBase.logger != nil {
		pageBase.logger.Info("RosterPage初始化，接收到的数据键: %v", getMapKeysDebug(initialData))
		if jsonBytes, err := json.Marshal(initialData); err == nil {
			pageBase.logger.Info("RosterPage初始化数据: %s", string(jsonBytes))
		}
	}
	if len(initialData) == 0 {
		cfg = &data.FullConfigModel{Bots: make(map[string]data.BotModel)}
		if pageBase != nil && pageBase.logger != nil {
			pageBase.logger.Info("初始数据为空，将从服务端获取")
		}
	} else {
		parsed, err := data.ParseConfigFromMap(initialData)
		if err != nil {
			if pageBase != nil && pageBase.logger != nil {
				pageBase.logger.Error("解析配置失败: %v", err)
			}
			dialog.ShowError(fmt.Errorf("解析配置失败: %w", err), mainWindow)
			cfg = &data.FullConfigModel{Bots: make(map[string]data.BotModel)}
		} else {
			cfg = parsed
			if pageBase != nil && pageBase.logger != nil {
				pageBase.logger.Info("配置解析成功，Bot数量: %d", len(cfg.Bots))
				for botID, bot := range cfg.Bots {
					pageBase.logger.Info("Bot %s 包含 %d 个插件", botID, len(bot.Plugins))
				}
			}
		}
	}
	p := &RosterPage{
		PageBase:   pageBase,
		data:       initialData,
		config:     cfg,
		onSave:     onSave,
		mainWindow: mainWindow,
		targetBotID: "",  
	}
	p.setupUI()
	if len(initialData) == 0 && pageBase != nil && pageBase.client != nil {
		cb := &network.RequestCallback{
			Success: func(payload interface{}) {
				var m map[string]interface{}
				if pm, ok := payload.(map[string]interface{}); ok {
					if d, has := pm["data"]; has {
						if mm, ok := d.(map[string]interface{}); ok {
							m = mm
						} else {
							m = pm
						}
					} else {
						m = pm
					}
				}
				if m == nil {
					return
				}
				if parsed, err := data.ParseConfigFromMap(m); err == nil {
					p.data = m
					p.config = parsed
					if p.configTree != nil {
						p.configTree.UpdateConfig(parsed)
					}
					p.updatePermissionPanel(nil)
				} else {
					dialog.ShowError(fmt.Errorf("解析服务端配置失败: %v", err), p.mainWindow)
				}
			},
			Error: func(e error) {
				dialog.ShowError(fmt.Errorf("获取名单失败: %v", e), p.mainWindow)
			},
		}
		_ = pageBase.client.SendRequestWithCallback("get_matchers", map[string]interface{}{}, cb)
	}
	return p
}
func NewRosterPageForBot(initialData map[string]interface{}, onSave func(data map[string]interface{}), mainWindow fyne.Window, pageBase *PageBase, botID string) *RosterPage {
	var cfg *data.FullConfigModel
	if pageBase != nil && pageBase.logger != nil {
		pageBase.logger.Info("RosterPageForBot初始化（Bot: %s），接收到的数据键: %v", botID, getMapKeysDebug(initialData))
		if jsonBytes, err := json.Marshal(initialData); err == nil {
			pageBase.logger.Info("RosterPageForBot初始化数据: %s", string(jsonBytes))
		}
	}
	if len(initialData) == 0 {
		cfg = &data.FullConfigModel{Bots: make(map[string]data.BotModel)}
		if pageBase != nil && pageBase.logger != nil {
			pageBase.logger.Info("初始数据为空，将从服务端获取")
		}
	} else {
		parsed, err := data.ParseConfigFromMap(initialData)
		if err != nil {
			if pageBase != nil && pageBase.logger != nil {
				pageBase.logger.Error("解析配置失败: %v", err)
			}
			dialog.ShowError(fmt.Errorf("解析配置失败: %w", err), mainWindow)
			cfg = &data.FullConfigModel{Bots: make(map[string]data.BotModel)}
		} else {
			cfg = parsed
			if pageBase != nil && pageBase.logger != nil {
				pageBase.logger.Info("配置解析成功，Bot数量: %d", len(cfg.Bots))
				for botIDInCfg, bot := range cfg.Bots {
					pageBase.logger.Info("Bot %s 包含 %d 个插件", botIDInCfg, len(bot.Plugins))
				}
			}
		}
	}
	p := &RosterPage{
		PageBase:   pageBase,
		data:       initialData,
		config:     cfg,
		onSave:     onSave,
		mainWindow: mainWindow,
		targetBotID: botID,  
	}
	p.setupUI()
	if len(initialData) == 0 && pageBase != nil && pageBase.client != nil {
		cb := &network.RequestCallback{
			Success: func(payload interface{}) {
				var m map[string]interface{}
				if pm, ok := payload.(map[string]interface{}); ok {
					if d, has := pm["data"]; has {
						if mm, ok := d.(map[string]interface{}); ok {
							m = mm
						} else {
							m = pm
						}
					} else {
						m = pm
					}
				}
				if m == nil {
					return
				}
				if parsed, err := data.ParseConfigFromMap(m); err == nil {
					p.data = m
					p.config = parsed
					if p.configTree != nil {
						p.configTree.UpdateConfigForBot(parsed, botID)  
					}
					p.updatePermissionPanel(nil)
				} else {
					dialog.ShowError(fmt.Errorf("解析服务端配置失败: %v", err), p.mainWindow)
				}
			},
			Error: func(e error) {
				dialog.ShowError(fmt.Errorf("获取名单失败: %v", e), p.mainWindow)
			},
		}
		_ = pageBase.client.SendRequestWithCallback("get_matchers", map[string]interface{}{}, cb)
	}
	return p
}
func (p *RosterPage) setupUI() {
	titleLabel := widget.NewLabelWithStyle("权限配置", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	titleLabel.Importance = widget.HighImportance
	if p.targetBotID != "" {
		p.configTree = roster.NewConfigTreeForBot(p.config, p.targetBotID, func(node *roster.TreeNode) {
			if p.PageBase != nil && p.PageBase.logger != nil {
				if node != nil {
					p.PageBase.logger.Info("Tree节点选中: %s (类型: %d)", node.DisplayName, node.Type)
				} else {
					p.PageBase.logger.Info("Tree节点选中: nil")
				}
			}
			p.currentNode = node
			p.updatePermissionPanel(node)
			if node != nil && p.tabs != nil {
				p.tabs.Select(p.tabs.Items[1])  
			}
		})
	} else {
		p.configTree = roster.NewConfigTree(p.config, func(node *roster.TreeNode) {
			if p.PageBase != nil && p.PageBase.logger != nil {
				if node != nil {
					p.PageBase.logger.Info("Tree节点选中: %s (类型: %d)", node.DisplayName, node.Type)
				} else {
					p.PageBase.logger.Info("Tree节点选中: nil")
				}
			}
			p.currentNode = node
			p.updatePermissionPanel(node)
			if node != nil && p.tabs != nil {
				p.tabs.Select(p.tabs.Items[1])  
			}
		})
	}
	p.permissionPanel = roster.NewPermissionPanel(p.config, p.mainWindow)
	p.permissionPanel.SetDataChangedCallback(func() {
		if p.PageBase != nil && p.PageBase.logger != nil {
			p.PageBase.logger.Info("权限面板数据发生变更，已保存到本地配置")
		}
	})
	p.permissionPanel.SetEditModeChangedCallback(func(editMode bool) {
		if p.PageBase != nil && p.PageBase.logger != nil {
			p.PageBase.logger.Info("权限面板编辑模式变更: %v", editMode)
		}
	})
	p.contentContainer = container.NewVBox()
	placeholder := widget.NewLabel("👈 请从「配置树」Tab中选择要配置的项目")
	placeholder.Alignment = fyne.TextAlignCenter
	placeholder.Importance = widget.MediumImportance
	p.contentContainer.Add(placeholder)
	treeCard := widget.NewCard("", "", p.configTree)
	treeCard.SetTitle("Bot & 插件")
	contentScroll := container.NewScroll(p.contentContainer)
	contentScroll.SetMinSize(fyne.NewSize(300, 400))
	contentCard := widget.NewCard("", "", contentScroll)
	contentCard.SetTitle("详细设置")
	p.tabs = container.NewAppTabs(
		container.NewTabItem("📋 配置树", treeCard),
		container.NewTabItem("⚙️ 权限设置", contentCard),
	)
	p.tabs.SetTabLocation(container.TabLocationTop)
	p.tabs.Resize(fyne.NewSize(320, 480))
	mainContainer := container.NewBorder(
		container.NewVBox(titleLabel, widget.NewSeparator()),  
		p.createBottomBar(),  
		nil,                  
		nil,                  
		p.tabs,               
	)
	p.SetContent(mainContainer)
}
func (p *RosterPage) updatePermissionPanel(node *roster.TreeNode) {
	p.contentContainer.Objects = nil
	if node == nil {
		placeholder := widget.NewLabel("👈 请从「配置树」Tab中选择要配置的项目")
		placeholder.Alignment = fyne.TextAlignCenter
		placeholder.Importance = widget.MediumImportance
		p.contentContainer.Add(placeholder)
		p.contentContainer.Refresh()
		return
	}
	breadcrumb := widget.NewRichTextFromMarkdown(fmt.Sprintf("**当前选择:** %s", node.DisplayName))
	breadcrumb.Wrapping = fyne.TextWrapWord
	p.contentContainer.Add(breadcrumb)
	p.contentContainer.Add(widget.NewSeparator())
	switch node.Type {
	case roster.NodeTypeBot:
		p.createBotPanel(node)
	case roster.NodeTypePlugin:
		p.createPluginPanel(node)
	case roster.NodeTypeMatcher:
		p.permissionPanel.UpdateNode(node)
		p.contentContainer.Add(p.permissionPanel)
	}
	p.contentContainer.Refresh()
}
func (p *RosterPage) createBotPanel(node *roster.TreeNode) {
	title := widget.NewLabelWithStyle(fmt.Sprintf("🤖 Bot: %s", node.DisplayName), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance
	info := widget.NewLabel(fmt.Sprintf("Bot ID: %s", node.BotID))
	info.Importance = widget.MediumImportance
	pluginCount := 0
	matcherCount := 0
	if botCfg, ok := p.config.Bots[node.BotID]; ok {
		pluginCount = len(botCfg.Plugins)
		for _, plugin := range botCfg.Plugins {
			matcherCount += len(plugin.Matchers)
		}
	}
	statsCard := widget.NewCard(
		"📊 统计信息",
		"",
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("插件数量: %d", pluginCount)),
			widget.NewLabel(fmt.Sprintf("匹配器数量: %d", matcherCount)),
			widget.NewSeparator(),
			widget.NewLabel("💡 提示: 展开左侧树形结构选择具体插件或匹配器进行权限配置"),
		),
	)
	p.contentContainer.Add(title)
	p.contentContainer.Add(info)
	p.contentContainer.Add(widget.NewSeparator())
	p.contentContainer.Add(statsCard)
}
func (p *RosterPage) createPluginPanel(node *roster.TreeNode) {
	title := widget.NewLabelWithStyle(fmt.Sprintf("🔌 插件: %s", node.DisplayName), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance
	info := widget.NewLabel(fmt.Sprintf("所属Bot: %s", node.BotID))
	info.Importance = widget.MediumImportance
	matchersLabel := widget.NewLabelWithStyle("⚡ 功能匹配器", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	var matcherDisplayNames []string
	if botCfg, ok := p.config.Bots[node.BotID]; ok {
		if plg, ok := botCfg.Plugins[node.PluginName]; ok {
			for _, matcher := range plg.Matchers {
				matcherDisplayNames = append(matcherDisplayNames, data.GetRuleDisplayName(matcher.Rule))
			}
		}
	}
	if len(matcherDisplayNames) == 0 {
		emptyLabel := widget.NewLabel("📭 此插件暂无匹配器")
		emptyLabel.Alignment = fyne.TextAlignCenter
		p.contentContainer.Add(title)
		p.contentContainer.Add(info)
		p.contentContainer.Add(widget.NewSeparator())
		p.contentContainer.Add(matchersLabel)
		p.contentContainer.Add(emptyLabel)
		return
	}
	matchersContainer := container.NewVBox()
	for i, matcherName := range matcherDisplayNames {
		matcherCard := widget.NewCard(
			fmt.Sprintf("🎯 匹配器 %d", i+1),
			matcherName,
			container.NewHBox(
				func() *widget.Button {
					btn := widget.NewButtonWithIcon("配置权限", theme.SettingsIcon(), func(idx int) func() {
						return func() {
							matcherNode := &roster.TreeNode{
								ID:          fmt.Sprintf("%s_matcher_%d", node.ID, idx),
								DisplayName: matcherDisplayNames[idx],
								Type:        roster.NodeTypeMatcher,
								BotID:       node.BotID,
								PluginName:  node.PluginName,
								MatcherIdx:  idx,
							}
							p.updatePermissionPanel(matcherNode)
						}
					}(i))
					btn.Importance = widget.HighImportance  
					return btn
				}(),
			),
		)
		matchersContainer.Add(matcherCard)
	}
	scroll := container.NewVScroll(matchersContainer)
	scroll.SetMinSize(fyne.NewSize(0, 300))
	p.contentContainer.Add(title)
	p.contentContainer.Add(info)
	p.contentContainer.Add(widget.NewSeparator())
	p.contentContainer.Add(matchersLabel)
	p.contentContainer.Add(scroll)
}
func (p *RosterPage) createBottomBar() *fyne.Container {
	saveBtn := widget.NewButtonWithIcon("保存", theme.DocumentSaveIcon(), func() {
		p.SaveConfig()
	})
	saveBtn.Importance = widget.HighImportance
	return container.NewHBox(
		layout.NewSpacer(),
		saveBtn,
	)
}
func (p *RosterPage) UpdateConfig(newData map[string]interface{}) {
	config, err := data.ParseConfigFromMap(newData)
	if err != nil {
		dialog.ShowError(fmt.Errorf("解析配置失败: %w", err), p.mainWindow)
		return
	}
	p.data = newData
	p.config = config
	p.configTree.UpdateConfig(config)
	p.updatePermissionPanel(nil)  
}
func (p *RosterPage) GetModifiedData() map[string]interface{} {
	if p.config == nil {
		if p.PageBase != nil && p.PageBase.logger != nil {
			p.PageBase.logger.Warn("GetModifiedData: config为空，返回原始数据")
		}
		return p.data
	}
	newData, err := p.config.ToMap()
	if err != nil {
		if p.PageBase != nil && p.PageBase.logger != nil {
			p.PageBase.logger.Error("GetModifiedData: 转换配置失败: %v", err)
		}
		return p.data
	}
	if p.PageBase != nil && p.PageBase.logger != nil {
		p.PageBase.logger.Info("GetModifiedData: 成功获取修改后的数据")
	}
	return newData
}
func (p *RosterPage) SaveConfig() {
	if p.PageBase != nil && p.PageBase.logger != nil {
		p.PageBase.logger.Info("SaveConfig: 开始保存配置")
	}
	if p.onSave != nil {
		p.onSave(p.GetModifiedData())
	}
	dataMap := p.GetModifiedData()
	if p.PageBase != nil && p.PageBase.logger != nil {
		if len(dataMap) > 0 {
			p.PageBase.logger.Info("SaveConfig: 准备发送的数据包含 %d 个顶级键", len(dataMap))
			if bots, ok := dataMap["bots"].(map[string]interface{}); ok {
				p.PageBase.logger.Info("SaveConfig: 数据包含 %d 个Bot", len(bots))
			}
		} else {
			p.PageBase.logger.Warn("SaveConfig: 数据为空")
		}
	}
	jsonBytes, err := json.Marshal(dataMap)
	if err != nil {
		if p.PageBase != nil && p.PageBase.logger != nil {
			p.PageBase.logger.Error("SaveConfig: 序列化配置失败: %v", err)
		}
		dialog.ShowError(fmt.Errorf("序列化配置失败: %v", err), p.mainWindow)
		return
	}
	if p.PageBase != nil && p.PageBase.logger != nil {
		p.PageBase.logger.Info("SaveConfig: JSON序列化成功，长度: %d 字节", len(jsonBytes))
	}
	params := map[string]interface{}{
		"new_roster": string(jsonBytes),
	}
	callback := &network.RequestCallback{
		Success: func(payload interface{}) {
			if p.PageBase != nil && p.PageBase.logger != nil {
				p.PageBase.logger.Info("SaveConfig: 服务端同步成功")
			}
			dialog.ShowInformation("保存成功", "配置已成功同步！", p.mainWindow)
		},
		Error: func(e error) {
			if p.PageBase != nil && p.PageBase.logger != nil {
				p.PageBase.logger.Error("SaveConfig: 服务端同步失败: %v", e)
			}
			dialog.ShowError(fmt.Errorf("同步失败: %v", e), p.mainWindow)
		},
	}
	if err := p.client.SendRequestWithCallback("sync_matchers", params, callback); err != nil {
		if p.PageBase != nil && p.PageBase.logger != nil {
			p.PageBase.logger.Error("SaveConfig: 请求发送失败: %v", err)
		}
		dialog.ShowError(fmt.Errorf("请求发送失败: %v", err), p.mainWindow)
		return
	}
}
func getMapKeysDebug(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
