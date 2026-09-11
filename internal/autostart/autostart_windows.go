//go:build windows

package autostart

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

// We use HKCU's "Run" key (per-user, no administrator permissions needed)
// instead of a scheduled task or a shortcut in shell:startup: it's a single
// registry call, with no files created.
const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

const valueName = "AutoSwitchMonitor"

func isEnabled() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false, nil // the "Run" key doesn't exist yet: not enabled
	}
	defer k.Close()

	_, _, err = k.GetStringValue(valueName)
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func enable() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	return k.SetStringValue(valueName, `"`+exe+`"`)
}

func disable() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return nil // without the key, it's already "disabled"
	}
	defer k.Close()

	err = k.DeleteValue(valueName)
	if err != nil && err != registry.ErrNotExist {
		return err
	}
	return nil
}
