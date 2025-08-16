package roster
import (
	"encoding/json"
	"fmt"
	"strings"
)
type MatcherRule struct {
	AlconnaCommands []string   `json:"alconna_commands,omitempty"`
	Commands        [][]string `json:"commands,omitempty"`
	RegexPatterns   []string   `json:"regex_patterns,omitempty"`
	Keywords        []string   `json:"keywords,omitempty"`
	StartsWith      []string   `json:"startswith,omitempty"`
	EndsWith        []string   `json:"endswith,omitempty"`
	FullMatch       []string   `json:"fullmatch,omitempty"`
	EventTypes      []string   `json:"event_types,omitempty"`
	ToMe            bool       `json:"to_me,omitempty"`
}
type Permission struct {
	WhiteList PermissionList `json:"white_list"`
	BanList   PermissionList `json:"ban_list"`
}
type PermissionList struct {
	User  []string `json:"user"`
	Group []string `json:"group"`
}
type MatcherConfig struct {
	Rule       MatcherRule `json:"rule"`
	Permission Permission  `json:"permission"`
	IsOn       bool        `json:"is_on"`
}
type PluginConfig struct {
	Matchers []MatcherConfig `json:"matchers"`
}
type BotConfig struct {
	Plugins map[string]PluginConfig `json:"plugins"`
}
type FullConfig struct {
	Bots map[string]BotConfig `json:"bots"`
}
func (mr *MatcherRule) GetDisplayName() string {
	if len(mr.AlconnaCommands) > 0 {
		return fmt.Sprintf("Alconna: %s", strings.Join(mr.AlconnaCommands, ", "))
	}
	if len(mr.Commands) > 0 {
		cmdStrs := make([]string, len(mr.Commands))
		for i, cmd := range mr.Commands {
			cmdStrs[i] = strings.Join(cmd, "/")
		}
		return fmt.Sprintf("命令: %s", strings.Join(cmdStrs, ", "))
	}
	if len(mr.RegexPatterns) > 0 {
		return fmt.Sprintf("正则: %s", strings.Join(mr.RegexPatterns, ", "))
	}
	if len(mr.Keywords) > 0 {
		return fmt.Sprintf("关键词: %s", strings.Join(mr.Keywords, ", "))
	}
	if len(mr.StartsWith) > 0 {
		return fmt.Sprintf("开头: %s", strings.Join(mr.StartsWith, ", "))
	}
	if len(mr.EndsWith) > 0 {
		return fmt.Sprintf("结尾: %s", strings.Join(mr.EndsWith, ", "))
	}
	if len(mr.FullMatch) > 0 {
		return fmt.Sprintf("完全匹配: %s", strings.Join(mr.FullMatch, ", "))
	}
	if len(mr.EventTypes) > 0 {
		return fmt.Sprintf("事件: %s", strings.Join(mr.EventTypes, ", "))
	}
	return "通用规则"
}
func (mr *MatcherRule) GetDetails() map[string]string {
	details := make(map[string]string)
	if len(mr.AlconnaCommands) > 0 {
		details["Alconna 命令"] = strings.Join(mr.AlconnaCommands, ", ")
	}
	if len(mr.Commands) > 0 {
		cmdStrs := make([]string, len(mr.Commands))
		for i, cmd := range mr.Commands {
			cmdStrs[i] = strings.Join(cmd, "/")
		}
		details["命令"] = strings.Join(cmdStrs, ", ")
	}
	if len(mr.RegexPatterns) > 0 {
		details["正则表达式"] = strings.Join(mr.RegexPatterns, "\n")
	}
	if len(mr.Keywords) > 0 {
		details["关键词"] = strings.Join(mr.Keywords, ", ")
	}
	if len(mr.StartsWith) > 0 {
		details["开头匹配"] = strings.Join(mr.StartsWith, ", ")
	}
	if len(mr.EndsWith) > 0 {
		details["结尾匹配"] = strings.Join(mr.EndsWith, ", ")
	}
	if len(mr.FullMatch) > 0 {
		details["完全匹配"] = strings.Join(mr.FullMatch, ", ")
	}
	if len(mr.EventTypes) > 0 {
		details["事件类型"] = strings.Join(mr.EventTypes, ", ")
	}
	if mr.ToMe {
		details["触发方式"] = "需要@机器人"
	}
	return details
}
func (p *Permission) CheckPermission(userID, groupID string) bool {
	for _, bannedUser := range p.BanList.User {
		if bannedUser == userID {
			return false
		}
	}
	if groupID != "" {
		for _, bannedGroup := range p.BanList.Group {
			if bannedGroup == groupID {
				return false
			}
		}
	}
	if len(p.WhiteList.User) == 0 && len(p.WhiteList.Group) == 0 {
		return true
	}
	for _, allowedUser := range p.WhiteList.User {
		if allowedUser == userID {
			return true
		}
	}
	if groupID != "" {
		for _, allowedGroup := range p.WhiteList.Group {
			if allowedGroup == groupID {
				return true
			}
		}
	}
	return false
}
func ParseConfig(data map[string]interface{}) (*FullConfig, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}
	var config FullConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &config, nil
}
func (fc *FullConfig) ToMap() (map[string]interface{}, error) {
	jsonData, err := json.Marshal(fc)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("unmarshal to map: %w", err)
	}
	return result, nil
}
func (fc *FullConfig) Clone() (*FullConfig, error) {
	data, err := fc.ToMap()
	if err != nil {
		return nil, err
	}
	return ParseConfig(data)
}
func NewMatcherRuleFromMap(data map[string]interface{}) (MatcherRule, error) {
	var rule MatcherRule
	jsonData, err := json.Marshal(data)
	if err != nil {
		return rule, fmt.Errorf("failed to marshal rule data: %w", err)
	}
	if err := json.Unmarshal(jsonData, &rule); err != nil {
		return rule, fmt.Errorf("failed to unmarshal rule data: %w", err)
	}
	return rule, nil
}
