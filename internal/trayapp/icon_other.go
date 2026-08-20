//go:build !windows

package trayapp

import (
	"bytes"
	"image/png"

	"github.com/AnderCMD/AutoSwitchMonitor/internal/appicon"
)

// trayIconBytes en macOS/Linux debe ser un PNG (systray lo espera así en
// estas plataformas).
func trayIconBytes() []byte {
	var buf bytes.Buffer
	_ = png.Encode(&buf, appicon.Draw(64))
	return buf.Bytes()
}
