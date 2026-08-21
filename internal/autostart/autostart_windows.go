//go:build windows

package autostart

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

// Usamos el "Run" key de HKCU (por-usuario, no necesita permisos de
// administrador) en vez de una tarea programada o un acceso directo en
// shell:startup: es una sola llamada de registro, sin crear archivos.
const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`

const valueName = "AutoSwitchMonitor"

func isEnabled() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false, nil // la clave "Run" no existe todavía: no está habilitado
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
		return nil // sin la clave, ya está "deshabilitado"
	}
	defer k.Close()

	err = k.DeleteValue(valueName)
	if err != nil && err != registry.ErrNotExist {
		return err
	}
	return nil
}
