//go:build !windows

package trayapp

import (
	"bytes"
	"image"
	"image/png"

	"github.com/AnderCMD/AutoSwitchMonitor/internal/appicon"
)

// trayIconBytes en macOS/Linux debe ser un PNG (systray lo espera así en
// estas plataformas). Es el ícono "regular" a color, usado como respaldo en
// Linux (macOS usa trayIconTemplateBytes vía SetTemplateIcon).
func trayIconBytes() []byte {
	return encodePNG(appicon.Draw(64))
}

// trayIconTemplateBytes es la silueta monocroma que macOS recolorea
// automáticamente (blanco en modo oscuro, negro en modo claro) para que el
// ícono de la barra de menú concuerde con el resto de los íconos nativos.
func trayIconTemplateBytes() []byte {
	return encodePNG(appicon.DrawTemplate(64))
}

func encodePNG(img *image.NRGBA) []byte {
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
