package pages
import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"lazytea-mobile/internal/data"
	"lazytea-mobile/internal/network"
	"lazytea-mobile/internal/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	fyneTheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)
type PluginPage struct {
	*PageBase
	pluginContainer *fyne.Container  
	plugins         []data.Plugin
	allPlugins      []data.Plugin
	searchEntry     *widget.Entry
	statusLabel     *widget.Label
	emptyLabel      *widget.Label  
	isSearching     bool
	configView      *fyne.Container  
	showingConfig   bool             
	currentPlugin   data.Plugin      
	backButton      *widget.Button   
	mainContent     *fyne.Container  
	listView        *fyne.Container  
}
func NewPluginPage(client *network.Client, storage *data.Storage, logger *utils.Logger) *PluginPage {
	page := &PluginPage{
		PageBase:   NewPageBase(client, storage, logger),
		plugins:    make([]data.Plugin, 0),
		allPlugins: make([]data.Plugin, 0),
	}
	page.setupUI()
	page.setupEventHandlers()
	return page
}
func (p *PluginPage) setupUI() {
	titleLabel := widget.NewLabelWithStyle("插件管理", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	p.statusLabel = widget.NewLabel("正在加载...")
	p.statusLabel.Alignment = fyne.TextAlignCenter
	p.searchEntry = widget.NewEntry()
	p.searchEntry.SetPlaceHolder("搜索插件名称...")
	p.searchEntry.OnChanged = func(text string) {
		p.filterPlugins(text)
	}
	backBtn := widget.NewButtonWithIcon("返回", fyneTheme.NavigateBackIcon(), func() {
		p.hideConfigView()
	})
	backBtn.Hide()  
	searchContainer := container.NewBorder(
		nil, nil, backBtn, nil,
		p.searchEntry,
	)
	controlContainer := container.NewHBox(
		widget.NewLabel("状态:"),
		p.statusLabel,
	)
	p.pluginContainer = container.NewVBox()
	pluginScroll := container.NewScroll(p.pluginContainer)
	p.emptyLabel = widget.NewLabel("暂无插件信息")
	p.emptyLabel.Alignment = fyne.TextAlignCenter
	p.emptyLabel.Hide()  
	topZone := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		searchContainer,
		controlContainer,
		widget.NewSeparator(),
	)
	p.configView = container.NewVBox()
	p.listView = container.NewStack(
		pluginScroll,
		p.emptyLabel,
	)
	p.mainContent = container.NewBorder(nil, nil, nil, nil, p.listView)
	content := container.NewBorder(
		topZone,
		nil,
		nil,
		nil,
		p.mainContent,
	)
	p.backButton = backBtn
	p.SetContent(content)
}
func (p *PluginPage) refreshPluginContainer() {
	p.pluginContainer.Objects = nil
	for i, plugin := range p.plugins {
		pluginCard := p.createPluginCard(plugin, i)
		p.pluginContainer.Add(pluginCard)
	}
	p.pluginContainer.Refresh()
	p.updateEmptyState()
}
func (p *PluginPage) createPluginCard(plugin data.Plugin, _ int) fyne.CanvasObject {
	iconLabel := widget.NewLabel("📦")
	iconLabel.TextStyle = fyne.TextStyle{Bold: true}
	iconLabel.Alignment = fyne.TextAlignCenter
	if plugin.Meta.IconAbspath != "" {
		iconLabel.SetText("⚡")
	}
	displayName := plugin.Meta.Name
	if displayName == "" {
		displayName = plugin.Name
	}
	nameLabel := widget.NewLabel(displayName)
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}
	updateLabel := widget.NewLabel("🔔")
	updateLabel.TextStyle = fyne.TextStyle{Bold: true}
	updateLabel.Importance = widget.HighImportance
	updateLabel.Hide()
	configBtn := widget.NewButtonWithIcon("", fyneTheme.SettingsIcon(), func() {
		p.showInlinePluginConfig(plugin)
	})
	configBtn.Importance = widget.LowImportance
	desc := plugin.Meta.Description
	if len(desc) > 100 {
		desc = desc[:100] + "..."
	}
	if desc == "" {
		desc = "暂无描述"
	}
	descLabel := widget.NewLabel(desc)
	descLabel.TextStyle = fyne.TextStyle{Italic: true}
	descLabel.Wrapping = fyne.TextWrapWord
	version := plugin.Meta.Version
	if version == "" || version == "未知版本" {
		version = "未知"
	}
	author := plugin.Meta.Author
	if author == "" || author == "未知作者" {
		author = "未知"
	}
	versionLabel := widget.NewLabel(version)
	versionLabel.Importance = widget.MediumImportance
	authorLabel := widget.NewLabel(author)
	authorLabel.Importance = widget.LowImportance
	topContainer := container.NewBorder(
		nil, nil,
		iconLabel,
		container.NewHBox(updateLabel, configBtn),
		nameLabel,
	)
	infoContainer := container.NewHBox(
		widget.NewLabel("v"),
		versionLabel,
		widget.NewLabel(" • "),
		authorLabel,
	)
	cardContent := container.NewVBox(
		topContainer,
		widget.NewSeparator(),
		descLabel,
		infoContainer,
	)
	cardButton := widget.NewButton("", func() {
		p.showPluginDetails(plugin)
	})
	cardButton.Importance = widget.LowImportance
	clickableCard := container.NewStack(
		cardButton,
		container.NewPadded(cardContent),
	)
	background := canvas.NewRectangle(fyneTheme.OverlayBackgroundColor())
	background.CornerRadius = 5
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = fyneTheme.SeparatorColor()
	border.StrokeWidth = 1
	border.CornerRadius = 5
	card := container.NewStack(
		background,
		border,
		clickableCard,
	)
	return container.NewPadded(card)
}
func (p *PluginPage) setupEventHandlers() {
	p.client.OnConnectionChanged(func(connected bool) {
		if connected {
			p.requestPluginList()
		}
	})
}
func (p *PluginPage) loadPlugins() {
	p.statusLabel.SetText("正在请求服务端数据...")
	p.plugins = nil
	p.refreshPluginContainer()
	if p.client.IsConnected() {
		p.requestPluginList()
	}
}
func (p *PluginPage) requestPluginList() {
	params := map[string]interface{}{}
	p.logger.Debug("Requesting plugin list with empty params")
	callback := &network.RequestCallback{
		Success: func(payload interface{}) {
			var dataAny interface{} = payload
			if pm, ok := payload.(map[string]interface{}); ok {
				if d, has := pm["data"]; has {
					dataAny = d
				}
			}
			parsed := make([]data.Plugin, 0)
			switch v := dataAny.(type) {
			case []interface{}:
				for _, item := range v {
					if m, ok := item.(map[string]interface{}); ok {
						parsed = append(parsed, data.Plugin{
							Name:   toString(m["name"]),
							Module: toString(m["module"]),
							Meta: data.PluginMeta{
								Name:        toString(getMeta(m, "name")),
								Description: toString(getMeta(m, "description")),
								Homepage:    toString(getMeta(m, "homepage")),
								ConfigExist: toBool(getMeta(m, "config_exist")),
								IconAbspath: toString(getMeta(m, "icon_abspath")),
								Author:      toString(getMeta(m, "author")),
								Version:     toString(getMeta(m, "version")),
							},
						})
					}
				}
			case map[string]interface{}:
				for _, item := range v {
					if m, ok := item.(map[string]interface{}); ok {
						parsed = append(parsed, data.Plugin{
							Name:   toString(m["name"]),
							Module: toString(m["module"]),
							Meta: data.PluginMeta{
								Name:        toString(getMeta(m, "name")),
								Description: toString(getMeta(m, "description")),
								Homepage:    toString(getMeta(m, "homepage")),
								ConfigExist: toBool(getMeta(m, "config_exist")),
								IconAbspath: toString(getMeta(m, "icon_abspath")),
								Author:      toString(getMeta(m, "author")),
								Version:     toString(getMeta(m, "version")),
							},
						})
					}
				}
			default:
				p.logger.Warn("Invalid plugin list format: %T", dataAny)
				return
			}
			p.allPlugins = parsed
			if p.isSearching {
				p.filterPlugins(p.searchEntry.Text)
			} else {
				p.plugins = parsed
				p.refreshPluginContainer()
				p.statusLabel.SetText(fmt.Sprintf("共 %d 个插件", len(parsed)))
			}
		},
		Error: func(err error) {
			p.logger.Error("Failed to request plugin list: %v", err)
		},
	}
	if err := p.client.SendRequestWithCallback("get_plugins", params, callback); err != nil {
		p.logger.Error("Failed to send plugin list request: %v", err)
	}
}
func getMeta(m map[string]interface{}, key string) interface{} {
	if meta, ok := m["meta"].(map[string]interface{}); ok {
		return meta[key]
	}
	return nil
}
func toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
func (p *PluginPage) filterPlugins(query string) {
	if strings.TrimSpace(query) == "" {
		p.isSearching = false
		p.plugins = append([]data.Plugin(nil), p.allPlugins...)
		p.refreshPluginContainer()
		p.statusLabel.SetText(fmt.Sprintf("共 %d 个插件", len(p.plugins)))
		return
	}
	p.isSearching = true
	query = strings.ToLower(strings.TrimSpace(query))
	go func() {
		var filteredPlugins []data.Plugin
		for _, plugin := range p.allPlugins {
			if strings.Contains(strings.ToLower(plugin.Meta.Name), query) ||
				strings.Contains(strings.ToLower(plugin.Name), query) ||
				strings.Contains(strings.ToLower(plugin.Meta.Description), query) ||
				strings.Contains(strings.ToLower(plugin.Meta.Author), query) {
				filteredPlugins = append(filteredPlugins, plugin)
			}
		}
		p.plugins = filteredPlugins
		p.refreshPluginContainer()
		p.statusLabel.SetText(fmt.Sprintf("找到 %d 个匹配插件", len(filteredPlugins)))
	}()
}
func (p *PluginPage) showPluginDetails(plugin data.Plugin) {
	details := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("插件名称:"),
			widget.NewLabel(func() string {
				if plugin.Meta.Name != "" {
					return plugin.Meta.Name
				}
				return plugin.Name
			}()),
		),
		container.NewHBox(
			widget.NewLabel("版本:"),
			widget.NewLabel(func() string {
				if plugin.Meta.Version != "" && plugin.Meta.Version != "未知版本" {
					return plugin.Meta.Version
				}
				return "未知"
			}()),
		),
		container.NewHBox(
			widget.NewLabel("作者:"),
			widget.NewLabel(func() string {
				if plugin.Meta.Author != "" && plugin.Meta.Author != "未知作者" {
					return plugin.Meta.Author
				}
				return "未知"
			}()),
		),
		widget.NewSeparator(),
		widget.NewLabel("描述:"),
	)
	descText := plugin.Meta.Description
	if descText == "" {
		descText = "暂无描述"
	}
	descLabel := widget.NewRichTextFromMarkdown(descText)
	descLabel.Wrapping = fyne.TextWrapWord
	descScroll := container.NewScroll(descLabel)
	descScroll.SetMinSize(fyne.NewSize(280, 100))
	details.Add(descScroll)
	details.Add(widget.NewSeparator())
	buttonContainer := container.NewHBox()
	if p.client.IsConnected() {
		configBtn := widget.NewButtonWithIcon("配置", fyneTheme.SettingsIcon(), func() {
			p.showInlinePluginConfig(plugin)
		})
		configBtn.Importance = widget.MediumImportance
		buttonContainer.Add(configBtn)
	}
	if len(buttonContainer.Objects) > 0 {
		details.Add(buttonContainer)
	}
	if plugin.Meta.Version != "" && plugin.Meta.Version != "未知版本" {
		details.Add(widget.NewSeparator())
		updateBtn := widget.NewButtonWithIcon("检查更新", fyneTheme.ViewRefreshIcon(), func() {
			p.checkPluginUpdateAndShow(plugin)
		})
		updateBtn.Importance = widget.MediumImportance
		if plugin.Meta.Homepage != "" && strings.Contains(plugin.Meta.Homepage, "github.com") {
			updateContainer := container.NewHBox(updateBtn)
			details.Add(updateContainer)
		} else {
			details.Add(updateBtn)
		}
	}
	dlg := dialog.NewCustom(
		fmt.Sprintf("插件详情 - %s", func() string {
			if plugin.Meta.Name != "" {
				return plugin.Meta.Name
			}
			return plugin.Name
		}()),
		"关闭",
		details,
		fyne.CurrentApp().Driver().AllWindows()[0],
	)
	dlg.Resize(fyne.NewSize(320, 400))
	dlg.Show()
}
func (p *PluginPage) showInlinePluginConfig(plugin data.Plugin) {
	params := map[string]interface{}{
		"name": plugin.Name,
	}
	callback := &network.RequestCallback{
		Success: func(payload interface{}) {
			p.handleInlinePluginConfigResponse(plugin, payload)
		},
		Error: func(err error) {
			p.logger.Error("Failed to get plugin config: %v", err)
			dialog.ShowError(fmt.Errorf("获取插件配置失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
		},
	}
	if err := p.client.SendRequestWithCallbackTimeout("get_plugin_config", params, callback, 30*time.Second); err != nil {
		p.logger.Error("Failed to send plugin config request: %v", err)
		dialog.ShowError(fmt.Errorf("请求插件配置失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
	}
}
func (p *PluginPage) hideConfigView() {
	p.showingConfig = false
	p.backButton.Hide()
	p.searchEntry.Show()
	p.statusLabel.Show()
	p.mainContent.Objects = nil
	p.mainContent.Add(p.listView)
	p.mainContent.Refresh()
}
func (p *PluginPage) showConfigView() {
	p.showingConfig = true
	p.backButton.Show()
	p.searchEntry.Hide()
	p.statusLabel.Hide()
	p.mainContent.Objects = nil
	p.mainContent.Add(p.configView)
	p.mainContent.Refresh()
	p.configView.Show()  
	p.configView.Refresh()
}
func (p *PluginPage) OnEnter() {
	p.logger.Info("Plugin page entered")
	p.loadPlugins()
}
func (p *PluginPage) updateEmptyState() {
	if len(p.plugins) == 0 {
		p.pluginContainer.Hide()
		p.emptyLabel.Show()
	} else {
		p.emptyLabel.Hide()
		p.pluginContainer.Show()
	}
}
func (p *PluginPage) handleInlinePluginConfigResponse(plugin data.Plugin, payload interface{}) {
	p.logger.Debug("Plugin config response for %s: %+v", plugin.Name, payload)
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		p.logger.Error("Invalid plugin config response format: %T", payload)
		dialog.ShowError(fmt.Errorf("插件配置响应格式无效"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	if errorMsg, hasError := payloadMap["error"]; hasError && errorMsg != nil && errorMsg != "<nil>" && toString(errorMsg) != "" {
		errorStr := toString(errorMsg)
		p.logger.Debug("Plugin %s config error: %s", plugin.Name, errorStr)
		var userMessage string
		switch errorStr {
		case "Plugin not found":
			userMessage = "插件未找到"
		case "Plugin config not found":
			userMessage = "此插件没有可配置的选项"
		default:
			userMessage = fmt.Sprintf("获取插件配置失败: %s", errorStr)
		}
		dialog.ShowInformation("插件配置", userMessage, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	var schema map[string]interface{}
	var data string
	var schemaExists bool
	if dataField, hasDataField := payloadMap["data"].(map[string]interface{}); hasDataField {
		p.logger.Debug("Plugin %s found nested data field", plugin.Name)
		schema, schemaExists = dataField["schema"].(map[string]interface{})
		if dataStr, ok := dataField["data"].(string); ok {
			data = dataStr
		}
	} else {
		schema, schemaExists = payloadMap["schema"].(map[string]interface{})
		if dataStr, ok := payloadMap["data"].(string); ok {
			data = dataStr
		}
	}
	if !schemaExists || schema == nil || len(schema) == 0 {
		p.logger.Debug("Plugin %s has no schema or empty schema", plugin.Name)
		dialog.ShowInformation("插件配置", "此插件没有可配置的选项", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	properties, hasProperties := schema["properties"].(map[string]interface{})
	if !hasProperties || len(properties) == 0 {
		p.logger.Debug("Plugin %s has no properties in schema", plugin.Name)
		dialog.ShowInformation("插件配置", "此插件没有可配置的选项", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	moduleName := plugin.Name
	if m, ok := payloadMap["module"]; ok {
		moduleName = toString(m)
	} else if m2, ok := payloadMap["module_name"]; ok {
		moduleName = toString(m2)
	}
	p.logger.Debug("Creating inline config view for plugin %s with %d properties", plugin.Name, len(properties))
	p.createInlineConfigView(plugin, schema, data, moduleName)
}
func (p *PluginPage) createInlineConfigView(plugin data.Plugin, schema map[string]interface{}, configData string, moduleName string) {
	p.currentPlugin = plugin
	var config map[string]interface{}
	if configData != "" {
		if err := json.Unmarshal([]byte(configData), &config); err != nil {
			p.logger.Error("Failed to parse config data: %v", err)
			config = make(map[string]interface{})
		}
	} else {
		config = make(map[string]interface{})
	}
	p.configView.Objects = nil
	p.configView.Refresh()  
	titleLabel := widget.NewLabelWithStyle(fmt.Sprintf("%s 配置", plugin.Name), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	p.configView.Add(titleLabel)
	p.configView.Add(widget.NewSeparator())
	configForm, valueGetters, validators, errorLabels := p.createConfigForm(schema, config)
	p.logger.Debug("Config form created with %d objects", len(configForm.Objects))
	if len(configForm.Objects) == 0 {
		p.logger.Debug("Config form is empty, adding debug info")
		configForm.Add(widget.NewLabel("表单为空 - 调试信息"))
	}
	configScroll := container.NewScroll(configForm)
	configScroll.SetMinSize(fyne.NewSize(0, 300))  
	p.configView.Add(configScroll)
	p.configView.Add(widget.NewSeparator())
	saveBtn := widget.NewButtonWithIcon("保存", fyneTheme.DocumentSaveIcon(), func() {
		p.saveInlinePluginConfig(plugin, moduleName, valueGetters, validators, errorLabels)
	})
	saveBtn.Importance = widget.HighImportance
	cancelBtn := widget.NewButtonWithIcon("取消", fyneTheme.CancelIcon(), func() {
		p.hideConfigView()
	})
	buttonContainer := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		saveBtn,
	)
	p.configView.Add(buttonContainer)
	p.logger.Debug("Refreshing config view before showing, objects count: %d", len(p.configView.Objects))
	p.configView.Refresh()
	p.showConfigView()
	go func() {
		time.Sleep(50 * time.Millisecond)  
		if p.configView.Visible() {
			p.logger.Debug("Config view is visible, performing final refresh")
			p.configView.Refresh()
			configScroll.Refresh()
			configForm.Refresh()
		} else {
			p.logger.Warn("Config view is not visible after showConfigView()")
		}
	}()
}
func (p *PluginPage) saveInlinePluginConfig(
	plugin data.Plugin,
	moduleName string,
	getters map[string]func() interface{},
	validators map[string]func(interface{}) error,
	errorLabels map[string]*widget.Label,
) {
	data := make(map[string]interface{})
	for k, get := range getters {
		if get != nil {
			data[k] = get()
		}
	}
	hasError := false
	for field, getter := range getters {
		val := getter()
		if validate, ok := validators[field]; ok && validate != nil {
			if err := validate(val); err != nil {
				hasError = true
				if lbl, ok := errorLabels[field]; ok && lbl != nil {
					lbl.SetText(err.Error())
					lbl.Show()
				}
			} else if lbl, ok := errorLabels[field]; ok && lbl != nil {
				lbl.Hide()
			}
		}
	}
	if hasError {
		dialog.ShowError(errors.New("请根据提示修正配置后重试"), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	_, err := json.Marshal(data)
	if err != nil {
		dialog.ShowError(fmt.Errorf("配置序列化失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	envParams := map[string]interface{}{
		"module_name": moduleName,
		"data":        data,
	}
	if err := p.client.SendRequestWithCallback("save_env", envParams, &network.RequestCallback{
		Success: func(_ interface{}) {
			dialog.ShowInformation("保存成功", fmt.Sprintf("插件 '%s' 的配置已保存", plugin.Name), fyne.CurrentApp().Driver().AllWindows()[0])
			p.hideConfigView()  
		},
		Error: func(e3 error) {
			dialog.ShowError(fmt.Errorf("保存失败: %v", e3), fyne.CurrentApp().Driver().AllWindows()[0])
		},
	}); err != nil {
		dialog.ShowError(fmt.Errorf("请求发送失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
	}
}
func (p *PluginPage) createConfigForm(schema map[string]interface{}, config map[string]interface{}) (*fyne.Container, map[string]func() interface{}, map[string]func(interface{}) error, map[string]*widget.Label) {
	form := container.NewVBox()
	getters := make(map[string]func() interface{})
	validators := make(map[string]func(interface{}) error)
	errorLabels := make(map[string]*widget.Label)
	if schema == nil {
		form.Add(widget.NewLabel("此插件没有可配置的选项"))
		return form, getters, validators, errorLabels
	}
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		form.Add(widget.NewLabel("无法解析配置架构"))
		return form, getters, validators, errorLabels
	}
	for propName, propSchema := range properties {
		p.logger.Debug("Processing property: %s", propName)
		propSchemaMap, ok := propSchema.(map[string]interface{})
		if !ok {
			p.logger.Debug("Skipping property %s: invalid schema type", propName)
			continue
		}
		propType, _ := propSchemaMap["type"].(string)
		title, _ := propSchemaMap["title"].(string)
		description, _ := propSchemaMap["description"].(string)
		p.logger.Debug("Property %s: type=%s, title=%s, desc=%s", propName, propType, title, description)
		required := false
		if reqArr, ok := schema["required"].([]interface{}); ok {
			for _, r := range reqArr {
				if toString(r) == propName {
					required = true
					break
				}
			}
		}
		if title == "" {
			title = propName
		}
		currentValue := config[propName]
		var inputWidget fyne.CanvasObject
		switch propType {
		case "boolean":
			check := widget.NewCheck(title, nil)
			if val, ok := currentValue.(bool); ok {
				check.SetChecked(val)
			}
			getters[propName] = func() interface{} { return check.Checked }
			inputWidget = check
		case "number", "integer":
			entry := widget.NewEntry()
			entry.SetPlaceHolder("输入数字")
			if currentValue != nil {
				entry.SetText(fmt.Sprintf("%v", currentValue))
			}
			inputWidget = container.NewVBox(
				widget.NewLabel(title),
				entry,
			)
			getters[propName] = func() interface{} {
				txt := strings.TrimSpace(entry.Text)
				if txt == "" {
					return 0
				}
				if v, err := strconv.ParseFloat(txt, 64); err == nil {
					return v
				}
				return txt
			}
			minV, hasMin := propSchemaMap["minimum"].(float64)
			maxV, hasMax := propSchemaMap["maximum"].(float64)
			validators[propName] = func(v interface{}) error {
				f, ok := v.(float64)
				if !ok {
					return errors.New("请输入数字")
				}
				if hasMin && f < minV {
					return fmt.Errorf("值不能小于 %v", minV)
				}
				if hasMax && f > maxV {
					return fmt.Errorf("值不能大于 %v", maxV)
				}
				return nil
			}
		case "array":
			arrayWidget, arrayGetter, arrayValidator := p.createArrayInput(propName, title, propSchemaMap, currentValue)
			inputWidget = arrayWidget
			getters[propName] = arrayGetter
			if arrayValidator != nil {
				validators[propName] = arrayValidator
			}
		default:  
			if enumVals, ok := propSchemaMap["enum"].([]interface{}); ok && len(enumVals) > 0 {
				options := make([]string, 0, len(enumVals))
				for _, e := range enumVals {
					options = append(options, toString(e))
				}
				sel := widget.NewSelect(options, nil)
				if str, ok := currentValue.(string); ok && str != "" {
					sel.SetSelected(str)
				}
				inputWidget = container.NewVBox(widget.NewLabel(title), sel)
				getters[propName] = func() interface{} { return sel.Selected }
				validators[propName] = func(v interface{}) error {
					s := toString(v)
					for _, opt := range options {
						if opt == s {
							return nil
						}
					}
					return errors.New("请选择有效选项")
				}
			} else {
				entry := widget.NewEntry()
				if val, ok := currentValue.(string); ok {
					entry.SetText(val)
				}
				inputWidget = container.NewVBox(widget.NewLabel(title), entry)
				getters[propName] = func() interface{} { return entry.Text }
			}
		}
		errLabel := widget.NewLabel("")
		errLabel.Importance = widget.DangerImportance
		errLabel.Hide()
		errorLabels[propName] = errLabel
		form.Add(inputWidget)
		if description != "" {
			descLabel := widget.NewLabel(description)
			descLabel.Importance = widget.LowImportance
			descLabel.Wrapping = fyne.TextWrapWord
			form.Add(descLabel)
		}
		if required {
			form.Add(widget.NewLabel("(必填)"))
			if _, exists := validators[propName]; !exists {
				validators[propName] = func(v interface{}) error {
					if v == nil {
						return errors.New("该字段为必填项")
					}
					s := toString(v)
					if strings.TrimSpace(s) == "" {
						return errors.New("该字段为必填项")
					}
					return nil
				}
			}
		}
		form.Add(errLabel)
		form.Add(widget.NewSeparator())
	}
	return form, getters, validators, errorLabels
}
func (p *PluginPage) createArrayInput(_, title string, propSchema map[string]interface{}, currentValue interface{}) (fyne.CanvasObject, func() interface{}, func(interface{}) error) {
	var currentArray []string
	if currentValue != nil {
		if arr, ok := currentValue.([]interface{}); ok {
			for _, item := range arr {
				currentArray = append(currentArray, toString(item))
			}
		}
	}
	inputContainer := container.NewVBox()
	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	inputContainer.Add(titleLabel)
	var entries []*widget.Entry
	var entriesContainer = container.NewVBox()
	inputContainer.Add(entriesContainer)
	for _, value := range currentArray {
		entry := widget.NewEntry()
		entry.SetText(value)
		entry.SetPlaceHolder("输入项目...")
		deleteBtn := widget.NewButtonWithIcon("", fyneTheme.DeleteIcon(), nil)
		deleteBtn.Importance = widget.DangerImportance
		itemContainer := container.NewBorder(nil, nil, nil, deleteBtn, entry)
		entries = append(entries, entry)
		entriesContainer.Add(itemContainer)
		func(entryToDelete *widget.Entry, containerToRemove *fyne.Container) {
			deleteBtn.OnTapped = func() {
				for i, e := range entries {
					if e == entryToDelete {
						entries = append(entries[:i], entries[i+1:]...)
						break
					}
				}
				entriesContainer.Remove(containerToRemove)
				entriesContainer.Refresh()
			}
		}(entry, itemContainer)
	}
	addBtn := widget.NewButtonWithIcon("添加项目", fyneTheme.ContentAddIcon(), func() {
		entry := widget.NewEntry()
		entry.SetPlaceHolder("输入项目...")
		deleteBtn := widget.NewButtonWithIcon("", fyneTheme.DeleteIcon(), nil)
		deleteBtn.Importance = widget.DangerImportance
		itemContainer := container.NewBorder(nil, nil, nil, deleteBtn, entry)
		entries = append(entries, entry)
		entriesContainer.Add(itemContainer)
		entriesContainer.Refresh()
		deleteBtn.OnTapped = func() {
			for i, e := range entries {
				if e == entry {
					entries = append(entries[:i], entries[i+1:]...)
					break
				}
			}
			entriesContainer.Remove(itemContainer)
			entriesContainer.Refresh()
		}
	})
	addBtn.Importance = widget.MediumImportance
	inputContainer.Add(addBtn)
	getter := func() interface{} {
		var result []string
		for _, entry := range entries {
			text := strings.TrimSpace(entry.Text)
			if text != "" {
				result = append(result, text)
			}
		}
		if uniqueItems, ok := propSchema["uniqueItems"].(bool); ok && uniqueItems {
			seen := make(map[string]bool)
			var unique []string
			for _, item := range result {
				if !seen[item] {
					seen[item] = true
					unique = append(unique, item)
				}
			}
			result = unique
		}
		return result
	}
	validator := func(v interface{}) error {
		arr, ok := v.([]string)
		if !ok {
			return errors.New("数组类型错误")
		}
		if uniqueItems, ok := propSchema["uniqueItems"].(bool); ok && uniqueItems {
			seen := make(map[string]bool)
			for _, item := range arr {
				if seen[item] {
					return fmt.Errorf("数组中包含重复项: %s", item)
				}
				seen[item] = true
			}
		}
		if minItems, ok := propSchema["minItems"].(float64); ok && len(arr) < int(minItems) {
			return fmt.Errorf("至少需要 %d 个项目", int(minItems))
		}
		if maxItems, ok := propSchema["maxItems"].(float64); ok && len(arr) > int(maxItems) {
			return fmt.Errorf("最多允许 %d 个项目", int(maxItems))
		}
		return nil
	}
	return inputContainer, getter, validator
}
func (p *PluginPage) checkPluginUpdateAndShow(plugin data.Plugin) {
	homepage := plugin.Meta.Homepage
	if homepage == "" || !strings.Contains(homepage, "github.com") {
		dialog.ShowInformation("检查更新", "该插件没有提供GitHub主页信息，无法检查更新", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	repoPattern := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)`)
	matches := repoPattern.FindStringSubmatch(homepage)
	if len(matches) < 3 {
		dialog.ShowInformation("检查更新", "无法解析插件的GitHub仓库信息", fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	owner := matches[1]
	repo := matches[2]
	currentVersion := plugin.Meta.Version
	progressDialog := dialog.NewInformation("检查更新", "正在检查插件更新...", fyne.CurrentApp().Driver().AllWindows()[0])
	progressDialog.Show()
	go func() {
		updateChecker := utils.NewUpdateCheckerWithLogger(p.logger)
		release, hasUpdate, err := updateChecker.CheckForUpdates(owner, repo, currentVersion)
		progressDialog.Hide()
		if err != nil {
			dialog.ShowError(fmt.Errorf("检查插件更新失败: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}
		if hasUpdate && release != nil {
			displayName := plugin.Meta.Name
			if displayName == "" {
				displayName = plugin.Name
			}
			p.showPluginUpdateDialog(plugin, displayName, release.TagName, release.Body, homepage)
		} else {
			displayName := plugin.Meta.Name
			if displayName == "" {
				displayName = plugin.Name
			}
			dialog.ShowInformation("检查更新", fmt.Sprintf("插件 %s 已是最新版本", displayName), fyne.CurrentApp().Driver().AllWindows()[0])
		}
	}()
}
func (p *PluginPage) showPluginUpdateDialog(plugin data.Plugin, pluginDisplayName, latestVersion, updateNotes, downloadURL string) {
	currentVersion := plugin.Meta.Version
	updateContent := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("插件名称: %s", pluginDisplayName)),
		widget.NewLabel(fmt.Sprintf("当前版本: v%s", currentVersion)),
		widget.NewLabel(fmt.Sprintf("最新版本: %s", latestVersion)),
		widget.NewSeparator(),
		widget.NewLabel("更新说明:"),
	)
	if updateNotes != "" {
		notesLabel := widget.NewRichTextFromMarkdown(updateNotes)
		notesLabel.Wrapping = fyne.TextWrapWord
		scrollContainer := container.NewScroll(notesLabel)
		scrollContainer.SetMinSize(fyne.NewSize(300, 120))
		updateContent.Add(scrollContainer)
	} else {
		updateContent.Add(widget.NewLabel("暂无更新说明"))
	}
	updateDialog := dialog.NewCustom("插件更新", "关闭", updateContent, fyne.CurrentApp().Driver().AllWindows()[0])
	updateBtn := widget.NewButtonWithIcon("立即更新", fyneTheme.DownloadIcon(), func() {
		updateDialog.Hide()  
		p.performPluginUpdate(plugin, pluginDisplayName, latestVersion)
	})
	updateBtn.Importance = widget.HighImportance
	laterBtn := widget.NewButton("稍后提醒", func() {
		updateDialog.Hide()  
	})
	githubBtn := widget.NewButtonWithIcon("GitHub页面", fyneTheme.ComputerIcon(), func() {
		if downloadURL != "" {
			dialog.ShowInformation("前往GitHub", fmt.Sprintf("请在浏览器中访问以下链接:\n%s", downloadURL), fyne.CurrentApp().Driver().AllWindows()[0])
		}
	})
	buttonContainer := container.NewGridWithColumns(3, laterBtn, githubBtn, updateBtn)
	updateContent.Add(widget.NewSeparator())
	updateContent.Add(buttonContainer)
	updateDialog.Resize(fyne.NewSize(400, 350))
	updateDialog.Show()
}
func (p *PluginPage) performPluginUpdate(plugin data.Plugin, pluginDisplayName, latestVersion string) {
	confirmTitle := "确认更新插件"
	confirmMessage := fmt.Sprintf("将更新插件 %s 到 v%s，请确认执行操作.\n更新完成后将弹窗提醒.\n请不要切换页面", pluginDisplayName, latestVersion)
	dialog.ShowConfirm(confirmTitle, confirmMessage, func(confirmed bool) {
		if !confirmed {
			return
		}
		p.startPluginUpdate(plugin, pluginDisplayName, latestVersion)
	}, fyne.CurrentApp().Driver().AllWindows()[0])
}
func (p *PluginPage) startPluginUpdate(plugin data.Plugin, pluginDisplayName, latestVersion string) {
	progressDialog := dialog.NewInformation("更新插件", fmt.Sprintf("正在更新插件 %s...", pluginDisplayName), fyne.CurrentApp().Driver().AllWindows()[0])
	progressDialog.Show()
	params := map[string]interface{}{
		"plugin_name": plugin.Name,  
	}
	callback := &network.RequestCallback{
		Success: func(payload interface{}) {
			progressDialog.Hide()
			dialog.ShowInformation("更新成功",
				fmt.Sprintf("插件 %s 已成功更新到 %s\n重启NoneBot以应用更新", pluginDisplayName, latestVersion),
				fyne.CurrentApp().Driver().AllWindows()[0])
		},
		Error: func(err error) {
			progressDialog.Hide()
			errorMessage := fmt.Sprintf("插件 %s 更新失败:\n%v", pluginDisplayName, err)
			dialog.ShowError(errors.New(errorMessage), fyne.CurrentApp().Driver().AllWindows()[0])
		},
	}
	if err := p.client.SendRequestWithCallbackTimeout("update_plugin", params, callback, 10*time.Minute); err != nil {
		progressDialog.Hide()
		dialog.ShowError(fmt.Errorf("无法发送插件更新请求: %v", err), fyne.CurrentApp().Driver().AllWindows()[0])
	}
}
