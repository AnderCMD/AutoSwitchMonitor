//go:build darwin

package usbwatch

import (
	"encoding/json"
	"os/exec"
	"strings"
)

type spItem struct {
	Name      string   `json:"_name"`
	VendorID  string   `json:"vendor_id"`
	ProductID string   `json:"product_id"`
	Items     []spItem `json:"_items"`
}

type spRoot struct {
	SPUSBDataType []spItem `json:"SPUSBDataType"`
}

// list usa la utilidad de sistema `system_profiler` (incluida en todo
// macOS, sin dependencias adicionales) para enumerar dispositivos USB.
func list() ([]Device, error) {
	out, err := exec.Command("system_profiler", "SPUSBDataType", "-json").Output()
	if err != nil {
		return nil, err
	}

	var root spRoot
	if err := json.Unmarshal(out, &root); err != nil {
		return nil, err
	}

	var devices []Device
	var walk func(items []spItem)
	walk = func(items []spItem) {
		for _, it := range items {
			if vid, pid, ok := extractIDs(it); ok {
				devices = append(devices, Device{VendorID: vid, ProductID: pid, Name: it.Name})
			}
			if len(it.Items) > 0 {
				walk(it.Items)
			}
		}
	}
	walk(root.SPUSBDataType)

	return devices, nil
}

// extractIDs normaliza campos como "0x046d  (Logitech Inc.)" a "046d".
func extractIDs(it spItem) (vid, pid string, ok bool) {
	vid = normalizeHexID(it.VendorID)
	pid = normalizeHexID(it.ProductID)
	if vid == "" || pid == "" {
		return "", "", false
	}
	return vid, pid, true
}

func normalizeHexID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	field := strings.Fields(raw)[0]
	field = strings.TrimPrefix(strings.ToLower(field), "0x")
	return field
}
