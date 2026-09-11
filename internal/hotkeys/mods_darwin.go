//go:build darwin

package hotkeys

import "golang.design/x/hotkey"

// modifierFromName maps config.yaml names to macOS constants.
// "win" is treated as an alias for "cmd" so the same config.yaml is
// portable between Windows and macOS.
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
