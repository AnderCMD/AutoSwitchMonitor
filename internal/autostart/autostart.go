// Package autostart enables/disables the app to launch on its own when the
// system session starts. The implementation is OS-specific (see
// autostart_windows.go and autostart_darwin.go) but both use only
// standard Go + OS APIs: no cgo, no new dependencies.
package autostart

// IsEnabled reports whether autostart is already configured for the
// current executable.
func IsEnabled() (bool, error) {
	return isEnabled()
}

// Enable configures the system to launch the current executable when the
// session starts.
func Enable() error {
	return enable()
}

// Disable removes that configuration.
func Disable() error {
	return disable()
}
