//go:build !windows

package trayapp

import (
	"bytes"
	"image"
	"image/png"

	"github.com/AnderCMD/AutoSwitchMonitor/internal/appicon"
)

// trayIconBytes on macOS/Linux must be a PNG (systray expects this on
// these platforms). This is the "regular" full-color icon, used as a
// fallback on Linux (macOS uses trayIconTemplateBytes via SetTemplateIcon).
func trayIconBytes() []byte {
	return encodePNG(appicon.Draw(64))
}

// trayIconTemplateBytes is the monochrome silhouette that macOS automatically
// recolors (white in dark mode, black in light mode) so the menu bar icon
// matches the rest of the native icons.
func trayIconTemplateBytes() []byte {
	return encodePNG(appicon.DrawTemplate(64))
}

func encodePNG(img *image.NRGBA) []byte {
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
