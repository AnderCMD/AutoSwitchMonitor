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

	"github.com/AnderCMD/AutoSwitchMonitor/internal/config"
	"github.com/AnderCMD/AutoSwitchMonitor/internal/ddc"
	"github.com/AnderCMD/AutoSwitchMonitor/internal/hotkeys"
	"github.com/AnderCMD/AutoSwitchMonitor/internal/settingsui"
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
	mSettings := systray.AddMenuItem("Configurar atajos de teclado...", "")
	mConfig := systray.AddMenuItem("Abrir carpeta de configuración", "")
	mQuit := systray.AddMenuItem("Salir", "")

	stop := make(chan struct{})

	var mgr hotkeys.Manager
	registerHotkeys(&mgr, cfg)

	go func() {
		for range mSettings.ClickedCh {
			settingsui.Open(cfg, func(newCfg config.Config) {
				cfg.Hotkeys = newCfg.Hotkeys
				mgr.UnregisterAll()
				registerHotkeys(&mgr, cfg)
			})
		}
	}()

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

// registerHotkeys registra en mgr los hotkeys definidos en cfg.Hotkeys.
// Se usa tanto al arrancar como después de guardar cambios desde
// settingsui (previo mgr.UnregisterAll()).
func registerHotkeys(mgr *hotkeys.Manager, cfg config.Config) {
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
