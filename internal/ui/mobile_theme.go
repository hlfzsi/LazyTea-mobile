package ui
import (
	"image/color"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)
type MobileTheme struct {
	isDark bool
}
var _ fyne.Theme = (*MobileTheme)(nil)
func NewMobileTheme(isDark bool) fyne.Theme {
	return &MobileTheme{isDark: isDark}
}
func (m *MobileTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	isDark := m.isDark || variant == theme.VariantDark
	switch name {
	case theme.ColorNamePrimary:
		if isDark {
			return color.RGBA{96, 165, 250, 255}  
		}
		return color.RGBA{59, 130, 246, 255}  
	case theme.ColorNameBackground:
		if isDark {
			return color.RGBA{15, 15, 17, 255}  
		}
		return color.RGBA{248, 250, 252, 255}  
	case theme.ColorNameForeground:
		if isDark {
			return color.RGBA{248, 250, 252, 255}  
		}
		return color.RGBA{15, 23, 42, 255}  
	case theme.ColorNameButton:
		if isDark {
			return color.RGBA{30, 41, 59, 255}  
		}
		return color.RGBA{255, 255, 255, 255}  
	case theme.ColorNameDisabledButton:
		if isDark {
			return color.RGBA{71, 85, 105, 255}  
		}
		return color.RGBA{226, 232, 240, 255}  
	case theme.ColorNamePlaceHolder:
		if isDark {
			return color.RGBA{148, 163, 184, 255}  
		}
		return color.RGBA{100, 116, 139, 255}  
	case theme.ColorNamePressed:
		if isDark {
			return color.RGBA{51, 65, 85, 255}  
		}
		return color.RGBA{241, 245, 249, 255}  
	case theme.ColorNameSelection:
		if isDark {
			return color.RGBA{96, 165, 250, 100}  
		}
		return color.RGBA{59, 130, 246, 80}  
	case theme.ColorNameSeparator:
		if isDark {
			return color.RGBA{51, 65, 85, 255}  
		}
		return color.RGBA{226, 232, 240, 255}  
	case theme.ColorNameSuccess:
		return color.RGBA{34, 197, 94, 255}  
	case theme.ColorNameError:
		return color.RGBA{239, 68, 68, 255}  
	case theme.ColorNameWarning:
		return color.RGBA{245, 158, 11, 255}  
	case theme.ColorNameInputBackground:
		if isDark {
			return color.RGBA{30, 41, 59, 255}  
		}
		return color.RGBA{255, 255, 255, 255}  
	case theme.ColorNameInputBorder:
		if isDark {
			return color.RGBA{71, 85, 105, 255}  
		}
		return color.RGBA{203, 213, 225, 255}  
	case theme.ColorNameHover:
		if isDark {
			return color.RGBA{41, 55, 75, 255}  
		}
		return color.RGBA{248, 250, 252, 255}  
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}
func (m *MobileTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}
func (m *MobileTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
func (m *MobileTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 15  
	case theme.SizeNameCaptionText:
		return 13  
	case theme.SizeNameHeadingText:
		return 22  
	case theme.SizeNameSubHeadingText:
		return 19  
	case theme.SizeNameInlineIcon:
		return 20  
	case theme.SizeNameInputBorder:
		return 1  
	case theme.SizeNameScrollBar:
		return 14  
	case theme.SizeNameScrollBarSmall:
		return 10
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNamePadding:
		return 8  
	case theme.SizeNameInnerPadding:
		return 4  
	case theme.SizeNameInputRadius:
		return 8  
	default:
		return theme.DefaultTheme().Size(name)
	}
}
