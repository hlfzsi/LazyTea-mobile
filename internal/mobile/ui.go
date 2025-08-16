package mobile
import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)
type UIManager struct {
	isMobile bool
}
func NewUIManager() *UIManager {
	return &UIManager{
		isMobile: IsMobile(),
	}
}
func (ui *UIManager) GetOptimalButtonSize() fyne.Size {
	if ui.isMobile {
		return fyne.NewSize(120, 48)
	}
	return fyne.NewSize(100, 36)
}
func (ui *UIManager) GetOptimalSpacing() float32 {
	if ui.isMobile {
		return 16.0
	}
	return 8.0
}
func (ui *UIManager) GetOptimalPadding() float32 {
	if ui.isMobile {
		return 20.0
	}
	return 12.0
}
func (ui *UIManager) GetOptimalMinSize() fyne.Size {
	if ui.isMobile {
		return fyne.NewSize(360, 640)
	}
	return fyne.NewSize(320, 480)
}
func (ui *UIManager) CreateTouchFriendlyButton(text string, onTapped func()) *widget.Button {
	btn := widget.NewButton(text, onTapped)
	if ui.isMobile {
		btn.Resize(ui.GetOptimalButtonSize())
	}
	return btn
}
func (ui *UIManager) CreateTouchFriendlyCard(title, subtitle string, content fyne.CanvasObject) *widget.Card {
	card := widget.NewCard(title, subtitle, content)
	if ui.isMobile {
		paddedContent := container.NewPadded(content)
		card.SetContent(paddedContent)
	}
	return card
}
func (ui *UIManager) CreateMobileOptimizedContainer(objects ...fyne.CanvasObject) *fyne.Container {
	if ui.isMobile {
		return container.NewVBox(objects...)
	}
	return container.NewVBox(objects...)
}
func (ui *UIManager) CreateMobileOptimizedScroll(content fyne.CanvasObject) *container.Scroll {
	scroll := container.NewScroll(content)
	if ui.isMobile {
		scroll.SetMinSize(ui.GetOptimalMinSize())
	}
	return scroll
}
func (ui *UIManager) CreateMobileOptimizedEntry() *widget.Entry {
	entry := widget.NewEntry()
	if ui.isMobile {
		entry.Resize(fyne.NewSize(250, 40))
	}
	return entry
}
func (ui *UIManager) CreateMobileOptimizedPasswordEntry() *widget.Entry {
	entry := widget.NewPasswordEntry()
	if ui.isMobile {
		entry.Resize(fyne.NewSize(250, 40))
	}
	return entry
}
func (ui *UIManager) ApplyMobileOptimizations(obj fyne.CanvasObject) {
	if !ui.isMobile {
		return
	}
	switch widget := obj.(type) {
	case *widget.Button:
		widget.Resize(ui.GetOptimalButtonSize())
	case *widget.Entry:
		widget.Resize(fyne.NewSize(250, 40))
	case *widget.Label:
	}
}
func (ui *UIManager) GetMobileNavigationHeight() float32 {
	if ui.isMobile {
		return 60.0
	}
	return 40.0
}
func (ui *UIManager) CreateMobileOptimizedTabs() *container.AppTabs {
	tabs := container.NewAppTabs()
	if ui.isMobile {
		tabs.SetTabLocation(container.TabLocationBottom)
	}
	return tabs
}
func (ui *UIManager) IsTouchInterface() bool {
	return ui.isMobile
}
func (ui *UIManager) GetDeviceType() string {
	if ui.isMobile {
		return "Mobile"
	}
	return "Desktop"
}
func (ui *UIManager) CreateMobileAlertDialog(title, message string, callback func(bool), parent fyne.Window) {
	content := container.NewVBox(
		widget.NewLabel(message),
	)
	if ui.isMobile {
		content.Resize(fyne.NewSize(320, 160))
	}
}
type TouchGestureHandler struct {
	onSwipeLeft  func()
	onSwipeRight func()
	onSwipeUp    func()
	onSwipeDown  func()
	onLongPress  func()
}
func NewTouchGestureHandler() *TouchGestureHandler {
	return &TouchGestureHandler{}
}
func (tgh *TouchGestureHandler) SetSwipeLeftCallback(callback func()) {
	tgh.onSwipeLeft = callback
}
func (tgh *TouchGestureHandler) SetSwipeRightCallback(callback func()) {
	tgh.onSwipeRight = callback
}
func (tgh *TouchGestureHandler) SetLongPressCallback(callback func()) {
	tgh.onLongPress = callback
}
