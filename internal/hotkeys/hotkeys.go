// Package hotkeys registra combinaciones de teclas globales (funcionan aun
// sin foco en ninguna ventana) usando golang.design/x/hotkey.
package hotkeys

import (
	"fmt"
	"strings"

	"golang.design/x/hotkey"
)

// Binding es una combinación de teclas y la acción a ejecutar al presionarla.
type Binding struct {
	Modifiers []string
	Key       string
	OnPress   func()
}

// Manager mantiene vivas las hotkeys registradas mientras la app corre.
type Manager struct {
	active []*hotkey.Hotkey
}

// RegisterAll registra todas las combinaciones. Si una falla en registrarse
// (ej. porque otra app ya la usa), se reporta el error pero se sigue con
// las demás.
func (m *Manager) RegisterAll(bindings []Binding) []error {
	var errs []error
	for _, b := range bindings {
		if err := m.register(b); err != nil {
			errs = append(errs, fmt.Errorf("hotkey %s+%s: %w", strings.Join(b.Modifiers, "+"), b.Key, err))
		}
	}
	return errs
}

func (m *Manager) register(b Binding) error {
	mods := make([]hotkey.Modifier, 0, len(b.Modifiers))
	for _, name := range b.Modifiers {
		mod, ok := modifierFromName(name)
		if !ok {
			return fmt.Errorf("modificador desconocido %q", name)
		}
		mods = append(mods, mod)
	}

	key, ok := keyFromName(b.Key)
	if !ok {
		return fmt.Errorf("tecla desconocida %q", b.Key)
	}

	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		return err
	}

	m.active = append(m.active, hk)

	go func() {
		for range hk.Keydown() {
			b.OnPress()
		}
	}()

	return nil
}

// UnregisterAll libera todas las hotkeys registradas.
func (m *Manager) UnregisterAll() {
	for _, hk := range m.active {
		hk.Unregister()
	}
	m.active = nil
}

var keyNames = map[string]hotkey.Key{
	"0": hotkey.Key0, "1": hotkey.Key1, "2": hotkey.Key2, "3": hotkey.Key3,
	"4": hotkey.Key4, "5": hotkey.Key5, "6": hotkey.Key6, "7": hotkey.Key7,
	"8": hotkey.Key8, "9": hotkey.Key9,
	"a": hotkey.KeyA, "b": hotkey.KeyB, "c": hotkey.KeyC, "d": hotkey.KeyD,
	"e": hotkey.KeyE, "f": hotkey.KeyF, "g": hotkey.KeyG, "h": hotkey.KeyH,
	"i": hotkey.KeyI, "j": hotkey.KeyJ, "k": hotkey.KeyK, "l": hotkey.KeyL,
	"m": hotkey.KeyM, "n": hotkey.KeyN, "o": hotkey.KeyO, "p": hotkey.KeyP,
	"q": hotkey.KeyQ, "r": hotkey.KeyR, "s": hotkey.KeyS, "t": hotkey.KeyT,
	"u": hotkey.KeyU, "v": hotkey.KeyV, "w": hotkey.KeyW, "x": hotkey.KeyX,
	"y": hotkey.KeyY, "z": hotkey.KeyZ,
}

func keyFromName(name string) (hotkey.Key, bool) {
	k, ok := keyNames[strings.ToLower(strings.TrimSpace(name))]
	return k, ok
}
