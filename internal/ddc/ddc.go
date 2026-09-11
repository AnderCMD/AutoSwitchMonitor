// Package ddc changes the video input of an external monitor using
// DDC/CI (VCP 0x60 = "Input Source").
package ddc

// SetInput changes the monitor's active input to the given VCP code.
// The implementation is OS-specific (see ddc_windows.go and
// ddc_darwin.go).
func SetInput(vcpCode int) error {
	return setInput(vcpCode)
}
