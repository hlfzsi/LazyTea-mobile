package mobile
import (
	"image/color"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)
type MobileTheme struct {
	baseTheme fyne.Theme
	isMobile  bool
}
func NewMobileTheme() fyne.Theme {
	return &MobileTheme{
		baseTheme: theme.DefaultTheme(),
		isMobile:  IsMobile(),
	}
}
func (mt *MobileTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return mt.baseTheme.Color(name, variant)
}
func (mt *MobileTheme) Font(style fyne.TextStyle) fyne.Resource {
	return mt.baseTheme.Font(style)
}
func (mt *MobileTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return mt.baseTheme.Icon(name)
}
func (mt *MobileTheme) Size(name fyne.ThemeSizeName) float32 {
	if !mt.isMobile {
		return mt.baseTheme.Size(name)
	}
	switch name {
	case theme.SizeNameText:
		return 16  
	case theme.SizeNameCaptionText:
		return 14  
	case theme.SizeNameHeadingText:
		return 20  
	case theme.SizeNameSubHeadingText:
		return 18  
	case theme.SizeNamePadding:
		return 8  
	case theme.SizeNameInnerPadding:
		return 6  
	case theme.SizeNameScrollBar:
		return 20  
	case theme.SizeNameScrollBarSmall:
		return 16  
	case theme.SizeNameSeparatorThickness:
		return 2  
	case theme.SizeNameInputBorder:
		return 2  
	case theme.SizeNameInputRadius:
		return 6  
	case theme.SizeNameSelectionRadius:
		return 4  
	default:
		return mt.baseTheme.Size(name)
	}
}
type MobileColorScheme struct {
	Primary       color.Color
	Secondary     color.Color
	Background    color.Color
	Surface       color.Color
	Error         color.Color
	OnPrimary     color.Color
	OnSecondary   color.Color
	OnBackground  color.Color
	OnSurface     color.Color
	OnError       color.Color
}
func GetMobileColorScheme() *MobileColorScheme {
	return &MobileColorScheme{
		Primary:      color.RGBA{R: 33, G: 150, B: 243, A: 255},    
		Secondary:    color.RGBA{R: 255, G: 193, B: 7, A: 255},     
		Background:   color.RGBA{R: 250, G: 250, B: 250, A: 255},  
		Surface:      color.RGBA{R: 255, G: 255, B: 255, A: 255},  
		Error:        color.RGBA{R: 244, G: 67, B: 54, A: 255},     
		OnPrimary:    color.RGBA{R: 255, G: 255, B: 255, A: 255},  
		OnSecondary:  color.RGBA{R: 0, G: 0, B: 0, A: 255},        
		OnBackground: color.RGBA{R: 0, G: 0, B: 0, A: 255},        
		OnSurface:    color.RGBA{R: 0, G: 0, B: 0, A: 255},        
		OnError:      color.RGBA{R: 255, G: 255, B: 255, A: 255},  
	}
}
func GetMobileDarkColorScheme() *MobileColorScheme {
	return &MobileColorScheme{
		Primary:      color.RGBA{R: 33, G: 150, B: 243, A: 255},   
		Secondary:    color.RGBA{R: 255, G: 193, B: 7, A: 255},    
		Background:   color.RGBA{R: 18, G: 18, B: 18, A: 255},     
		Surface:      color.RGBA{R: 33, G: 33, B: 33, A: 255},     
		Error:        color.RGBA{R: 244, G: 67, B: 54, A: 255},    
		OnPrimary:    color.RGBA{R: 255, G: 255, B: 255, A: 255},  
		OnSecondary:  color.RGBA{R: 0, G: 0, B: 0, A: 255},        
		OnBackground: color.RGBA{R: 255, G: 255, B: 255, A: 255},  
		OnSurface:    color.RGBA{R: 255, G: 255, B: 255, A: 255},  
		OnError:      color.RGBA{R: 255, G: 255, B: 255, A: 255},  
	}
}
func ApplyMobileTheme(app fyne.App) {
	if IsMobile() {
		mobileTheme := NewMobileTheme()
		app.Settings().SetTheme(mobileTheme)
	}
}
func GetMobileFontSizes() map[string]float32 {
	if IsMobile() {
		return map[string]float32{
			"caption":    14,
			"body":       16,
			"subtitle":   18,
			"title":      20,
			"headline":   24,
			"display":    28,
		}
	}
	return map[string]float32{
		"caption":    11,
		"body":       13,
		"subtitle":   15,
		"title":      17,
		"headline":   20,
		"display":    24,
	}
}
func GetMobileIconSizes() map[string]float32 {
	if IsMobile() {
		return map[string]float32{
			"small":  20,
			"medium": 24,
			"large":  32,
			"xlarge": 48,
		}
	}
	return map[string]float32{
		"small":  16,
		"medium": 20,
		"large":  24,
		"xlarge": 32,
	}
}