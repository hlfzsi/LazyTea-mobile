package ui
import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)
type ResponsiveContainer struct {
	widget.BaseWidget
	content         fyne.CanvasObject
	mobileContent   fyne.CanvasObject
	tabletContent   fyne.CanvasObject
	desktopContent  fyne.CanvasObject
	currentMode     ScreenMode
	breakpoints     Breakpoints
}
type ScreenMode int
const (
	ScreenMobile ScreenMode = iota
	ScreenTablet
	ScreenDesktop
)
type Breakpoints struct {
	Mobile  float32  
	Tablet  float32  
	Desktop float32  
}
var DefaultBreakpoints = Breakpoints{
	Mobile:  600,
	Tablet:  1024,
	Desktop: 1024,
}
func NewResponsiveContainer() *ResponsiveContainer {
	rc := &ResponsiveContainer{
		breakpoints: DefaultBreakpoints,
		currentMode: ScreenMobile,
	}
	rc.ExtendBaseWidget(rc)
	return rc
}
func (rc *ResponsiveContainer) SetContent(mobile, tablet, desktop fyne.CanvasObject) {
	rc.mobileContent = mobile
	rc.tabletContent = tablet
	rc.desktopContent = desktop
	rc.updateContent()
}
func (rc *ResponsiveContainer) updateContent() {
	switch rc.currentMode {
	case ScreenMobile:
		if rc.mobileContent != nil {
			rc.content = rc.mobileContent
		} else {
			rc.content = rc.getDefaultContent()
		}
	case ScreenTablet:
		if rc.tabletContent != nil {
			rc.content = rc.tabletContent
		} else if rc.mobileContent != nil {
			rc.content = rc.mobileContent
		} else {
			rc.content = rc.getDefaultContent()
		}
	case ScreenDesktop:
		if rc.desktopContent != nil {
			rc.content = rc.desktopContent
		} else if rc.tabletContent != nil {
			rc.content = rc.tabletContent
		} else if rc.mobileContent != nil {
			rc.content = rc.mobileContent
		} else {
			rc.content = rc.getDefaultContent()
		}
	}
	rc.Refresh()
}
func (rc *ResponsiveContainer) getDefaultContent() fyne.CanvasObject {
	return widget.NewLabel("No content set")
}
func (rc *ResponsiveContainer) Resize(size fyne.Size) {
	rc.BaseWidget.Resize(size)
	newMode := rc.getScreenMode(size.Width)
	if newMode != rc.currentMode {
		rc.currentMode = newMode
		rc.updateContent()
	}
	if rc.content != nil {
		rc.content.Resize(size)
	}
}
func (rc *ResponsiveContainer) getScreenMode(width float32) ScreenMode {
	if width <= rc.breakpoints.Mobile {
		return ScreenMobile
	} else if width <= rc.breakpoints.Tablet {
		return ScreenTablet
	}
	return ScreenDesktop
}
func (rc *ResponsiveContainer) CreateRenderer() fyne.WidgetRenderer {
	if rc.content == nil {
		rc.content = rc.getDefaultContent()
	}
	return widget.NewSimpleRenderer(rc.content)
}
type MobileLayout struct {
	objects []fyne.CanvasObject
}
func NewMobileLayout(objects ...fyne.CanvasObject) *MobileLayout {
	return &MobileLayout{objects: objects}
}
func (ml *MobileLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	if len(objects) == 0 {
		return
	}
	padding := float32(16)
	spacing := float32(12)
	y := padding
	itemHeight := (containerSize.Height - padding*2 - spacing*float32(len(objects)-1)) / float32(len(objects))
	for _, obj := range objects {
		if obj.Visible() {
			obj.Move(fyne.NewPos(padding, y))
			obj.Resize(fyne.NewSize(containerSize.Width-padding*2, itemHeight))
			y += itemHeight + spacing
		}
	}
}
func (ml *MobileLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minWidth := float32(320)  
	minHeight := float32(480)  
	for _, obj := range objects {
		if obj.Visible() {
			objMin := obj.MinSize()
			if objMin.Width > minWidth {
				minWidth = objMin.Width
			}
			minHeight += objMin.Height + 12  
		}
	}
	return fyne.NewSize(minWidth, minHeight+32)  
}
type TouchOptimizedButton struct {
	*widget.Button
}
func NewTouchOptimizedButton(text string, tapped func()) *TouchOptimizedButton {
	btn := &TouchOptimizedButton{
		Button: widget.NewButton(text, tapped),
	}
	btn.Resize(fyne.NewSize(48, 48))
	return btn
}
func (tb *TouchOptimizedButton) MinSize() fyne.Size {
	min := tb.Button.MinSize()
	if min.Width < 48 {
		min.Width = 48
	}
	if min.Height < 48 {
		min.Height = 48
	}
	return min
}
type SwipeContainer struct {
	widget.BaseWidget
	content    *fyne.Container
	onSwipeLeft func()
	onSwipeRight func()
}
func NewSwipeContainer(content *fyne.Container) *SwipeContainer {
	sc := &SwipeContainer{
		content: content,
	}
	sc.ExtendBaseWidget(sc)
	return sc
}
func (sc *SwipeContainer) SetSwipeCallbacks(left, right func()) {
	sc.onSwipeLeft = left
	sc.onSwipeRight = right
}
func (sc *SwipeContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(sc.content)
}