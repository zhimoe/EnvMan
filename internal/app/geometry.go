package app

// The corner radii, after Windows' Fluent system: controls take
// ControlCornerRadius, overlays OverlayCornerRadius, and pills and badges are
// round.
const (
	radiusControl float32 = 4 // Windows' ControlCornerRadius: buttons, inputs, rows, nav items
	radiusOverlay float32 = 8 // Windows' OverlayCornerRadius: cards, banners, menus, dialogs, windows
	radiusPill    float32 = 99
)

// tileRadius rounds the corners of a square icon tile of side s: as a control
// below 32 DIPs, as an overlay above, keeping it a rounded square.
func tileRadius(s float32) float32 {
	if s < 32 {
		return radiusControl
	}
	return radiusOverlay
}
