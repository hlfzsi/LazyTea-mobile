package common
import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	fyneTheme "fyne.io/fyne/v2/theme"
)
type MobileCard struct {
	widget.Card
}
func NewMobileCard(title, subtitle string, content fyne.CanvasObject) *MobileCard {
	card := &MobileCard{}
	card.ExtendBaseWidget(card)
	if content != nil {
		card.SetContent(content)
	}
	if title != "" {
		card.SetTitle(title)
	}
	if subtitle != "" {
		card.SetSubTitle(subtitle)
	}
	return card
}
func (c *MobileCard) CreateRenderer() fyne.WidgetRenderer {
	renderer := c.Card.CreateRenderer()
	return renderer
}
type MobileList struct {
	widget.List
	emptyText string
	emptyIcon fyne.Resource
}
func NewMobileList(length func() int, createItem func() fyne.CanvasObject, updateItem func(widget.ListItemID, fyne.CanvasObject)) *MobileList {
	list := &MobileList{}
	list.ExtendBaseWidget(list)
	list.Length = length
	list.CreateItem = createItem
	list.UpdateItem = updateItem
	list.emptyText = "暂无数据"
	list.emptyIcon = fyneTheme.InfoIcon()
	return list
}
func (l *MobileList) SetEmptyState(text string, icon fyne.Resource) {
	l.emptyText = text
	l.emptyIcon = icon
}
func (l *MobileList) CreateRenderer() fyne.WidgetRenderer {
	if l.Length() == 0 {
		emptyContainer := container.NewVBox(
			widget.NewIcon(l.emptyIcon),
			widget.NewLabel(l.emptyText),
		)
		return widget.NewSimpleRenderer(emptyContainer)
	}
	return l.List.CreateRenderer()
}
type MobileButton struct {
	widget.Button
}
func NewMobileButton(text string, icon fyne.Resource, tapped func()) *MobileButton {
	btn := &MobileButton{}
	btn.ExtendBaseWidget(btn)
	btn.SetText(text)
	btn.SetIcon(icon)
	btn.OnTapped = tapped
	return btn
}
func (b *MobileButton) MinSize() fyne.Size {
	baseSize := b.Button.MinSize()
	if baseSize.Height < 44 {
		baseSize.Height = 44
	}
	if baseSize.Width < 88 {
		baseSize.Width = 88
	}
	return baseSize
}
type LoadingIndicator struct {
	widget.ProgressBarInfinite
	label *widget.Label
}
func NewLoadingIndicator(text string) *LoadingIndicator {
	indicator := &LoadingIndicator{}
	indicator.label = widget.NewLabel(text)
	indicator.label.Alignment = fyne.TextAlignCenter
	return indicator
}
func (l *LoadingIndicator) SetText(text string) {
	l.label.SetText(text)
}
func (l *LoadingIndicator) CreateRenderer() fyne.WidgetRenderer {
	container := container.NewVBox(
		&l.ProgressBarInfinite,
		l.label,
	)
	return widget.NewSimpleRenderer(container)
}
type StatusBadge struct {
	widget.BaseWidget
	text       string
	background fyne.Resource
	importance widget.Importance
}
func NewStatusBadge(text string, importance widget.Importance) *StatusBadge {
	badge := &StatusBadge{
		text:       text,
		importance: importance,
	}
	badge.ExtendBaseWidget(badge)
	return badge
}
func (s *StatusBadge) SetText(text string) {
	s.text = text
	s.Refresh()
}
func (s *StatusBadge) SetImportance(importance widget.Importance) {
	s.importance = importance
	s.Refresh()
}
func (s *StatusBadge) CreateRenderer() fyne.WidgetRenderer {
	label := widget.NewLabel(s.text)
	label.Importance = s.importance
	label.Alignment = fyne.TextAlignCenter
	container := container.NewStack(label)
	return widget.NewSimpleRenderer(container)
}
func (s *StatusBadge) MinSize() fyne.Size {
	return fyne.NewSize(60, 24)
}