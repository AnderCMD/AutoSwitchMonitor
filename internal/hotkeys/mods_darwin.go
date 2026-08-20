//go:build darwin

package hotkeys

import "golang.design/x/hotkey"

// modifierFromName mapea nombres de config.yaml a constantes de macOS.
// "win" se trata como alias de "cmd" para que el mismo config.yaml sea
// portable entre Windows y macOS.
func modifierFromName(name string) (hotkey.Modifier, bool) {
	switch name {
	case "ctrl":
		return hotkey.ModCtrl, true
	case "shift":
		return hotkey.ModShift, true
	case "alt", "option":
		return hotkey.ModOption, true
	case "cmd", "win":
		return hotkey.ModCmd, true
	default:
		return 0, false
	}
}
