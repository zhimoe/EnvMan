package app

import (
	"github.com/egoist/mygo/ui"
)

// buildTheme returns the app's compact theme and its color palette.
func (a *App) buildTheme(c *ui.Context) (*ui.Theme, palette) {
	prefs := c.Preferences()
	dark := c.Theme().Dark
	switch a.themeMode {
	case ThemeLight:
		dark = false
	case ThemeDark:
		dark = true
	}
	p := lightPalette()
	if dark {
		p = darkPalette()
	}
	if prefs.HighContrast {
		p = contrastPalette(dark)
	}
	scale := prefs.TextScale
	if scale < 1 {
		scale = 1
	}
	a.metrics = newMetrics(scale)
	base := ui.LightTheme()
	if dark {
		base = ui.DarkTheme()
	}
	t := *base
	t.Dark = dark
	t.Spacing = space4
	t.FontSize = a.metrics.fontBody
	t.Radius = radiusControl
	t.Font = ""
	t.Background = p.window
	t.Surface = p.surface
	t.SurfaceHover = p.surface2
	t.SurfacePressed = p.surface3
	t.Border = p.line
	t.Text = p.text
	t.TextMuted = p.text3
	t.Accent = p.accent
	t.AccentHover = p.accentHover
	t.AccentPressed = p.accentPress
	t.AccentText = p.accentText
	t.Danger = p.danger
	t.Warning = p.warn
	t.Success = p.success
	t.Selection = p.accentSoft2
	t.Focus = p.accent.Alpha(0.55)
	t.Scrollbar = p.text3.Alpha(0.35)
	t.ScrollbarWidth = 6
	if dark {
		t.Inverse = ui.Hex("#e9ebf1")
		t.InverseText = ui.Hex("#17181e")
	} else {
		t.Inverse = ui.Hex("#1d1f26")
		t.InverseText = ui.Hex("#f3f4f7")
	}
	if prefs.HighContrast {
		t.Focus = p.text
		t.Scrollbar = p.text
	}
	return &t, p
}
