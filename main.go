package main
import (
	"lazytea-mobile/internal/app"
	"os"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)
func main() {
	os.Setenv("UIVERSION", "0.1.0a1")
	os.Setenv("UIAUTHOR", "hlfzsi")
	fyneApplication := fyneApp.NewWithID("com.lazytea.mobile")
	fyneApplication.Settings().SetTheme(theme.DarkTheme())  
	lazyteaApp := app.NewApp(fyneApplication)
	lazyteaApp.Run()
}
