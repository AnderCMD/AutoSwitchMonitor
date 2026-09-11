// Package hotkeys registers global key combinations (they work even
// without focus on any window) using golang.design/x/hotkey.
package hotkeys

import (
	"fmt"
	"strings"

	"golang.design/x/hotkey"
)

// Binding is a key combination and the action to run when it's pressed.
type Binding struct {
	Modifiers []string
	Key       string
	OnPress   func()
}

// Manager keeps the registered hotkeys alive while the app runs.
type Manager struct {
	active []*hotkey.Hotkey
}

// RegisterAll registers all the combinations. If one fails to register
// (e.g. because another app already uses it), the error is reported but
// the rest continue.
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
			return fmt.Errorf("unknown modifier %q", name)
		}
		mods = append(mods, mod)
	}

	key, ok := keyFromName(b.Key)
	if !ok {
		return fmt.Errorf("unknown key %q", b.Key)
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

// UnregisterAll releases all registered hotkeys.
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
