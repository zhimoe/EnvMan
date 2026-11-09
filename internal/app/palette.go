package app

import "github.com/egoist/mygo/ui"

type palette struct {
	window      ui.Color
	sidebar     ui.Color
	surface     ui.Color
	surface2    ui.Color
	surface3    ui.Color
	text        ui.Color
	text2       ui.Color
	text3       ui.Color
	line        ui.Color
	lineSoft    ui.Color
	accent      ui.Color
	accentHover ui.Color
	accentPress ui.Color
	accentText  ui.Color
	accentSoft  ui.Color
	accentSoft2 ui.Color
	warn        ui.Color
	warnSoft    ui.Color
	danger      ui.Color
	dangerText  ui.Color
	dangerSoft  ui.Color
	success     ui.Color
	successSoft ui.Color
	hit         ui.Color
	dark        bool
}

func lightPalette() palette {
	return palette{
		window:      ui.Hex("#f5f6f8"),
		sidebar:     ui.Hex("#eff1f4"),
		surface:     ui.Hex("#ffffff"),
		surface2:    ui.Hex("#f8f9fb"),
		surface3:    ui.Hex("#f1f2f6"),
		text:        ui.Hex("#1b1c21"),
		text2:       ui.Hex("#5b5e68"),
		text3:       ui.Hex("#626570"),
		line:        ui.Hex("#e5e7ec"),
		lineSoft:    ui.Hex("#edeef2"),
		accent:      ui.Hex("#4a5cf0"),
		accentHover: ui.Hex("#3d4ddb"),
		accentPress: ui.Hex("#3342c4"),
		accentText:  ui.Hex("#ffffff"),
		accentSoft:  ui.RGBA(74, 92, 240, 0.09),
		accentSoft2: ui.RGBA(74, 92, 240, 0.16),
		warn:        ui.Hex("#a95c08"),
		warnSoft:    ui.RGBA(196, 108, 16, 0.11),
		danger:      ui.Hex("#bd2934"),
		dangerText:  ui.Hex("#ffffff"),
		dangerSoft:  ui.RGBA(220, 61, 67, 0.09),
		success:     ui.Hex("#13753e"),
		successSoft: ui.RGBA(24, 138, 74, 0.10),
		hit:         ui.RGBA(255, 198, 0, 0.28),
	}
}

func darkPalette() palette {
	return palette{
		window:      ui.Hex("#0f1014"),
		sidebar:     ui.Hex("#0c0d11"),
		surface:     ui.Hex("#17181e"),
		surface2:    ui.Hex("#1c1e25"),
		surface3:    ui.Hex("#20222b"),
		text:        ui.Hex("#eceef3"),
		text2:       ui.Hex("#a7abb7"),
		text3:       ui.Hex("#7f8492"),
		line:        ui.Hex("#242731"),
		lineSoft:    ui.Hex("#1f222a"),
		accent:      ui.Hex("#8090ff"),
		accentHover: ui.Hex("#93a1ff"),
		accentPress: ui.Hex("#6d7df0"),
		accentText:  ui.Hex("#10121a"),
		accentSoft:  ui.RGBA(128, 144, 255, 0.12),
		accentSoft2: ui.RGBA(128, 144, 255, 0.20),
		warn:        ui.Hex("#e5a44a"),
		warnSoft:    ui.RGBA(229, 164, 74, 0.12),
		danger:      ui.Hex("#ff6a70"),
		dangerText:  ui.Hex("#10121a"),
		dangerSoft:  ui.RGBA(255, 106, 112, 0.12),
		success:     ui.Hex("#4ecb85"),
		successSoft: ui.RGBA(78, 203, 133, 0.12),
		hit:         ui.RGBA(255, 210, 80, 0.22),
		dark:        true,
	}
}

// contrastPalette is a deterministic fallback: the framework exposes the
// contrast preference, but not the Windows user's custom contrast colors.
func contrastPalette(dark bool) palette {
	p := lightPalette()
	bg, fg := ui.Hex("#ffffff"), ui.Hex("#000000")
	accent := ui.Hex("#000080")
	if dark {
		p = darkPalette()
		bg, fg = fg, bg
		accent = ui.Hex("#ffff00")
	}
	p.window, p.sidebar, p.surface, p.surface2, p.surface3 = bg, bg, bg, bg, bg
	p.text, p.text2, p.text3, p.line, p.lineSoft = fg, fg, fg, fg, fg
	p.accent, p.accentHover, p.accentPress, p.accentText = accent, accent, accent, bg
	p.accentSoft, p.accentSoft2 = bg, bg
	p.warn, p.danger, p.success, p.dangerText = fg, fg, fg, bg
	p.warnSoft, p.dangerSoft, p.successSoft = bg, bg, bg
	p.hit = accent.Alpha(0.15)
	return p
}
