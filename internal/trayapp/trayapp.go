// Package trayapp implements the tray/menu icon and wires config, ddc,
// hotkeys, and usbwatch together in a single running process.
package trayapp

import (
	"log"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/getlantern/systray"

	"github.com/AnderCMD/AutoSwitchMonitor/internal/autostart"
	"github.com/AnderCMD/AutoSwitchMonitor/internal/config"
	"github.com/AnderCMD/AutoSwitchMonitor/internal/ddc"
	"github.com/AnderCMD/AutoSwitchMonitor/internal/hotkeys"
	"github.com/AnderCMD/AutoSwitchMonitor/internal/usbwatch"
)

// Run starts the tray app. Blocks until the user chooses Quit.
func Run(cfg config.Config, cfgPath string) {
	if runtime.GOOS == "darwin" {
		// golang.design/x/hotkey/mainthread already runs its own [NSApp run]
		// on the process's real main thread (it needs this for CGEventTap).
		// If we call systray.Run here (which on macOS also does [NSApp run]
		// via nativeLoop), that second loop starts from the goroutine that
		// mainthread.Init uses to wrap this function, not necessarily on the
		// main thread — and when the process is launched via LaunchServices
		// (Finder/`open`, not a direct shell) that almost always lands on a
		// different OS thread, and AppKit crashes with SIGTRAP inside
		// [NSApp run] ("nothing opens" when double-clicking the .app).
		// systray.Register leaves the real loop in charge of mainthread.Init
		// and only registers the icon/menu.
		done := make(chan struct{})
		systray.Register(func() { onReady(cfg, cfgPath) }, func() { close(done) })
		<-done
		return
	}
	systray.Run(func() { onReady(cfg, cfgPath) }, func() {})
}

func onReady(cfg config.Config, cfgPath string) {
	systray.SetTemplateIcon(trayIconTemplateBytes(), trayIconBytes())
	systray.SetTooltip("AutoSwitchMonitor — monitor input switching")

	names := make([]string, 0, len(cfg.Inputs))
	for name := range cfg.Inputs {
		names = append(names, name)
	}
	sort.Strings(names)

	systray.AddMenuItem("AutoSwitchMonitor", "").Disable()
	systray.AddSeparator()

	for _, name := range names {
		code := cfg.Inputs[name]
		label := strings.ToUpper(name)
		item := systray.AddMenuItem("Switch to "+label, "Switches the monitor to the "+label+" input")
		go func(code int) {
			for range item.ClickedCh {
				if err := ddc.SetInput(code); err != nil {
					log.Printf("error switching to input %#x: %v", code, err)
				}
			}
		}(code)
	}

	systray.AddSeparator()

	autostartEnabled, err := autostart.IsEnabled()
	if err != nil {
		log.Printf("could not read autostart state: %v", err)
	}
	mAutostart := systray.AddMenuItemCheckbox("Start with system", "Opens AutoSwitchMonitor automatically at login", autostartEnabled)
	go func() {
		for range mAutostart.ClickedCh {
			var toggleErr error
			if mAutostart.Checked() {
				toggleErr = autostart.Disable()
			} else {
				toggleErr = autostart.Enable()
			}
			if toggleErr != nil {
				log.Printf("error toggling autostart: %v", toggleErr)
				continue
			}
			if mAutostart.Checked() {
				mAutostart.Uncheck()
			} else {
				mAutostart.Check()
			}
		}
	}()

	mConfig := systray.AddMenuItem("Open config folder", "")
	mQuit := systray.AddMenuItem("Quit", "")

	stop := make(chan struct{})

	var mgr hotkeys.Manager
	var bindings []hotkeys.Binding
	for _, hk := range cfg.Hotkeys {
		code, ok := cfg.Inputs[hk.Target]
		if !ok {
			log.Printf("hotkey ignored: input %q does not exist in config.inputs", hk.Target)
			continue
		}
		codeCopy := code
		bindings = append(bindings, hotkeys.Binding{
			Modifiers: hk.Modifiers,
			Key:       hk.Key,
			OnPress: func() {
				if err := ddc.SetInput(codeCopy); err != nil {
					log.Printf("error switching to input %#x: %v", codeCopy, err)
				}
			},
		})
	}
	for _, err := range mgr.RegisterAll(bindings) {
		log.Println(err)
	}

	if cfg.USBWatch.Enabled {
		interval := time.Duration(cfg.PollIntervalMS) * time.Millisecond
		if interval <= 0 {
			interval = 400 * time.Millisecond
		}
		go func() {
			err := usbwatch.Watch(cfg.USBWatch.VendorID, cfg.USBWatch.ProductID, interval, stop, func() {
				code, ok := cfg.Inputs[cfg.OwnInput]
				if !ok {
					log.Printf("own_input %q does not exist in config.inputs", cfg.OwnInput)
					return
				}
				if err := ddc.SetInput(code); err != nil {
					log.Printf("error switching to own_input: %v", err)
				}
			})
			if err != nil {
				log.Printf("usbwatch stopped: %v", err)
			}
		}()
	}

	go func() {
		for range mConfig.ClickedCh {
			openInFileManager(cfgPath)
		}
	}()

	go func() {
		<-mQuit.ClickedCh
		close(stop)
		mgr.UnregisterAll()
		systray.Quit()
	}()
}

func openInFileManager(path string) {
	dir := path
	if idx := lastSlash(path); idx != -1 {
		dir = path[:idx]
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", dir)
	case "darwin":
		cmd = exec.Command("open", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("could not open %s: %v", dir, err)
	}
}

func lastSlash(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '\\' || s[i] == '/' {
			return i
		}
	}
	return -1
}
