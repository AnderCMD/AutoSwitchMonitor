// Package trayapp implementa el ícono de bandeja/menú y conecta config,
// ddc, hotkeys y usbwatch en un solo proceso en ejecución.
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

// Run arranca la app de bandeja. Bloquea hasta que el usuario elige Salir.
func Run(cfg config.Config, cfgPath string) {
	systray.Run(func() { onReady(cfg, cfgPath) }, func() {})
}

func onReady(cfg config.Config, cfgPath string) {
	systray.SetIcon(trayIconBytes())
	systray.SetTooltip("AutoSwitchMonitor — cambio de entrada del monitor")

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
		item := systray.AddMenuItem("Cambiar a "+label, "Cambia el monitor a la entrada "+label)
		go func(code int) {
			for range item.ClickedCh {
				if err := ddc.SetInput(code); err != nil {
					log.Printf("error cambiando a entrada %#x: %v", code, err)
				}
			}
		}(code)
	}

	systray.AddSeparator()

	autostartEnabled, err := autostart.IsEnabled()
	if err != nil {
		log.Printf("no se pudo leer el estado de autoarranque: %v", err)
	}
	mAutostart := systray.AddMenuItemCheckbox("Iniciar con el sistema", "Abre AutoSwitchMonitor automáticamente al iniciar sesión", autostartEnabled)
	go func() {
		for range mAutostart.ClickedCh {
			var toggleErr error
			if mAutostart.Checked() {
				toggleErr = autostart.Disable()
			} else {
				toggleErr = autostart.Enable()
			}
			if toggleErr != nil {
				log.Printf("error cambiando autoarranque: %v", toggleErr)
				continue
			}
			if mAutostart.Checked() {
				mAutostart.Uncheck()
			} else {
				mAutostart.Check()
			}
		}
	}()

	mConfig := systray.AddMenuItem("Abrir carpeta de configuración", "")
	mQuit := systray.AddMenuItem("Salir", "")

	stop := make(chan struct{})

	var mgr hotkeys.Manager
	var bindings []hotkeys.Binding
	for _, hk := range cfg.Hotkeys {
		code, ok := cfg.Inputs[hk.Target]
		if !ok {
			log.Printf("hotkey ignorada: entrada %q no existe en config.inputs", hk.Target)
			continue
		}
		codeCopy := code
		bindings = append(bindings, hotkeys.Binding{
			Modifiers: hk.Modifiers,
			Key:       hk.Key,
			OnPress: func() {
				if err := ddc.SetInput(codeCopy); err != nil {
					log.Printf("error cambiando a entrada %#x: %v", codeCopy, err)
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
					log.Printf("own_input %q no existe en config.inputs", cfg.OwnInput)
					return
				}
				if err := ddc.SetInput(code); err != nil {
					log.Printf("error cambiando a own_input: %v", err)
				}
			})
			if err != nil {
				log.Printf("usbwatch detenido: %v", err)
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
		log.Printf("no se pudo abrir %s: %v", dir, err)
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
