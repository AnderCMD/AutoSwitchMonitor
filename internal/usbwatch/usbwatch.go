// Package usbwatch detects, by polling, when a specific USB device
// (identified by vendor_id:product_id) appears or disappears from the
// system. It's used to know when the KVM switch has selected this PC: when
// the watched device (e.g. the keyboard/mouse that goes through the KVM)
// goes from absent to present, it means the KVM is pointing at this PC.
package usbwatch

import (
	"strings"
	"time"
)

// Device is a USB device detected on the system.
type Device struct {
	VendorID  string // lowercase hex, no "0x", e.g. "046d"
	ProductID string // lowercase hex, no "0x", e.g. "c52b"
	Name      string
}

// List returns the USB devices currently connected.
// OS-specific implementation (see usbwatch_windows.go / usbwatch_darwin.go).
func List() ([]Device, error) {
	return list()
}

func matches(d Device, vendorID, productID string) bool {
	return strings.EqualFold(d.VendorID, vendorID) && strings.EqualFold(d.ProductID, productID)
}

func isPresent(vendorID, productID string) (bool, error) {
	devices, err := List()
	if err != nil {
		return false, err
	}
	for _, d := range devices {
		if matches(d, vendorID, productID) {
			return true, nil
		}
	}
	return false, nil
}

// Watch watches for the presence of vendorID:productID every interval, and
// calls onBecamePresent each time the device goes from absent to present
// (rising edge). Blocks until stop is closed.
func Watch(vendorID, productID string, interval time.Duration, stop <-chan struct{}, onBecamePresent func()) error {
	wasPresent, err := isPresent(vendorID, productID)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			present, err := isPresent(vendorID, productID)
			if err != nil {
				// Transient enumeration error (permissions, driver, etc.):
				// ignore it and retry on the next tick.
				continue
			}
			if present && !wasPresent {
				onBecamePresent()
			}
			wasPresent = present
		}
	}
}
