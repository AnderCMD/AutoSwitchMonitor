//go:build windows

package usbwatch

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	digcfPresent    = 0x00000002
	digcfAllClasses = 0x00000004
	errorNoMoreItems syscall.Errno = 259
)

var (
	setupapi = windows.NewLazySystemDLL("setupapi.dll")

	procSetupDiGetClassDevsW        = setupapi.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInfo       = setupapi.NewProc("SetupDiEnumDeviceInfo")
	procSetupDiGetDeviceInstanceIdW = setupapi.NewProc("SetupDiGetDeviceInstanceIdW")
	procSetupDiDestroyDeviceInfoList = setupapi.NewProc("SetupDiDestroyDeviceInfoList")
)

type spDevinfoData struct {
	cbSize    uint32
	ClassGUID windows.GUID
	DevInst   uint32
	Reserved  uintptr
}

// list enumera todos los dispositivos actualmente presentes bajo el árbol
// "USB" usando SetupAPI (no requiere libusb ni cgo).
func list() ([]Device, error) {
	enumerator, err := syscall.UTF16PtrFromString("USB")
	if err != nil {
		return nil, err
	}

	hDevInfo, _, callErr := procSetupDiGetClassDevsW.Call(
		0,
		uintptr(unsafe.Pointer(enumerator)),
		0,
		uintptr(digcfPresent|digcfAllClasses),
	)
	if hDevInfo == uintptr(windows.InvalidHandle) {
		return nil, callErr
	}
	defer procSetupDiDestroyDeviceInfoList.Call(hDevInfo)

	var devices []Device
	var index uint32

	for {
		data := spDevinfoData{}
		data.cbSize = uint32(unsafe.Sizeof(data))

		ret, _, callErr := procSetupDiEnumDeviceInfo.Call(
			hDevInfo,
			uintptr(index),
			uintptr(unsafe.Pointer(&data)),
		)
		index++
		if ret == 0 {
			if callErr == errorNoMoreItems {
				break
			}
			break
		}

		buf := make([]uint16, 512)
		var required uint32
		ret, _, _ = procSetupDiGetDeviceInstanceIdW.Call(
			hDevInfo,
			uintptr(unsafe.Pointer(&data)),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)),
			uintptr(unsafe.Pointer(&required)),
		)
		if ret == 0 {
			continue
		}

		instanceID := syscall.UTF16ToString(buf)
		vid, pid, ok := parseInstanceID(instanceID)
		if !ok {
			continue
		}
		devices = append(devices, Device{
			VendorID:  vid,
			ProductID: pid,
			Name:      instanceID,
		})
	}

	return devices, nil
}

// parseInstanceID extrae VID/PID de un instance ID tipo:
// "USB\VID_046D&PID_C52B\6&2f1c3a0&0&1"
func parseInstanceID(id string) (vid, pid string, ok bool) {
	upper := strings.ToUpper(id)
	vidIdx := strings.Index(upper, "VID_")
	pidIdx := strings.Index(upper, "PID_")
	if vidIdx == -1 || pidIdx == -1 {
		return "", "", false
	}
	vid = safeSlice(upper, vidIdx+4, vidIdx+8)
	pid = safeSlice(upper, pidIdx+4, pidIdx+8)
	if len(vid) != 4 || len(pid) != 4 {
		return "", "", false
	}
	return strings.ToLower(vid), strings.ToLower(pid), true
}

func safeSlice(s string, start, end int) string {
	if start < 0 || end > len(s) || start > end {
		return ""
	}
	return s[start:end]
}
