package message

import (
	"encoding/json"
	"fmt"
	"lazytea-mobile/internal/data"
	"time"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MessageBubble struct {
	widget.BaseWidget
	message     data.Message
	accentColor string
	isFromBot   bool
	content     *widget.RichText
	timeLabel   *widget.Label
	senderLabel *widget.Label
	bubbleCard  fyne.CanvasObject
	metadata    map[string]interface{}
}

func NewMessageBubble(message data.Message, accentColor string) *MessageBubble {
	isFromBot := message.FromBot || (message.User == "" && message.Bot != "")
	mb := &MessageBubble{
		message:     message,
		accentColor: accentColor,
		isFromBot:   isFromBot,
	}
	mb.ExtendBaseWidget(mb)
	mb.setupUI()
	return mb
}
func NewMessageBubbleWithMeta(message data.Message, metadata map[string]interface{}, accentColor string) *MessageBubble {
	mb := &MessageBubble{
		message:     message,
		accentColor: accentColor,
		isFromBot:   message.FromBot,
		metadata:    metadata,
	}
	mb.ExtendBaseWidget(mb)
	mb.setupUIWithMeta()
	return mb
}
func (mb *MessageBubble) setupUI() {
	mb.timeLabel = widget.NewLabel(mb.formatTime())
	mb.timeLabel.Importance = widget.LowImportance
	mb.senderLabel = widget.NewLabel(mb.getSenderText())
	mb.senderLabel.TextStyle = fyne.TextStyle{Bold: true}
	content := mb.message.Content
	if content == "" {
		content = mb.message.Plaintext
	}
	if content == "" {
		content = " "
	}
	fmt.Printf("MessageBubble - ID: %d, Content: '%s', Plaintext: '%s', FinalContent: '%s'\n",
		mb.message.ID, mb.message.Content, mb.message.Plaintext, content)
	mb.content = widget.NewRichTextFromMarkdown(content)
	mb.content.Wrapping = fyne.TextWrapWord
	mb.setDarkTextColors()
	if mb.isFromBot {
		mb.setupBotStyle()
	} else {
		mb.setupUserStyle()
	}
	mb.createBubbleCard()
}
func (mb *MessageBubble) setupUIWithMeta() {
	botInfo := mb.getStringFromMeta("bot")
	timeInfo := mb.getStringFromMeta("time")
	if timeInfo != "" {
		mb.timeLabel = widget.NewLabel(timeInfo)
	} else {
		mb.timeLabel = widget.NewLabel(mb.formatTime())
	}
	mb.timeLabel.Importance = widget.LowImportance
	if botInfo != "" {
		mb.senderLabel = widget.NewLabel("🤖 " + truncateString(botInfo, 15))
		mb.senderLabel.Importance = widget.SuccessImportance
		mb.isFromBot = true
	} else {
		userName := mb.getSenderText()
		mb.senderLabel = widget.NewLabel("👤 " + truncateString(userName, 15))
		mb.senderLabel.Importance = widget.HighImportance
	}
	mb.senderLabel.TextStyle = fyne.TextStyle{Bold: true}
	content := mb.message.Content
	if content == "" {
		content = mb.message.Plaintext
	}
	if content == "" {
		content = " "
	}
	fmt.Printf("MessageBubbleWithMeta - ID: %d, Content: '%s', BotInfo: '%s', TimeInfo: '%s'\n",
		mb.message.ID, content, botInfo, timeInfo)
	mb.content = widget.NewRichTextFromMarkdown(content)
	mb.content.Wrapping = fyne.TextWrapWord
	mb.setDarkTextColors()
	mb.createBubbleCard()
}
func (mb *MessageBubble) setupBotStyle() {
	botName := mb.message.BotID
	if botName == "" {
		botName = mb.message.Bot
	}
	mb.senderLabel.SetText("🤖 " + truncateString(botName, 15))
	mb.senderLabel.Importance = widget.SuccessImportance
}
func (mb *MessageBubble) setupUserStyle() {
	if mb.message.GroupID != nil && *mb.message.GroupID != "" {
		groupName := *mb.message.GroupID
		if mb.message.GroupName != nil && *mb.message.GroupName != "" {
			groupName = *mb.message.GroupName
		}
		mb.senderLabel.SetText(fmt.Sprintf("👥 %s", truncateString(groupName, 15)))
		mb.senderLabel.Importance = widget.MediumImportance
		userName := mb.message.UserName
		if userName == "" {
			userName = mb.message.User
		}
		if userName == "" {
			userName = mb.message.UserID
		}
		if userName == "" {
			userName = "未知用户"
		}
		originalContent := mb.message.Content
		mb.content.ParseMarkdown(fmt.Sprintf("**%s**: %s", truncateString(userName, 15), originalContent))
	} else {
		userName := mb.message.UserName
		if userName == "" {
			userName = mb.message.User
		}
		if userName == "" {
			userName = mb.message.UserID
		}
		if userName == "" {
			userName = "未知用户"
		}
		mb.senderLabel.SetText("👤 " + truncateString(userName, 15))
		mb.senderLabel.Importance = widget.HighImportance
	}
}
func (mb *MessageBubble) createBubbleCard() {
	header := container.NewBorder(nil, nil, mb.senderLabel, mb.timeLabel)
	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		mb.content,
	)
	contentWithPadding := container.NewPadded(content)
	rect := canvas.NewRectangle(theme.BackgroundColor())
	rect.StrokeColor = theme.InputBorderColor()
	rect.StrokeWidth = 1
	rect.CornerRadius = theme.Padding()
	mb.bubbleCard = container.NewStack(rect, contentWithPadding)
}
func (mb *MessageBubble) MinSize() fyne.Size {
	if mb.bubbleCard == nil {
		return fyne.NewSize(280, 120)
	}
	minSize := mb.bubbleCard.MinSize()
	if minSize.Width < 280 {
		minSize.Width = 280
	}
	if minSize.Height < 130 {
		minSize.Height = 130
	}
	return minSize
}
func (mb *MessageBubble) formatTime() string {
	if !mb.message.Timestamp.IsZero() {
		return mb.message.Timestamp.Format("01-02 15:04:05")
	}
	if mb.message.Timestamps > 0 {
		return time.UnixMilli(mb.message.Timestamps).Format("01-02 15:04:05")
	}
	return time.Now().Format("01-02 15:04:05")
}
func truncateString(s string, length int) string {
	if utf8.RuneCountInString(s) > length {
		runes := []rune(s)
		return string(runes[:length]) + "..."
	}
	return s
}
func (mb *MessageBubble) getSenderText() string {
	if mb.isFromBot {
		if mb.message.BotID != "" {
			return mb.message.BotID
		}
		return mb.message.Bot
	}
	if mb.message.UserName != "" {
		return mb.message.UserName
	}
	if mb.message.User != "" {
		return mb.message.User
	}
	if mb.message.UserID != "" {
		return mb.message.UserID
	}
	return "未知用户"
}
func (mb *MessageBubble) GetContent() string {
	return mb.message.Content
}
func (mb *MessageBubble) GetPlaintext() string {
	if mb.message.Plaintext != "" {
		return mb.message.Plaintext
	}
	return mb.message.Content
}
func (mb *MessageBubble) CreateRenderer() fyne.WidgetRenderer {
	fmt.Printf("CreateRenderer called for message ID: %d\n", mb.message.ID)
	if mb.bubbleCard == nil {
		fmt.Printf("Warning: bubbleCard is nil in CreateRenderer\n")
		mb.bubbleCard = widget.NewLabel("Loading...")
	}
	return widget.NewSimpleRenderer(mb.bubbleCard)
}
func (mb *MessageBubble) Tapped(*fyne.PointEvent) {
}
func (mb *MessageBubble) TappedSecondary(*fyne.PointEvent) {
	mb.showContextMenu()
}
func (mb *MessageBubble) showContextMenu() {
}
func (mb *MessageBubble) UpdateMessage(message data.Message) {
	fmt.Printf("UpdateMessage called - ID: %d, Content: '%s', User: '%s'\n",
		message.ID, message.Content, message.User)
	mb.message = message
	if message.Meta != nil && *message.Meta != "" {
		fmt.Printf("Using setupUIWithMeta for message %d\n", message.ID)
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(*message.Meta), &metadata); err == nil {
			mb.metadata = metadata
			mb.setupUIWithMeta()
		} else {
			fmt.Printf("Failed to parse metadata, using setupUI for message %d\n", message.ID)
			mb.setupUI()
		}
	} else {
		fmt.Printf("Using setupUI for message %d\n", message.ID)
		mb.setupUI()
	}
	fmt.Printf("After setup - BubbleCard is nil: %v\n", mb.bubbleCard == nil)
	mb.Refresh()
	if mb.bubbleCard != nil {
		mb.bubbleCard.Refresh()
	}
	if mb.senderLabel != nil {
		mb.senderLabel.Refresh()
	}
	if mb.timeLabel != nil {
		mb.timeLabel.Refresh()
	}
	if mb.content != nil {
		mb.content.Refresh()
	}
}
func (mb *MessageBubble) SetAccentColor(color string) {
	mb.accentColor = color
}
func (mb *MessageBubble) GetMessage() data.Message {
	return mb.message
}
func (mb *MessageBubble) getStringFromMeta(key string) string {
	if mb.metadata == nil {
		return ""
	}
	value, exists := mb.metadata[key]
	if !exists {
		return ""
	}
	if arr, ok := value.([]interface{}); ok && len(arr) > 0 {
		if str, ok := arr[0].(string); ok {
			return str
		}
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}
func (mb *MessageBubble) setDarkTextColors() {
	if mb.senderLabel != nil {
		mb.senderLabel.TextStyle.Bold = true
	}
	if mb.timeLabel != nil {
		mb.timeLabel.Importance = widget.LowImportance
	}
	if mb.content != nil {
		content := mb.message.Content
		if content == "" {
			content = mb.message.Plaintext
		}
		if content != "" {
			mb.content.ParseMarkdown("**" + content + "**")
		}
	}
}
