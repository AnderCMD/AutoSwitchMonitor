// Package settingsui muestra una ventana nativa (webview, sin dependencias
// pesadas: usa WebView2 en Windows y WKWebView en macOS) para editar los
// hotkeys de config.yaml sin tocar el archivo a mano — incluye captura en
// vivo de la combinación de teclas.
package settingsui

import (
	"encoding/json"
	"fmt"
	"sort"

	webview "github.com/webview/webview_go"
	"golang.design/x/mainthread"

	"github.com/AnderCMD/AutoSwitchMonitor/internal/config"
)

// Open muestra la ventana y bloquea hasta que el usuario la cierra
// (Guardar o Cancelar). Si guardó cambios, onSave se llama con la
// configuración actualizada (ya persistida en disco).
//
// cfg se recibe y se modifica por valor: no comparte estado con quien
// llama, así que es seguro invocar Open desde una gorutina distinta a la
// que sigue usando la config original (ej. el watcher de USB).
func Open(cfg config.Config, onSave func(config.Config)) {
	mainthread.Call(func() {
		runWindow(cfg, onSave)
	})
}

func runWindow(cfg config.Config, onSave func(config.Config)) {
	w := webview.New(false)
	defer w.Destroy()

	w.SetTitle("AutoSwitchMonitor — Atajos de teclado")
	w.SetSize(560, 460, webview.HintNone)

	saved := false

	_ = w.Bind("getState", func() map[string]any {
		return map[string]any{
			"hotkeys": cfg.Hotkeys,
			"inputs":  inputNames(cfg),
		}
	})

	_ = w.Bind("saveHotkeys", func(raw string) (string, error) {
		var hks []config.Hotkey
		if err := json.Unmarshal([]byte(raw), &hks); err != nil {
			return "", fmt.Errorf("formato inválido: %w", err)
		}
		for _, hk := range hks {
			if _, ok := cfg.Inputs[hk.Target]; !ok {
				return "", fmt.Errorf("entrada desconocida: %q", hk.Target)
			}
			if len(hk.Modifiers) == 0 || hk.Key == "" {
				return "", fmt.Errorf("hay un atajo sin combinación de teclas")
			}
		}
		cfg.Hotkeys = hks
		if err := config.Save(cfg); err != nil {
			return "", fmt.Errorf("guardando config.yaml: %w", err)
		}
		saved = true
		return "ok", nil
	})

	_ = w.Bind("closeWindow", func() {
		w.Terminate()
	})

	w.SetHtml(pageHTML)
	w.Run()

	if saved && onSave != nil {
		onSave(cfg)
	}
}

func inputNames(cfg config.Config) []string {
	names := make([]string, 0, len(cfg.Inputs))
	for name := range cfg.Inputs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
