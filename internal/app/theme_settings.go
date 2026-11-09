package app

import "github.com/egoist/mygo"

type ThemeMode int

const (
	ThemeSystem ThemeMode = iota
	ThemeLight
	ThemeDark
)

func (m ThemeMode) String() string {
	switch m {
	case ThemeLight:
		return "light"
	case ThemeDark:
		return "dark"
	default:
		return "system"
	}
}

func parseThemeMode(s string) ThemeMode {
	switch s {
	case "light":
		return ThemeLight
	case "dark":
		return ThemeDark
	default:
		return ThemeSystem
	}
}

type appSettings struct {
	Theme string `json:"theme,omitempty"`
}

func (a *App) toggleTheme() {
	if a.themeMode == ThemeSystem {
		if a.lastDark {
			a.themeMode = ThemeLight
		} else {
			a.themeMode = ThemeDark
		}
	} else if a.themeMode == ThemeDark {
		a.themeMode = ThemeLight
	} else {
		a.themeMode = ThemeDark
	}
	a.saveSettings()
	if a.win != nil {
		src := mygo.ThemeLight
		if a.themeMode == ThemeDark {
			src = mygo.ThemeDark
		}
		a.win.SetWindowButtonsTheme(src)
	}
}

func (a *App) saveSettings() {
	_ = a.store.SaveSettings(appSettings{Theme: a.themeMode.String()})
}

func (a *App) loadSettings() appSettings {
	var settings appSettings
	_ = a.store.LoadSettings(&settings)
	return settings
}
