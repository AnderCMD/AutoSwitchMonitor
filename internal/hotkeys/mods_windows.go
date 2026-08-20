//go:build windows

package hotkeys

import "golang.design/x/hotkey"

// modifierFromName mapea nombres de config.yaml a constantes de Windows.
// "cmd" se trata como alias de "win" para que el mismo config.yaml sea
// portable entre Windows y macOS.
func modifierFromName(name string) (hotkey.Modifier, bool) {
	switch name {
	case "ctrl":
		return hotkey.ModCtrl, true
	case "shift":
		return hotkey.ModShift, true
	case "alt", "option":
		return hotkey.ModAlt, true
	case "win", "cmd":
		return hotkey.ModWin, true
	default:
		return 0, false
	}
}
