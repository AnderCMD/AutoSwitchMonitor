// Package usbwatch detecta, por sondeo (polling), cuándo un dispositivo USB
// específico (identificado por vendor_id:product_id) aparece o desaparece
// del sistema. Se usa para saber cuándo el switch KVM seleccionó esta PC:
// cuando el dispositivo vigilado (ej. el teclado/mouse que pasa por el KVM)
// pasa de ausente a presente, significa que el KVM apunta a esta PC.
package usbwatch

import (
	"strings"
	"time"
)

// Device es un dispositivo USB detectado en el sistema.
type Device struct {
	VendorID  string // hex minúsculas, sin "0x", ej "046d"
	ProductID string // hex minúsculas, sin "0x", ej "c52b"
	Name      string
}

// List devuelve los dispositivos USB actualmente conectados.
// Implementación específica de SO (ver usbwatch_windows.go / usbwatch_darwin.go).
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

// Watch vigila la presencia de vendorID:productID cada interval, y llama a
// onBecamePresent cada vez que el dispositivo pasa de ausente a presente
// (flanco de subida). Bloquea hasta que stop se cierra.
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
				// Error transitorio de enumeración (permisos, driver, etc.):
				// lo ignoramos y reintentamos en el siguiente tick.
				continue
			}
			if present && !wasPresent {
				onBecamePresent()
			}
			wasPresent = present
		}
	}
}
