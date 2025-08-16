package ui
import (
	"image/color"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)
type LazyTeaTheme struct {
	isDark bool
}
func NewLazyTeaTheme(isDark bool) *LazyTeaTheme {
	return &LazyTeaTheme{isDark: isDark}
}
var _ fyne.Theme = (*LazyTeaTheme)(nil)
func (m *LazyTeaTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	isDark := m.isDark || variant == theme.VariantDark
	switch name {
	case theme.ColorNamePrimary:
		if isDark {
			return color.RGBA{56, 165, 253, 255}  
		}
		return color.RGBA{56, 165, 253, 255}  
	case theme.ColorNameBackground:
		if isDark {
			return color.RGBA{24, 24, 27, 255}  
		}
		return color.RGBA{250, 250, 250, 255}  
	case theme.ColorNameForeground:
		if isDark {
			return color.RGBA{255, 255, 255, 255}  
		}
		return color.RGBA{34, 34, 34, 255}  
	case theme.ColorNameButton:
		if isDark {
			return color.RGBA{56, 165, 253, 255}  
		}
		return color.RGBA{56, 165, 253, 255}  
	case theme.ColorNameDisabled:
		if isDark {
			return color.RGBA{120, 120, 120, 255}  
		}
		return color.RGBA{150, 150, 150, 255}  
	case theme.ColorNameError:
		return color.RGBA{244, 67, 54, 255}  
	case theme.ColorNameSuccess:
		return color.RGBA{76, 175, 80, 255}  
	case theme.ColorNameWarning:
		return color.RGBA{255, 152, 0, 255}  
	case theme.ColorNameHover:
		if isDark {
			return color.RGBA{56, 58, 64, 255}  
		}
		return color.RGBA{240, 244, 248, 255}  
	case theme.ColorNameInputBackground:
		if isDark {
			return color.RGBA{40, 44, 52, 255}  
		}
		return color.RGBA{255, 255, 255, 255}  
	case theme.ColorNameInputBorder:
		if isDark {
			return color.RGBA{96, 96, 96, 255}  
		}
		return color.RGBA{224, 224, 224, 255}  
	case theme.ColorNameMenuBackground:
		if isDark {
			return color.RGBA{32, 32, 35, 255}  
		}
		return color.RGBA{255, 255, 255, 255}  
	case theme.ColorNameOverlayBackground:
		return color.RGBA{0, 0, 0, 128}  
	case theme.ColorNameSeparator:
		if isDark {
			return color.RGBA{64, 64, 64, 255}  
		}
		return color.RGBA{224, 224, 224, 255}  
	}
	return theme.DefaultTheme().Color(name, variant)
}
func (m *LazyTeaTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return theme.DefaultTheme().Font(style)
	}
	if style.Bold {
		return theme.DefaultTheme().Font(style)
	}
	return theme.DefaultTheme().Font(style)
}
func (m *LazyTeaTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
func (m *LazyTeaTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 14  
	case theme.SizeNameCaptionText:
		return 11  
	case theme.SizeNameHeadingText:
		return 18  
	case theme.SizeNameSubHeadingText:
		return 16  
	case theme.SizeNamePadding:
		return 8  
	case theme.SizeNameInnerPadding:
		return 4  
	case theme.SizeNameScrollBar:
		return 12  
	case theme.SizeNameScrollBarSmall:
		return 8  
	case theme.SizeNameSeparatorThickness:
		return 1  
	case theme.SizeNameInputBorder:
		return 1  
	case theme.SizeNameInputRadius:
		return 6  
	}
	return theme.DefaultTheme().Size(name)
}
var (
	ColorPrimary     = color.RGBA{56, 165, 253, 255}   
	ColorSecondary   = color.RGBA{108, 92, 231, 255}   
	ColorAccent      = color.RGBA{255, 159, 67, 255}   
	ColorSuccess     = color.RGBA{46, 213, 115, 255}   
	ColorWarning     = color.RGBA{255, 159, 67, 255}   
	ColorError       = color.RGBA{255, 71, 87, 255}    
	ColorInfo        = color.RGBA{52, 152, 219, 255}   
	ColorGray100     = color.RGBA{248, 249, 250, 255}  
	ColorGray200     = color.RGBA{233, 236, 239, 255}  
	ColorGray300     = color.RGBA{206, 212, 218, 255}  
	ColorGray400     = color.RGBA{173, 181, 189, 255}  
	ColorGray500     = color.RGBA{108, 117, 125, 255}  
	ColorGray600     = color.RGBA{73, 80, 87, 255}     
	ColorGray700     = color.RGBA{52, 58, 64, 255}     
	ColorGray800     = color.RGBA{33, 37, 41, 255}     
	ColorGray900     = color.RGBA{13, 16, 23, 255}     
	ColorBackgroundLight = color.RGBA{255, 255, 255, 255}  
	ColorBackgroundDark  = color.RGBA{24, 24, 27, 255}     
	ColorSurfaceLight    = color.RGBA{248, 249, 250, 255}  
	ColorSurfaceDark     = color.RGBA{39, 39, 42, 255}     
)
func GetBotColor(botID string) color.Color {
	colors := []color.Color{
		ColorPrimary,
		ColorSecondary, 
		ColorAccent,
		color.RGBA{229, 115, 115, 255},  
		color.RGBA{149, 117, 205, 255},  
		color.RGBA{129, 199, 132, 255},  
		color.RGBA{255, 183, 77, 255},   
		color.RGBA{100, 181, 246, 255},  
		color.RGBA{240, 98, 146, 255},   
		color.RGBA{77, 182, 172, 255},   
	}
	hash := 0
	for _, c := range botID {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return colors[hash%len(colors)]
}