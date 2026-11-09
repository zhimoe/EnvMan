package app

import (
	"github.com/egoist/mygo/ui"
	"math"
	"testing"
)

func TestTextScaleAndThemePreferences(t *testing.T) {
	for _, scale := range []float32{1, 1.25, 1.5, 2, 2.25} {
		for _, dark := range []bool{false, true} {
			a, tt := newTestApp(t)
			if dark {
				a.themeMode = ThemeDark
			}
			tt.SetPreferences(ui.Preferences{TextScale: scale, HighContrast: true})
			tt.Frame()
			if a.metrics.fontBody != float32(math.Round(float64(fontBodyBase*scale)*2))/2 {
				t.Fatal("text scale not applied")
			}
			if a.pal.text != a.pal.text3 || a.pal.surface != a.pal.window {
				t.Fatal("inconsistent contrast palette")
			}
			if a.metrics.controlMd < a.metrics.lineBody+2*space8 {
				t.Fatal("control cannot contain text")
			}
		}
	}
}

func TestPaletteTextContrast(t *testing.T) {
	luminance := func(c ui.Color) float64 {
		channel := func(v uint8) float64 {
			x := float64(v) / 255
			if x <= 0.04045 {
				return x / 12.92
			}
			return math.Pow((x+0.055)/1.055, 2.4)
		}
		return 0.2126*channel(c.R) + 0.7152*channel(c.G) + 0.0722*channel(c.B)
	}
	for _, p := range []palette{lightPalette(), darkPalette(), contrastPalette(false), contrastPalette(true)} {
		for _, pair := range [][2]ui.Color{{p.text, p.surface}, {p.text3, p.sidebar}, {p.danger, p.surface}, {p.dangerText, p.danger}, {p.accentText, p.accent}} {
			a, b := luminance(pair[0]), luminance(pair[1])
			ratio := (max(a, b) + 0.05) / (min(a, b) + 0.05)
			if ratio < 4.5 {
				t.Errorf("text contrast %.2f below 4.5 for %+v", ratio, pair)
			}
		}
	}
}
