package common
import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)
type StatusIndicatorComponent struct {
	container   *fyne.Container
	statusIcon  *widget.Icon
	statusLabel *widget.Label
}
func NewStatusIndicatorComponent(initialText string) *StatusIndicatorComponent {
	statusIcon := widget.NewIcon(theme.CancelIcon())
	statusLabel := widget.NewLabel(initialText)
	container := container.NewHBox(statusIcon, statusLabel)
	return &StatusIndicatorComponent{
		container:   container,
		statusIcon:  statusIcon,
		statusLabel: statusLabel,
	}
}
func (c *StatusIndicatorComponent) SetStatus(connected bool, text string) {
	if connected {
		c.statusIcon.SetResource(theme.ConfirmIcon())
		c.statusLabel.Importance = widget.SuccessImportance
	} else {
		c.statusIcon.SetResource(theme.CancelIcon())
		c.statusLabel.Importance = widget.DangerImportance
	}
	c.statusLabel.SetText(text)
}
func (c *StatusIndicatorComponent) GetContainer() *fyne.Container {
	return c.container
}
type LoadingComponent struct {
	container *fyne.Container
	progress  *widget.ProgressBarInfinite
	label     *widget.Label
}
func NewLoadingComponent(text string) *LoadingComponent {
	progress := widget.NewProgressBarInfinite()
	label := widget.NewLabel(text)
	label.Alignment = fyne.TextAlignCenter
	container := container.NewVBox(
		progress,
		label,
	)
	return &LoadingComponent{
		container: container,
		progress:  progress,
		label:     label,
	}
}
func (c *LoadingComponent) Start() {
	c.progress.Start()
}
func (c *LoadingComponent) Stop() {
	c.progress.Stop()
}
func (c *LoadingComponent) SetText(text string) {
	c.label.SetText(text)
}
func (c *LoadingComponent) GetContainer() *fyne.Container {
	return c.container
}
type EmptyStateComponent struct {
	container *fyne.Container
}
func NewEmptyStateComponent(message string, icon fyne.Resource) *EmptyStateComponent {
	iconWidget := widget.NewIcon(icon)
	iconWidget.Resize(fyne.NewSize(48, 48))
	messageLabel := widget.NewLabel(message)
	messageLabel.Alignment = fyne.TextAlignCenter
	messageLabel.Importance = widget.MediumImportance
	container := container.NewVBox(
		container.NewCenter(iconWidget),
		container.NewCenter(messageLabel),
	)
	return &EmptyStateComponent{
		container: container,
	}
}
func (c *EmptyStateComponent) GetContainer() *fyne.Container {
	return c.container
}