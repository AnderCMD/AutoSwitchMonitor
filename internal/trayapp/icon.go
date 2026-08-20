//go:build windows

package trayapp

import "github.com/AnderCMD/AutoSwitchMonitor/internal/appicon"

// trayIconBytes en Windows debe ser un .ico.
func trayIconBytes() []byte {
	return appicon.EncodeICO(appicon.Draw(32))
}
