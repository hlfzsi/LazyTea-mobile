package roster
import (
	"fmt"
	"lazytea-mobile/internal/data"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)
type NodeType int
const (
	NodeTypeBot NodeType = iota
	NodeTypePlugin
	NodeTypeMatcher
)
type TreeNode struct {
	ID          string
	DisplayName string
	Type        NodeType
	BotID       string
	PluginName  string
	MatcherIdx  int
	Children    []*TreeNode
}
type ConfigTree struct {
	widget.BaseWidget
	tree  *widget.Tree
	roots []*TreeNode
}
func NewConfigTree(config *data.FullConfigModel, onSelect func(node *TreeNode)) *ConfigTree {
	ct := &ConfigTree{}
	ct.ExtendBaseWidget(ct)
	ct.roots = BuildTreeFromConfig(config)
	ct.tree = widget.NewTree(
		func(id string) (children []string) {
			if id == "" {
				for _, root := range ct.roots {
					children = append(children, root.ID)
				}
				return
			}
			node := FindNode(ct.roots, id)
			if node != nil {
				for _, child := range node.Children {
					children = append(children, child.ID)
				}
			}
			return
		},
		func(id string) bool {
			if id == "" {
				return len(ct.roots) > 0
			}
			node := FindNode(ct.roots, id)
			return node != nil && len(node.Children) > 0
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return widget.NewLabel("Parent")
			}
			return widget.NewLabel("Leaf")
		},
		func(id string, branch bool, o fyne.CanvasObject) {
			if id == "" {
				o.(*widget.Label).SetText("Root")
				return
			}
			node := FindNode(ct.roots, id)
			if node != nil {
				o.(*widget.Label).SetText(node.DisplayName)
			} else {
				o.(*widget.Label).SetText("Unknown")
			}
		},
	)
	ct.tree.OnSelected = func(uid string) {
		node := FindNode(ct.roots, uid)
		if onSelect != nil {
			onSelect(node)
		}
	}
	for _, root := range ct.roots {
		ct.tree.OpenBranch(root.ID)
	}
	return ct
}
func NewConfigTreeForBot(config *data.FullConfigModel, targetBotID string, onSelect func(node *TreeNode)) *ConfigTree {
	ct := &ConfigTree{}
	ct.ExtendBaseWidget(ct)
	ct.roots = BuildTreeFromConfigForBot(config, targetBotID)
	ct.tree = widget.NewTree(
		func(id string) (children []string) {
			if id == "" {
				for _, root := range ct.roots {
					children = append(children, root.ID)
				}
				return
			}
			node := FindNode(ct.roots, id)
			if node != nil {
				for _, child := range node.Children {
					children = append(children, child.ID)
				}
			}
			return
		},
		func(id string) bool {
			if id == "" {
				return len(ct.roots) > 0
			}
			node := FindNode(ct.roots, id)
			return node != nil && len(node.Children) > 0
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return widget.NewLabel("Parent")
			}
			return widget.NewLabel("Leaf")
		},
		func(id string, branch bool, o fyne.CanvasObject) {
			if id == "" {
				o.(*widget.Label).SetText("Root")
				return
			}
			node := FindNode(ct.roots, id)
			if node != nil {
				o.(*widget.Label).SetText(node.DisplayName)
			} else {
				o.(*widget.Label).SetText("Unknown")
			}
		},
	)
	ct.tree.OnSelected = func(uid string) {
		node := FindNode(ct.roots, uid)
		if onSelect != nil {
			onSelect(node)
		}
	}
	for _, root := range ct.roots {
		ct.tree.OpenBranch(root.ID)
	}
	return ct
}
func (ct *ConfigTree) UpdateConfig(config *data.FullConfigModel) {
	ct.roots = BuildTreeFromConfig(config)
	ct.tree.Refresh()
	for _, root := range ct.roots {
		ct.tree.OpenBranch(root.ID)
	}
}
func (ct *ConfigTree) UpdateConfigForBot(config *data.FullConfigModel, targetBotID string) {
	ct.roots = BuildTreeFromConfigForBot(config, targetBotID)
	ct.tree.Refresh()
	for _, root := range ct.roots {
		ct.tree.OpenBranch(root.ID)
	}
}
func (ct *ConfigTree) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ct.tree)
}
func BuildTreeFromConfig(config *data.FullConfigModel) []*TreeNode {
	var roots []*TreeNode
	if config == nil {
		return roots
	}
	for botID, bot := range config.Bots {
		botNode := &TreeNode{
			ID:          "bot_" + botID,
			DisplayName: botID,
			Type:        NodeTypeBot,
			BotID:       botID,
		}
		for pluginName, plugin := range bot.Plugins {
			pluginNode := &TreeNode{
				ID:          botNode.ID + "_plugin_" + pluginName,
				DisplayName: pluginName,
				Type:        NodeTypePlugin,
				BotID:       botID,
				PluginName:  pluginName,
			}
			for i, matcher := range plugin.Matchers {
				displayName := data.GetRuleDisplayName(matcher.Rule)
				if displayName == "" {
					displayName = fmt.Sprintf("Matcher_%d", i)
				}
				matcherNode := &TreeNode{
					ID:          fmt.Sprintf("%s_matcher_%d", pluginNode.ID, i),
					DisplayName: displayName,
					Type:        NodeTypeMatcher,
					BotID:       botID,
					PluginName:  pluginName,
					MatcherIdx:  i,
				}
				pluginNode.Children = append(pluginNode.Children, matcherNode)
			}
			if len(plugin.Matchers) > 0 {
				botNode.Children = append(botNode.Children, pluginNode)
			}
		}
		if len(botNode.Children) > 0 {
			roots = append(roots, botNode)
		}
	}
	return roots
}
func BuildTreeFromConfigForBot(config *data.FullConfigModel, targetBotID string) []*TreeNode {
	var roots []*TreeNode
	if config == nil {
		return roots
	}
	if bot, exists := config.Bots[targetBotID]; exists {
		botNode := &TreeNode{
			ID:          "bot_" + targetBotID,
			DisplayName: targetBotID,
			Type:        NodeTypeBot,
			BotID:       targetBotID,
		}
		for pluginName, plugin := range bot.Plugins {
			pluginNode := &TreeNode{
				ID:          botNode.ID + "_plugin_" + pluginName,
				DisplayName: pluginName,
				Type:        NodeTypePlugin,
				BotID:       targetBotID,
				PluginName:  pluginName,
			}
			for i, matcher := range plugin.Matchers {
				displayName := data.GetRuleDisplayName(matcher.Rule)
				if displayName == "" {
					displayName = fmt.Sprintf("Matcher_%d", i)
				}
				matcherNode := &TreeNode{
					ID:          fmt.Sprintf("%s_matcher_%d", pluginNode.ID, i),
					DisplayName: displayName,
					Type:        NodeTypeMatcher,
					BotID:       targetBotID,
					PluginName:  pluginName,
					MatcherIdx:  i,
				}
				pluginNode.Children = append(pluginNode.Children, matcherNode)
			}
			if len(plugin.Matchers) > 0 {
				botNode.Children = append(botNode.Children, pluginNode)
			}
		}
		if len(botNode.Children) > 0 {
			roots = append(roots, botNode)
		}
	}
	return roots
}
func FindNode(nodes []*TreeNode, id string) *TreeNode {
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
		if len(node.Children) > 0 {
			if found := FindNode(node.Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}
