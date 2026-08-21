// Package autostart activa/desactiva que la app arranque sola al iniciar
// sesión en el sistema. La implementación es específica de cada SO (ver
// autostart_windows.go y autostart_darwin.go) pero ambas son solo
// Go estándar + APIs del SO: sin cgo, sin dependencias nuevas.
package autostart

// IsEnabled indica si el autoarranque ya está configurado para el
// ejecutable actual.
func IsEnabled() (bool, error) {
	return isEnabled()
}

// Enable configura el sistema para lanzar el ejecutable actual al iniciar
// sesión.
func Enable() error {
	return enable()
}

// Disable quita esa configuración.
func Disable() error {
	return disable()
}
