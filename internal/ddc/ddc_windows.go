//go:build windows

package ddc

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32  = windows.NewLazySystemDLL("user32.dll")
	dxva2   = windows.NewLazySystemDLL("dxva2.dll")

	procEnumDisplayMonitors               = user32.NewProc("EnumDisplayMonitors")
	procGetNumberOfPhysicalMonitors        = dxva2.NewProc("GetNumberOfPhysicalMonitorsFromHMONITOR")
	procGetPhysicalMonitorsFromHMONITOR    = dxva2.NewProc("GetPhysicalMonitorsFromHMONITOR")
	procSetVCPFeature                      = dxva2.NewProc("SetVCPFeature")
	procDestroyPhysicalMonitors            = dxva2.NewProc("DestroyPhysicalMonitors")
)

// physicalMonitor mirrors the Win32 PHYSICAL_MONITOR struct.
// szPhysicalMonitorDescription is a 128 WCHAR buffer.
type physicalMonitor struct {
	Handle      windows.Handle
	Description [128]uint16
}

// setInput enumera todos los monitores físicos conectados y envía
// SetVCPFeature(0x60, vcpCode) a cada uno. Si un monitor no soporta DDC/CI
// o no está en la entrada activa, la llamada puede fallar silenciosamente
// para ese monitor (comportamiento normal de DDC/CI); seguimos con los
// demás y solo devolvemos error si ninguno lo aceptó.
func setInput(vcpCode int) error {
	hmonitors, err := enumMonitors()
	if err != nil {
		return err
	}
	if len(hmonitors) == 0 {
		return fmt.Errorf("no se encontraron monitores")
	}

	var lastErr error
	succeeded := 0

	for _, hMon := range hmonitors {
		phys, err := getPhysicalMonitors(hMon)
		if err != nil {
			lastErr = err
			continue
		}
		for _, pm := range phys {
			ok, _, callErr := procSetVCPFeature.Call(
				uintptr(pm.Handle),
				uintptr(0x60),
				uintptr(vcpCode),
			)
			if ok == 0 {
				lastErr = fmt.Errorf("SetVCPFeature falló: %v", callErr)
			} else {
				succeeded++
			}
		}
		destroyPhysicalMonitors(phys)
	}

	if succeeded == 0 && lastErr != nil {
		return lastErr
	}
	return nil
}

func enumMonitors() ([]windows.Handle, error) {
	var handles []windows.Handle

	cb := windows.NewCallback(func(hMonitor windows.Handle, _ windows.Handle, _ uintptr, _ uintptr) uintptr {
		handles = append(handles, hMonitor)
		return 1 // continuar enumeración
	})

	ret, _, callErr := procEnumDisplayMonitors.Call(0, 0, cb, 0)
	if ret == 0 {
		return nil, fmt.Errorf("EnumDisplayMonitors falló: %v", callErr)
	}
	return handles, nil
}

func getPhysicalMonitors(hMonitor windows.Handle) ([]physicalMonitor, error) {
	var count uint32
	ret, _, callErr := procGetNumberOfPhysicalMonitors.Call(
		uintptr(hMonitor),
		uintptr(unsafe.Pointer(&count)),
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetNumberOfPhysicalMonitorsFromHMONITOR falló: %v", callErr)
	}
	if count == 0 {
		return nil, nil
	}

	monitors := make([]physicalMonitor, count)
	ret, _, callErr = procGetPhysicalMonitorsFromHMONITOR.Call(
		uintptr(hMonitor),
		uintptr(count),
		uintptr(unsafe.Pointer(&monitors[0])),
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetPhysicalMonitorsFromHMONITOR falló: %v", callErr)
	}
	return monitors, nil
}

func destroyPhysicalMonitors(monitors []physicalMonitor) {
	if len(monitors) == 0 {
		return
	}
	procDestroyPhysicalMonitors.Call(
		uintptr(len(monitors)),
		uintptr(unsafe.Pointer(&monitors[0])),
	)
}
