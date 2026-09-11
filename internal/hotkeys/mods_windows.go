//go:build windows

package hotkeys

import "golang.design/x/hotkey"

// modifierFromName maps config.yaml names to Windows constants.
// "cmd" is treated as an alias for "win" so the same config.yaml is
// portable between Windows and macOS.
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
