//go:build windows

package trayapp

import "github.com/AnderCMD/AutoSwitchMonitor/internal/appicon"

// trayIconBytes on Windows must be an .ico.
func trayIconBytes() []byte {
	return appicon.EncodeICO(appicon.Draw(32))
}

// trayIconTemplateBytes has no visible effect on Windows (systray uses the
// regular icon there), but must still return a valid .ico: systray.SetTemplateIcon
// falls back to the regular icon on this platform.
func trayIconTemplateBytes() []byte {
	return trayIconBytes()
}
