//go:build darwin

package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// A per-user LaunchAgent (~/Library/LaunchAgents) is macOS's equivalent of
// Windows' "Run" key: it doesn't need administrator privileges, and macOS
// checks it automatically at login.
const label = "dev.andercmd.autoswitchmonitor"

func plistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist"), nil
}

func isEnabled() (bool, error) {
	path, err := plistPath()
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func enable() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	path, err := plistPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`, label, exe)

	if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
		return err
	}

	// launchctl may fail if it was already loaded with old content; that's
	// not fatal, since macOS re-reads it at the next login anyway.
	_ = exec.Command("launchctl", "load", path).Run()
	return nil
}

func disable() error {
	path, err := plistPath()
	if err != nil {
		return err
	}
	if _, statErr := os.Stat(path); statErr == nil {
		_ = exec.Command("launchctl", "unload", path).Run()
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
