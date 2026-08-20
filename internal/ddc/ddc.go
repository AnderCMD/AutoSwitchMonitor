// Package ddc cambia la entrada de video de un monitor externo usando
// DDC/CI (VCP 0x60 = "Input Source").
package ddc

// SetInput cambia la entrada activa del monitor al código VCP indicado.
// La implementación es específica de cada sistema operativo (ver
// ddc_windows.go y ddc_darwin.go).
func SetInput(vcpCode int) error {
	return setInput(vcpCode)
}
