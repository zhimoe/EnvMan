package app

import "math"

const monoFont = "SF Mono, Consolas, monospace"

// The type scale: five steps cover every text the app shows, in DIPs at the
// system's text scale of 1.
const (
	fontCaptionBase  float32 = 10 // section labels, footers, in-row badges
	fontSmallBase    float32 = 11 // hints, callouts, secondary lines
	fontBodyBase     float32 = 12 // default: row values, buttons, nav items
	fontSubtitleBase float32 = 14 // sheet, dialog and empty-state titles
	fontTitleBase    float32 = 18 // page title
)

// The weights, as Windows uses them: semibold marks what the eye should catch
// first, medium the rest.
const (
	weightMedium   = 500
	weightSemibold = 600
)

// The spacing scale, a 4 DIP grid with a 2 DIP half step: every gap and
// container padding is one of these, so that spacing stays in rhythm across
// the app. Windows' own ramp is 4/8/12/16/24/32.
const (
	space2  float32 = 2
	space4  float32 = 4
	space6  float32 = 6
	space8  float32 = 8
	space12 float32 = 12
	space16 float32 = 16
	space24 float32 = 24
	space32 float32 = 32
)

// The line box of Windows' Segoe UI, as a multiple of the font size: its
// ascent, descent and line gap.
const lineFactor = 4.0 / 3.0

// metrics are the sizes the view lays out with, recomputed each frame from
// the system's text scale: the font sizes scale, and the heights that hold
// text follow their line boxes, to reduce clipping as text grows. Font fallback and wrapping still require layout QA.
type metrics struct {
	fontCaption  float32
	fontSmall    float32
	fontBody     float32
	fontSubtitle float32
	fontTitle    float32

	// lineBody is the line box the body text takes, which the heights below
	// grow from and a fixed line height matches.
	lineBody float32

	// The heights, each a line box and its padding, on the 4 DIP grid.
	controlXs    float32 // inline buttons and icon buttons
	controlSm    float32 // quiet buttons, fields inside rows
	controlMd    float32 // primary and secondary buttons, fields
	rowH         float32 // table rows
	barH         float32 // table header and footer, banners
	titleBarH    float32 // the window's title bar, under the window controls
	paneHeaderH  float32 // a pane's header
	pageHeaderH  float32 // a page's or a sheet's header
	sheetFooterH float32
}

// newMetrics returns the sizes for a system text scale, which is 1 or more.
func newMetrics(textScale float32) metrics {
	scale := textScale
	if scale < 1 || math.IsNaN(float64(scale)) || math.IsInf(float64(scale), 0) {
		scale = 1
	}
	// Font sizes land on half DIPs, as the system's text scale does.
	font := func(base float32) float32 {
		return float32(math.Round(float64(base)*float64(scale)*2)) / 2
	}
	m := metrics{
		fontCaption:  font(fontCaptionBase),
		fontSmall:    font(fontSmallBase),
		fontBody:     font(fontBodyBase),
		fontSubtitle: font(fontSubtitleBase),
		fontTitle:    font(fontTitleBase),
	}
	line := func(size float32) float32 {
		return float32(math.Ceil(float64(size) * lineFactor))
	}
	body := line(m.fontBody)
	m.lineBody = body
	m.controlXs = body + space8             // 24 at scale 1
	m.controlSm = body + space12            // 28
	m.controlMd = body + 2*space8           // 32
	m.rowH = m.controlMd                    // 32
	m.barH = m.controlMd                    // 32
	m.titleBarH = m.controlMd + 2*space4    // 40
	m.paneHeaderH = m.controlXs + 2*space8  // 40
	m.pageHeaderH = m.controlMd + 2*space8  // 48
	m.sheetFooterH = m.controlMd + 3*space8 // 56
	return m
}
