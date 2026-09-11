// Command autoswitchmonitor runs in the background (system tray) and:
//
//  1. Watches a specific USB device (the one that moves with the KVM
//     switch) and, when it appears, automatically switches the monitor's
//     video input to the one that corresponds to this PC (DDC/CI).
//  2. Registers configurable global hotkeys to force the input switch
//     manually (useful for the PC that isn't connected to the KVM, or to
//     jump to any input at any time).
//
// Each PC runs its own instance with its own config.yaml (see
// internal/config). Use `-scan` to identify the vendor_id/product_id of
// the USB device you need to watch.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"golang.design/x/hotkey/mainthread"

	"github.com/AnderCMD/AutoSwitchMonitor/internal/config"
	"github.com/AnderCMD/AutoSwitchMonitor/internal/trayapp"
	"github.com/AnderCMD/AutoSwitchMonitor/internal/usbwatch"
)

func main() {
	scan := flag.Bool("scan", false, "Lists connected USB devices every second, to identify the vendor_id/product_id of the device that changes with the KVM (turn the KVM on/off between PCs while this runs and compare the list)")
	printConfigPath := flag.Bool("config-path", false, "Prints the path to config.yaml and exits")
	flag.Parse()

	if *printConfigPath {
		path, err := config.Path()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(path)
		return
	}

	if *scan {
		runScan()
		return
	}

	cfg, path, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}
	fmt.Fprintf(os.Stderr, "AutoSwitchMonitor using config: %s\n", path)

	// golang.design/x/hotkey needs the registration and event loop to run
	// on the OS main thread (critical on macOS because of CGEventTap).
	mainthread.Init(func() {
		trayapp.Run(cfg, path)
	})
}

func runScan() {
	fmt.Println("Scanning USB devices. Switch the KVM between your PCs and watch which")
	fmt.Println("vendor_id:product_id appears/disappears. Ctrl+C to quit.")
	fmt.Println()

	prev := map[string]usbwatch.Device{}
	for {
		devices, err := usbwatch.List()
		if err != nil {
			log.Fatalf("listing USB devices: %v", err)
		}

		curr := map[string]usbwatch.Device{}
		for _, d := range devices {
			key := d.VendorID + ":" + d.ProductID
			curr[key] = d
		}

		for key, d := range curr {
			if _, existed := prev[key]; !existed {
				fmt.Printf("+ CONNECTED    %-4s:%-4s  %s\n", d.VendorID, d.ProductID, d.Name)
			}
		}
		for key, d := range prev {
			if _, still := curr[key]; !still {
				fmt.Printf("- DISCONNECTED %-4s:%-4s  %s\n", d.VendorID, d.ProductID, d.Name)
			}
		}

		prev = curr
		time.Sleep(1 * time.Second)
	}
}
