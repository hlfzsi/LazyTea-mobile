package data
import (
	"encoding/json"
	"fmt"
)
type PermissionListDivide struct {
	User  []string `json:"user"`
	Group []string `json:"group"`
}
type MatcherPermission struct {
	WhiteList PermissionListDivide `json:"white_list"`
	BanList   PermissionListDivide `json:"ban_list"`
}
type MatcherRuleModel struct {
	Rule       map[string]interface{} `json:"rule"`
	Permission MatcherPermission      `json:"permission"`
	IsOn       bool                   `json:"is_on"`
}
type PluginModel struct {
	Matchers []MatcherRuleModel `json:"matchers"`
}
type BotModel struct {
	Plugins map[string]PluginModel `json:"plugins"`
}
type FullConfigModel struct {
	Bots map[string]BotModel `json:"bots"`
}
var ReadableRoster FullConfigModel
func init() {
	ReadableRoster.Bots = make(map[string]BotModel)
}
func UpdateConfig(newConfig FullConfigModel) {
	ReadableRoster = newConfig
}
func GetConfig() FullConfigModel {
	return ReadableRoster
}
func ToJson() (string, error) {
	bytes, err := json.Marshal(ReadableRoster)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
func FromJson(jsonStr string) error {
	var newConfig FullConfigModel
	err := json.Unmarshal([]byte(jsonStr), &newConfig)
	if err != nil {
		return err
	}
	UpdateConfig(newConfig)
	return nil
}
func ParseConfigFromMap(data map[string]interface{}) (*FullConfigModel, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}
	var config FullConfigModel
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &config, nil
}
func (fcm *FullConfigModel) ToMap() (map[string]interface{}, error) {
	jsonData, err := json.Marshal(fcm)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("unmarshal to map: %w", err)
	}
	return result, nil
}
func (p *MatcherPermission) CheckPermission(userID, groupID string) bool {
	if stringInSlice(userID, p.BanList.User) || stringInSlice(groupID, p.BanList.Group) {
		return false
	}
	if len(p.WhiteList.User) == 0 && len(p.WhiteList.Group) == 0 {
		return true  
	}
	if stringInSlice(userID, p.WhiteList.User) || stringInSlice(groupID, p.WhiteList.Group) {
		return true
	}
	return false
}
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func (fcm *FullConfigModel) CheckPermission(bot, plugin, matcherKey, userID string, groupID *string) bool {
	botConfig, exists := fcm.Bots[bot]
	if !exists {
		return true  
	}
	pluginConfig, exists := botConfig.Plugins[plugin]
	if !exists {
		return true  
	}
	for _, matcher := range pluginConfig.Matchers {
		if GetRuleDisplayName(matcher.Rule) == matcherKey {
			return evaluateMatcherConfig(matcher, userID, groupID)
		}
	}
	return true  
}
func evaluateMatcherConfig(matcherConfig MatcherRuleModel, userID string, groupID *string) bool {
	isOn := matcherConfig.IsOn
	permission := matcherConfig.Permission
	whiteList := permission.WhiteList
	banList := permission.BanList
	inWhiteUser := contains(whiteList.User, userID)
	inWhiteGroup := false
	if groupID != nil {
		inWhiteGroup = contains(whiteList.Group, *groupID)
	}
	inBanUser := contains(banList.User, userID)
	inBanGroup := false
	if groupID != nil {
		inBanGroup = contains(banList.Group, *groupID)
	}
	if inBanUser || inBanGroup {
		return false
	}
	if isOn {
		if len(whiteList.User) > 0 || len(whiteList.Group) > 0 {
			return inWhiteUser || inWhiteGroup
		}
		return true
	} else {
		return inWhiteUser || inWhiteGroup
	}
}
func GetRuleDisplayName(ruleData map[string]interface{}) string {
	if ruleData == nil {
		return "空规则"
	}
	var parts []string
	if alconnaCommands, ok := ruleData["alconna_commands"].([]interface{}); ok && len(alconnaCommands) > 0 {
		var cmds []string
		for _, cmd := range alconnaCommands {
			if cmdStr, ok := cmd.(string); ok {
				cmds = append(cmds, cmdStr)
			}
		}
		if len(cmds) > 0 {
			parts = append(parts, fmt.Sprintf("Alconna: %s", joinStrings(cmds, ", ")))
		}
	}
	if commands, ok := ruleData["commands"].([]interface{}); ok && len(commands) > 0 {
		var cmds []string
		for _, cmd := range commands {
			if cmdSlice, ok := cmd.([]interface{}); ok {
				var cmdParts []string
				for _, part := range cmdSlice {
					if partStr, ok := part.(string); ok {
						cmdParts = append(cmdParts, partStr)
					}
				}
				cmds = append(cmds, joinStrings(cmdParts, "/"))
			}
		}
		if len(cmds) > 0 {
			parts = append(parts, fmt.Sprintf("命令: %s", joinStrings(cmds, ", ")))
		}
	}
	if startswith, ok := ruleData["startswith"].([]interface{}); ok && len(startswith) > 0 {
		var items []string
		for _, item := range startswith {
			if itemStr, ok := item.(string); ok {
				items = append(items, itemStr)
			}
		}
		if len(items) > 0 {
			parts = append(parts, fmt.Sprintf("开头: %s", joinStrings(items, ", ")))
		}
	}
	if endswith, ok := ruleData["endswith"].([]interface{}); ok && len(endswith) > 0 {
		var items []string
		for _, item := range endswith {
			if itemStr, ok := item.(string); ok {
				items = append(items, itemStr)
			}
		}
		if len(items) > 0 {
			parts = append(parts, fmt.Sprintf("结尾: %s", joinStrings(items, ", ")))
		}
	}
	if fullmatch, ok := ruleData["fullmatch"].([]interface{}); ok && len(fullmatch) > 0 {
		var items []string
		for _, item := range fullmatch {
			if itemStr, ok := item.(string); ok {
				items = append(items, itemStr)
			}
		}
		if len(items) > 0 {
			parts = append(parts, fmt.Sprintf("全匹配: %s", joinStrings(items, ", ")))
		}
	}
	if keywords, ok := ruleData["keywords"].([]interface{}); ok && len(keywords) > 0 {
		var items []string
		for _, item := range keywords {
			if itemStr, ok := item.(string); ok {
				items = append(items, itemStr)
			}
		}
		if len(items) > 0 {
			parts = append(parts, fmt.Sprintf("关键词: %s", joinStrings(items, ", ")))
		}
	}
	if regexPatterns, ok := ruleData["regex_patterns"].([]interface{}); ok && len(regexPatterns) > 0 {
		var patterns []string
		for _, pattern := range regexPatterns {
			if patternStr, ok := pattern.(string); ok {
				if len(patternStr) > 30 {
					patternStr = patternStr[:30] + "..."
				}
				patterns = append(patterns, patternStr)
			}
		}
		if len(patterns) > 0 {
			parts = append(parts, fmt.Sprintf("正则: %s", joinStrings(patterns, ", ")))
		}
	}
	if toMe, ok := ruleData["to_me"].(bool); ok && toMe {
		parts = append(parts, "@机器人")
	}
	if eventTypes, ok := ruleData["event_types"].([]interface{}); ok && len(eventTypes) > 0 {
		var types []string
		for _, eventType := range eventTypes {
			if typeStr, ok := eventType.(string); ok {
				types = append(types, typeStr)
			}
		}
		if len(types) > 0 {
			parts = append(parts, fmt.Sprintf("事件: %s", joinStrings(types, ", ")))
		}
	}
	if len(parts) == 0 {
		return "通用规则"
	}
	return joinStrings(parts, " | ")
}
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
