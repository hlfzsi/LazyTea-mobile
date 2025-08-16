package pages
import (
	"encoding/json"
	"fmt"
	"lazytea-mobile/internal/data"
	"lazytea-mobile/internal/network"
	"lazytea-mobile/internal/ui/components/message"
	"lazytea-mobile/internal/utils"
	"strings"
	"time"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fyneTheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)
type MessagePage struct {
	*PageBase
	messageContainer *fyne.Container
	messageScroll    *container.Scroll
	messages         []data.Message
	messageBubbles   []*message.MessageBubble
	searchEntry      *widget.Entry
	autoScrollBtn    *widget.Button
	clearBtn         *widget.Button
	statusLabel      *widget.Label
	emptyLabel       *widget.Label
	autoScroll       bool
	isSearching      bool
	accentColor      string
}
func NewMessagePage(client *network.Client, storage *data.Storage, logger *utils.Logger) *MessagePage {
	page := &MessagePage{
		PageBase:       NewPageBase(client, storage, logger),
		messages:       make([]data.Message, 0),
		messageBubbles: make([]*message.MessageBubble, 0),
		autoScroll:     true,
		accentColor:    "#38A5FD",
	}
	page.setupUI()
	page.setupEventHandlers()
	page.loadRecentMessages()
	return page
}
func (p *MessagePage) setupUI() {
	titleLabel := widget.NewLabelWithStyle("消息记录", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	p.statusLabel = widget.NewLabel("正在加载...")
	p.statusLabel.Alignment = fyne.TextAlignCenter
	p.searchEntry = widget.NewEntry()
	p.searchEntry.SetPlaceHolder("搜索消息内容...")
	p.searchEntry.OnSubmitted = func(query string) {
		if strings.TrimSpace(query) != "" {
			p.searchMessages(query)
		} else {
			p.loadRecentMessages()
		}
	}
	searchBtn := widget.NewButtonWithIcon("", fyneTheme.SearchIcon(), func() {
		query := strings.TrimSpace(p.searchEntry.Text)
		if query != "" {
			p.searchMessages(query)
		}
	})
	p.clearBtn = widget.NewButtonWithIcon("", fyneTheme.CancelIcon(), func() {
		p.searchEntry.SetText("")
		p.loadRecentMessages()
	})
	p.clearBtn.Hide()  
	p.autoScrollBtn = widget.NewButtonWithIcon("自动滚动", fyneTheme.MoveDownIcon(), func() {
		p.toggleAutoScroll()
	})
	p.autoScrollBtn.Importance = widget.MediumImportance
	refreshBtn := widget.NewButtonWithIcon("", fyneTheme.ViewRefreshIcon(), func() {
		if p.isSearching {
			p.searchMessages(p.searchEntry.Text)
		} else {
			p.loadRecentMessages()
		}
	})
	searchContainer := container.NewBorder(
		nil, nil, nil,
		container.NewHBox(searchBtn, p.clearBtn, refreshBtn),
		p.searchEntry,
	)
	controlContainer := container.NewHBox(
		p.autoScrollBtn,
		widget.NewSeparator(),
		p.statusLabel,
	)
	p.messageContainer = container.NewVBox()
	p.messageScroll = container.NewScroll(p.messageContainer)
	p.emptyLabel = widget.NewLabel("暂无消息记录")
	p.emptyLabel.Alignment = fyne.TextAlignCenter
	p.emptyLabel.Hide()  
	topZone := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		searchContainer,
		controlContainer,
		widget.NewSeparator(),
	)
	content := container.NewBorder(
		topZone,
		nil,
		nil,
		nil,
		container.NewStack(
			p.messageScroll,
			p.emptyLabel,
		),
	)
	p.SetContent(content)
}
func (p *MessagePage) addMessageBubbleToContainer(msg data.Message) {
	bubble := message.NewMessageBubble(msg, p.accentColor)
	bubbleWithMargin := container.NewPadded(bubble)
	p.messageContainer.Add(bubbleWithMargin)
	if len(p.messageContainer.Objects) > 50 {
		p.messageContainer.Objects = p.messageContainer.Objects[1:]
		p.messageContainer.Refresh()
	}
	if p.autoScroll {
		p.scrollToBottom()
	}
}
func (p *MessagePage) clearMessageBubbles() {
	p.messageContainer.Objects = []fyne.CanvasObject{}
	p.messageContainer.Refresh()
}
func (p *MessagePage) scrollToBottom() {
	if len(p.messages) > 0 {
		p.messageScroll.ScrollToBottom()
	}
}
func (p *MessagePage) addMessageBubbleWithMeta(msg data.Message) {
	p.addMessageBubbleToContainer(msg)
}
func (p *MessagePage) setupEventHandlers() {
	p.client.OnConnectionChanged(func(connected bool) {
		if connected {
			p.statusLabel.SetText("已连接")
			p.statusLabel.Importance = widget.SuccessImportance
		} else {
			p.statusLabel.SetText("未连接")
			p.statusLabel.Importance = widget.DangerImportance
		}
	})
	p.client.OnMessage("message", func(header network.MessageHeader, payload interface{}) {
		p.handleNewMessage(payload)
	})
	p.client.OnMessage("call_api", func(header network.MessageHeader, payload interface{}) {
		p.handleNewMessage(payload)
	})
	p.client.OnMessage("plugin_call", func(header network.MessageHeader, payload interface{}) {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			p.logger.Error("Failed to marshal plugin_call payload: %v", err)
			return
		}
		var m map[string]interface{}
		if err := json.Unmarshal(payloadBytes, &m); err != nil {
			p.logger.Error("Failed to unmarshal plugin_call payload: %v", err)
			return
		}
		rec := data.PluginCallRecord{
			Bot:        p.getString(m, "bot"),
			Platform:   p.getString(m, "platform"),
			PluginName: p.getString(m, "plugin"),
			MatcherHash: func() string {
				if v, ok := m["matcher_hash"].(string); ok {
					return v
				}
				if arr, ok := m["matcher_hash"].([]interface{}); ok {  
					parts := make([]string, 0, len(arr))
					for _, it := range arr {
						parts = append(parts, p.getString(map[string]interface{}{"v": it}, "v"))
					}
					return strings.Join(parts, ",")
				}
				return ""
			}(),
			TimeCosted: func() float64 {
				if v, ok := m["time_costed"].(float64); ok {
					return v
				}
				return 0
			}(),
			Timestamp: func() int64 {
				if v, ok := m["time"].(float64); ok {
					return int64(v)
				}
				if v2, ok := m["time"].(int64); ok {
					return v2
				}
				return time.Now().Unix()
			}(),
		}
		if gid, ok := m["groupid"].(string); ok && gid != "" {
			rec.GroupID = &gid
		}
		if uid, ok := m["userid"].(string); ok && uid != "" {
			rec.UserID = &uid
		}
		if ex, ok := m["exception"].(map[string]interface{}); ok {
			if n, ok := ex["name"].(string); ok && n != "" {
				rec.ExceptionName = &n
			}
			if d, ok := ex["detail"].(string); ok && d != "" {
				rec.ExceptionDetail = &d
			}
		}
		if err := p.storage.SavePluginCall(rec); err != nil {
			p.logger.Error("Failed to save plugin_call record: %v", err)
		}
	})
}
func (p *MessagePage) handleNewMessage(payload interface{}) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		p.logger.Error("Failed to marshal message payload: %v", err)
		return
	}
	var msgData map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &msgData); err != nil {
		p.logger.Error("Failed to unmarshal message data: %v", err)
		return
	}
	metadata := map[string]interface{}{
		"bot":        []interface{}{p.getString(msgData, "bot"), "color: {bot_color}; font-weight: bold;"},
		"time":       []interface{}{msgData["time"], "color: #757575; font-size: 12px;"},
		"session":    []interface{}{fmt.Sprintf("会话：%s", p.getString(msgData, "session")), "color: #616161; font-style: italic;"},
		"avatar":     []interface{}{p.getString(msgData, "avatar"), 0},  
		"timestamps": []interface{}{msgData["time"], "hidden"},
	}
	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		p.logger.Error("Failed to marshal metadata: %v", err)
		return
	}
	metaStr := string(metaBytes)
	msg := data.Message{
		Bot:       p.getString(msgData, "bot"),
		BotID:     p.getString(msgData, "bot"),
		Content:   p.extractContent(msgData),
		FromBot:   msgData["from_bot"] == true,
		Timestamp: time.Now(),
		User:      p.getString(msgData, "userid"),
		UserID:    p.getString(msgData, "userid"),
		UserName:  p.getString(msgData, "username"),
		Plaintext: p.extractContent(msgData),  
		Meta:      &metaStr,                   
	}
	if groupID, ok := msgData["groupid"].(string); ok && groupID != "" {
		msg.GroupID = &groupID
		if groupName, ok := msgData["groupname"].(string); ok && groupName != "" {
			msg.GroupName = &groupName
		}
	}
	if tv, ok := msgData["time"].(float64); ok {
		if tv > 1e10 {  
			msg.Timestamps = int64(tv)
		} else {
			msg.Timestamps = int64(tv * 1000)
		}
	}
	if err := p.storage.SaveMessage(msg); err != nil {
		p.logger.Error("Failed to save message: %v", err)
	}
	if !p.isSearching && p.autoScroll {
		p.messages = append(p.messages, msg)
		p.addMessageBubbleWithMeta(msg)
	}
	go func() {
		if totalMsgs, err := p.storage.GetTotalMessageCount(); err == nil {
			p.statusLabel.SetText(fmt.Sprintf("共 %d 条消息", totalMsgs))
		} else {
			p.statusLabel.SetText(fmt.Sprintf("本次 %d 条消息", len(p.messages)))
		}
	}()
	p.updateEmptyState()  
}
func (p *MessagePage) getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
func (p *MessagePage) extractContent(data map[string]interface{}) string {
	if content, ok := data["content"]; ok {
		if contentSlice, ok := content.([]interface{}); ok {
			var contentParts []string
			for _, part := range contentSlice {
				if partSlice, ok := part.([]interface{}); ok && len(partSlice) >= 2 {
					if partType, ok := partSlice[0].(string); ok {
						if partType == "text" {
							if text, ok := partSlice[1].(string); ok {
								contentParts = append(contentParts, text)
							}
						} else {
							contentParts = append(contentParts, fmt.Sprintf("[%s]", partType))
						}
					}
				}
			}
			return strings.Join(contentParts, "")
		}
		if str, ok := content.(string); ok {
			return str
		}
	}
	return ""
}
func (p *MessagePage) loadRecentMessages() {
	p.isSearching = false
	p.clearBtn.Hide()
	p.statusLabel.SetText("正在加载...")
	go func() {
		messages, err := p.storage.GetMessages(50, 0)
		if err != nil {
			p.logger.Error("Failed to load messages: %v", err)
			p.statusLabel.SetText("加载失败")
			p.updateEmptyState()
			return
		}
		p.messages = messages
		p.clearMessageBubbles()
		if len(messages) > 0 {
			for _, msg := range messages {
				p.logger.Debug("Loading message - ID: %d, Content: '%s', Plaintext: '%s', User: '%s', Bot: '%s', Meta: '%s'",
					msg.ID, msg.Content, msg.Plaintext, msg.User, msg.Bot,
					func() string {
						if msg.Meta != nil {
							return *msg.Meta
						} else {
							return "nil"
						}
					}())
				p.addMessageBubbleToContainer(msg)
			}
			p.statusLabel.SetText(fmt.Sprintf("共 %d 条消息", len(p.messages)))
		} else {
			p.statusLabel.SetText("暂无消息记录")
		}
		p.updateEmptyState()  
	}()
}
func (p *MessagePage) searchMessages(query string) {
	p.isSearching = true
	p.clearBtn.Show()
	p.statusLabel.SetText("正在搜索...")
	go func() {
		messages, err := p.storage.SearchMessages(query, 100)
		if err != nil {
			p.logger.Error("Failed to search messages: %v", err)
			p.statusLabel.SetText("搜索失败")
			return
		}
		p.messages = messages
		p.clearMessageBubbles()
		for _, msg := range messages {
			p.addMessageBubbleToContainer(msg)
		}
		p.statusLabel.SetText(fmt.Sprintf("找到 %d 条匹配消息", len(p.messages)))
		p.updateEmptyState()  
	}()
}
func (p *MessagePage) toggleAutoScroll() {
	p.autoScroll = !p.autoScroll
	if p.autoScroll {
		p.autoScrollBtn.SetText("自动滚动")
		p.autoScrollBtn.SetIcon(fyneTheme.MoveDownIcon())
		p.autoScrollBtn.Importance = widget.MediumImportance
		p.scrollToBottom()
	} else {
		p.autoScrollBtn.SetText("手动浏览")
		p.autoScrollBtn.SetIcon(fyneTheme.NavigateNextIcon())
		p.autoScrollBtn.Importance = widget.LowImportance
	}
}
func (p *MessagePage) updateEmptyState() {
	if len(p.messages) == 0 {
		p.messageScroll.Hide()
		p.emptyLabel.Show()
	} else {
		p.emptyLabel.Hide()
		p.messageScroll.Show()
	}
}
