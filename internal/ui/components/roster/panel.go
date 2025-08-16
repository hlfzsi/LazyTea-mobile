package roster
import (
	"fmt"
	"lazytea-mobile/internal/data"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)
type PermissionPanel struct {
	widget.BaseWidget
	config            *data.FullConfigModel
	mainWindow        fyne.Window
	content           *fyne.Container
	currentNode       *TreeNode
	onDataChanged     func()      
	editMode          bool        
	onEditModeChanged func(bool)  
	currentTabIndex   int         
}
func NewPermissionPanel(config *data.FullConfigModel, mainWindow fyne.Window) *PermissionPanel {
	panel := &PermissionPanel{
		config:     config,
		mainWindow: mainWindow,
		content:    container.NewVBox(),
	}
	panel.ExtendBaseWidget(panel)
	panel.setupUI()
	return panel
}
func (p *PermissionPanel) SetDataChangedCallback(callback func()) {
	p.onDataChanged = callback
}
func (p *PermissionPanel) SetEditModeChangedCallback(callback func(bool)) {
	p.onEditModeChanged = callback
}
func (p *PermissionPanel) setupUI() {
	placeholder := widget.NewLabel("请选择要配置的项目")
	placeholder.Alignment = fyne.TextAlignCenter
	placeholder.Importance = widget.MediumImportance
	p.content.Add(placeholder)
}
func (p *PermissionPanel) UpdateNode(node *TreeNode) {
	p.currentNode = node
	p.updateContent()
}
func (p *PermissionPanel) updateContent() {
	p.content.Objects = nil
	if p.currentNode == nil {
		placeholder := widget.NewLabel("请选择要配置的项目")
		placeholder.Alignment = fyne.TextAlignCenter
		placeholder.Importance = widget.MediumImportance
		p.content.Add(placeholder)
		p.content.Refresh()
		return
	}
	switch p.currentNode.Type {
	case NodeTypeBot:
		p.createBotContent()
	case NodeTypePlugin:
		p.createPluginContent()
	case NodeTypeMatcher:
		if p.editMode {
			p.createMatcherEditContent()
		} else {
			p.createMatcherContent()
		}
	}
	p.content.Refresh()
}
func (p *PermissionPanel) createBotContent() {
	title := widget.NewLabelWithStyle(fmt.Sprintf("Bot: %s", p.currentNode.DisplayName), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance
	info := widget.NewLabel(fmt.Sprintf("Bot ID: %s", p.currentNode.BotID))
	info.Importance = widget.MediumImportance
	stats := p.getBotStats(p.currentNode.BotID)
	statsLabel := widget.NewLabel(fmt.Sprintf("插件数量: %d | 匹配器数量: %d", stats.pluginCount, stats.matcherCount))
	p.content.Add(title)
	p.content.Add(widget.NewSeparator())
	p.content.Add(info)
	p.content.Add(statsLabel)
}
func (p *PermissionPanel) createPluginContent() {
	title := widget.NewLabelWithStyle(p.currentNode.DisplayName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance
	info := widget.NewLabel(fmt.Sprintf("Bot: %s", p.currentNode.BotID))
	info.Importance = widget.MediumImportance
	matchers := p.getPluginMatchers(p.currentNode.BotID, p.currentNode.PluginName)
	matchersLabel := widget.NewLabelWithStyle("匹配器列表", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	matchersList := widget.NewList(
		func() int { return len(matchers) },
		func() fyne.CanvasObject {
			return widget.NewLabel("匹配器")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(matchers) {
				matcher := matchers[id]
				obj.(*widget.Label).SetText(matcher.displayName)
			}
		},
	)
	p.content.Add(title)
	p.content.Add(widget.NewSeparator())
	p.content.Add(info)
	p.content.Add(matchersLabel)
	p.content.Add(matchersList)
}
func (p *PermissionPanel) createMatcherContent() {
	title := widget.NewLabelWithStyle(fmt.Sprintf("匹配器: %s", p.currentNode.DisplayName), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance
	botInfo := widget.NewLabel(fmt.Sprintf("Bot: %s", p.currentNode.BotID))
	pluginInfo := widget.NewLabel(fmt.Sprintf("插件: %s", p.currentNode.PluginName))
	botInfo.Importance = widget.MediumImportance
	pluginInfo.Importance = widget.MediumImportance
	infoContainer := container.NewVBox(botInfo, pluginInfo)
	matcherConfig := p.getMatcherConfig(p.currentNode.BotID, p.currentNode.PluginName, p.currentNode.MatcherIdx)
	enableCheck := widget.NewCheck("启用此匹配器", func(checked bool) {
		p.updateMatcherStatus(checked)
	})
	if matcherConfig != nil {
		enableCheck.SetChecked(matcherConfig.IsOn)
	}
	permissionStats := p.getPermissionStats(matcherConfig)
	whiteListStats := widget.NewLabelWithStyle(fmt.Sprintf("✅ 白名单: 用户 %d 个, 群组 %d 个",
		permissionStats.whiteUsers, permissionStats.whiteGroups), fyne.TextAlignLeading, fyne.TextStyle{})
	blackListStats := widget.NewLabelWithStyle(fmt.Sprintf("❌ 黑名单: 用户 %d 个, 群组 %d 个",
		permissionStats.banUsers, permissionStats.banGroups), fyne.TextAlignLeading, fyne.TextStyle{})
	statsContainer := container.NewVBox(
		widget.NewLabelWithStyle("📊 权限统计", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		whiteListStats,
		blackListStats,
	)
	editBtn := widget.NewButtonWithIcon("编辑权限", theme.SettingsIcon(), func() {
		p.showPermissionEditor()
	})
	editBtn.Importance = widget.HighImportance
	p.content.Add(title)
	p.content.Add(widget.NewSeparator())
	p.content.Add(infoContainer)
	p.content.Add(enableCheck)
	p.content.Add(statsContainer)
	p.content.Add(widget.NewSeparator())
	p.content.Add(editBtn)
}
func (p *PermissionPanel) createMatcherEditContent() {
	title := widget.NewLabelWithStyle(fmt.Sprintf("🎯 编辑权限 - %s", p.currentNode.DisplayName), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance
	botInfo := widget.NewLabel(fmt.Sprintf("Bot: %s", p.currentNode.BotID))
	pluginInfo := widget.NewLabel(fmt.Sprintf("插件: %s", p.currentNode.PluginName))
	botInfo.Importance = widget.MediumImportance
	pluginInfo.Importance = widget.MediumImportance
	infoContainer := container.NewVBox(botInfo, pluginInfo)
	matcherConfig := p.getMatcherConfig(p.currentNode.BotID, p.currentNode.PluginName, p.currentNode.MatcherIdx)
	if matcherConfig == nil {
		p.content.Add(widget.NewLabel("❌ 无法获取匹配器配置"))
		return
	}
	backBtn := widget.NewButtonWithIcon("返回上层", theme.NavigateBackIcon(), func() {
		p.editMode = false
		if p.onEditModeChanged != nil {
			p.onEditModeChanged(false)
		}
		p.updateContent()
	})
	backBtn.Importance = widget.MediumImportance
	tabs := container.NewAppTabs()
	whiteListContent := p.createInlinePermissionEditor("白名单", &matcherConfig.Permission.WhiteList)
	tabs.Append(container.NewTabItem("✅ 白名单", whiteListContent))
	blackListContent := p.createInlinePermissionEditor("黑名单", &matcherConfig.Permission.BanList)
	tabs.Append(container.NewTabItem("❌ 黑名单", blackListContent))
	if p.currentTabIndex < len(tabs.Items) {
		tabs.SelectTab(tabs.Items[p.currentTabIndex])
	}
	tabs.OnChanged = func(tab *container.TabItem) {
		for i, item := range tabs.Items {
			if item == tab {
				p.currentTabIndex = i
				break
			}
		}
	}
	tabs.Resize(fyne.NewSize(300, 500))  
	p.content.Add(title)
	p.content.Add(widget.NewSeparator())
	p.content.Add(infoContainer)
	p.content.Add(backBtn)
	p.content.Add(widget.NewSeparator())
	p.content.Add(tabs)
	p.content.Add(widget.NewSeparator())
}
func (p *PermissionPanel) getBotStats(botID string) struct {
	pluginCount  int
	matcherCount int
} {
	stats := struct {
		pluginCount  int
		matcherCount int
	}{}
	if bot, exists := p.config.Bots[botID]; exists {
		stats.pluginCount = len(bot.Plugins)
		for _, plugin := range bot.Plugins {
			stats.matcherCount += len(plugin.Matchers)
		}
	}
	return stats
}
func (p *PermissionPanel) getPluginMatchers(botID, pluginName string) []struct {
	displayName string
	index       int
} {
	var matchers []struct {
		displayName string
		index       int
	}
	if bot, exists := p.config.Bots[botID]; exists {
		if plugin, exists := bot.Plugins[pluginName]; exists {
			for i, matcher := range plugin.Matchers {
				displayName := p.getRuleDisplayName(matcher.Rule)
				matchers = append(matchers, struct {
					displayName string
					index       int
				}{displayName: displayName, index: i})
			}
		}
	}
	return matchers
}
func (p *PermissionPanel) getMatcherConfig(botID, pluginName string, matcherIdx int) *data.MatcherRuleModel {
	if bot, exists := p.config.Bots[botID]; exists {
		if plugin, exists := bot.Plugins[pluginName]; exists {
			if matcherIdx < len(plugin.Matchers) {
				return &plugin.Matchers[matcherIdx]
			}
		}
	}
	return nil
}
func (p *PermissionPanel) getPermissionStats(matcher *data.MatcherRuleModel) struct {
	whiteUsers  int
	whiteGroups int
	banUsers    int
	banGroups   int
} {
	stats := struct {
		whiteUsers  int
		whiteGroups int
		banUsers    int
		banGroups   int
	}{}
	if matcher != nil {
		stats.whiteUsers = len(matcher.Permission.WhiteList.User)
		stats.whiteGroups = len(matcher.Permission.WhiteList.Group)
		stats.banUsers = len(matcher.Permission.BanList.User)
		stats.banGroups = len(matcher.Permission.BanList.Group)
	}
	return stats
}
func (p *PermissionPanel) getRuleDisplayName(rule map[string]interface{}) string {
	if commands, ok := rule["commands"].([]interface{}); ok && len(commands) > 0 {
		return fmt.Sprintf("命令: %v", commands[0])
	}
	if keywords, ok := rule["keywords"].([]interface{}); ok && len(keywords) > 0 {
		return fmt.Sprintf("关键词: %v", keywords[0])
	}
	return "通用规则"
}
func (p *PermissionPanel) updateMatcherStatus(isOn bool) {
	if p.currentNode == nil || p.currentNode.Type != NodeTypeMatcher {
		return
	}
	if bot, exists := p.config.Bots[p.currentNode.BotID]; exists {
		if plugin, exists := bot.Plugins[p.currentNode.PluginName]; exists {
			if p.currentNode.MatcherIdx < len(plugin.Matchers) {
				plugin.Matchers[p.currentNode.MatcherIdx].IsOn = isOn
				if p.onDataChanged != nil {
					p.onDataChanged()
				}
			}
		}
	}
}
func (p *PermissionPanel) createInlinePermissionEditor(_ string, permList *data.PermissionListDivide) *fyne.Container {
	mainContainer := container.NewBorder(
		nil,                             
		p.createQuickActions(permList),  
		nil,                             
		nil,                             
		container.NewVBox(  
			p.createPermissionSection("👤 用户", &permList.User),
			widget.NewSeparator(),
			p.createPermissionSection("👥 群组", &permList.Group),
		),
	)
	return mainContainer
}
func (p *PermissionPanel) createQuickActions(permList *data.PermissionListDivide) *fyne.Container {
	return container.NewHBox(
		widget.NewButtonWithIcon("清空用户", theme.ContentClearIcon(), func() {
			dialog.ShowConfirm("确认清空", "确定要清空所有用户吗？", func(confirmed bool) {
				if confirmed {
					permList.User = []string{}
					if p.onDataChanged != nil {
						p.onDataChanged()
					}
					if p.editMode {
						p.refreshEditContentPreserveTabs()
					} else {
						p.updateContent()
					}
				}
			}, p.mainWindow)
		}),
		widget.NewButtonWithIcon("清空群组", theme.ContentClearIcon(), func() {
			dialog.ShowConfirm("确认清空", "确定要清空所有群组吗？", func(confirmed bool) {
				if confirmed {
					permList.Group = []string{}
					if p.onDataChanged != nil {
						p.onDataChanged()
					}
					if p.editMode {
						p.refreshEditContentPreserveTabs()
					} else {
						p.updateContent()
					}
				}
			}, p.mainWindow)
		}),
	)
}
func (p *PermissionPanel) createPermissionSection(sectionTitle string, list *[]string) *fyne.Container {
	titleLabel := widget.NewLabelWithStyle(sectionTitle, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	entry := widget.NewEntry()
	entry.SetPlaceHolder("输入ID")
	addBtn := widget.NewButtonWithIcon("添加", theme.ContentAddIcon(), func() {
		value := entry.Text
		if value != "" && !p.containsString(*list, value) {
			*list = append(*list, value)
			entry.SetText("")
			if p.onDataChanged != nil {
				p.onDataChanged()
			}
			if p.editMode {
				p.refreshEditContentPreserveTabs()
			} else {
				p.updateContent()
			}
		}
	})
	addContainer := container.NewBorder(nil, nil, nil, addBtn, entry)
	var listWidget fyne.CanvasObject
	if len(*list) > 0 {
		listWidget = widget.NewList(
			func() int { return len(*list) },
			func() fyne.CanvasObject {
				return container.NewBorder(
					nil, nil, nil,
					widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
					widget.NewLabel(""),
				)
			},
			func(id widget.ListItemID, obj fyne.CanvasObject) {
				if id < len(*list) {
					item := (*list)[id]
					container := obj.(*fyne.Container)
					for _, child := range container.Objects {
						if label, ok := child.(*widget.Label); ok {
							label.SetText(item)
							break
						}
					}
					if borderContainer, ok := obj.(*fyne.Container); ok {
						if deleteBtn, ok := borderContainer.Objects[len(borderContainer.Objects)-1].(*widget.Button); ok {
							deleteBtn.OnTapped = func() {
								dialog.ShowConfirm("确认删除", fmt.Sprintf("确定要删除 '%s' 吗？", item), func(confirmed bool) {
									if confirmed && id < len(*list) {
										*list = append((*list)[:id], (*list)[id+1:]...)
										if p.onDataChanged != nil {
											p.onDataChanged()
										}
										if p.editMode {
											p.refreshEditContentPreserveTabs()
										} else {
											p.updateContent()
										}
									}
								}, p.mainWindow)
							}
						}
					}
				}
			},
		)
	} else {
		listWidget = widget.NewLabel("📭 暂无数据")
		listWidget.(*widget.Label).Alignment = fyne.TextAlignCenter
	}
	return container.NewBorder(
		container.NewVBox(titleLabel, addContainer, widget.NewSeparator()),  
		nil,         
		nil,         
		nil,         
		listWidget,  
	)
}
func (p *PermissionPanel) refreshEditContentPreserveTabs() {
	if p.currentNode != nil && p.currentNode.Type == NodeTypeMatcher && p.editMode {
		p.content.Objects = nil
		p.createMatcherEditContent()
		p.content.Refresh()
	}
}
func (p *PermissionPanel) showPermissionEditor() {
	if p.currentNode == nil || p.currentNode.Type != NodeTypeMatcher {
		return
	}
	p.editMode = true
	if p.onEditModeChanged != nil {
		p.onEditModeChanged(true)
	}
	p.updateContent()
}
func (p *PermissionPanel) containsString(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}
func (p *PermissionPanel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.content)
}
