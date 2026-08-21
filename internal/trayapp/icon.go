//go:build windows

package trayapp

import "github.com/AnderCMD/AutoSwitchMonitor/internal/appicon"

// trayIconBytes en Windows debe ser un .ico.
func trayIconBytes() []byte {
	return appicon.EncodeICO(appicon.Draw(32))
}

// trayIconTemplateBytes no tiene efecto visible en Windows (systray usa el
// ícono regular ahí), pero debe devolver un .ico válido: systray.SetTemplateIcon
// cae de vuelta al ícono regular en esta plataforma.
func trayIconTemplateBytes() []byte {
	return trayIconBytes()
}
