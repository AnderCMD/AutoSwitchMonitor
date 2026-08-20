// Package config carga y guarda la configuración de AutoSwitchMonitor.
//
// Cada PC conectada al KVM corre su propia instancia de la app con su propio
// config.yaml: cada instancia solo sabe "cuál es mi propia entrada de video"
// y "qué dispositivo USB debo vigilar para saber cuándo el KVM me seleccionó".
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Hotkey define una combinación de teclas que fuerza al monitor a cambiar
// a una entrada específica, sin importar el estado del KVM.
type Hotkey struct {
	// Modifiers: combinación de "ctrl", "alt", "shift", "win"/"cmd".
	Modifiers []string `yaml:"modifiers"`
	// Key: tecla final, ej. "1", "k", "f13".
	Key string `yaml:"key"`
	// Target: nombre lógico de la entrada (debe existir en Inputs).
	Target string `yaml:"target"`
}

// USBWatch identifica el dispositivo USB cuya presencia indica que el KVM
// seleccionó esta PC.
type USBWatch struct {
	VendorID  string `yaml:"vendor_id"`  // ej. "046d"
	ProductID string `yaml:"product_id"` // ej. "c52b"
	// Enabled permite desactivar la autodetección (ej. en la PC que no
	// está conectada al KVM) dejando solo los hotkeys manuales.
	Enabled bool `yaml:"enabled"`
}

// Config es el archivo config.yaml completo.
type Config struct {
	// OwnInput es la entrada lógica de este PC (debe existir en Inputs).
	// Se aplica automáticamente cuando USBWatch detecta que el KVM cambió
	// a este PC.
	OwnInput string `yaml:"own_input"`

	// Inputs mapea nombres lógicos ("dp1", "hdmi1", "hdmi2") al código VCP
	// de entrada DDC/CI (VCP 0x60) que usa TU monitor. Estos códigos varían
	// por fabricante; usa `autoswitchmonitor -list-inputs` o el manual del
	// monitor para confirmarlos. Valores típicos: DisplayPort1=0x0f,
	// HDMI1=0x11, HDMI2=0x12.
	Inputs map[string]int `yaml:"inputs"`

	USBWatch USBWatch `yaml:"usb_watch"`

	// PollIntervalMS: cada cuánto se revisa la lista de dispositivos USB.
	PollIntervalMS int `yaml:"poll_interval_ms"`

	Hotkeys []Hotkey `yaml:"hotkeys"`
}

func Default() Config {
	return Config{
		OwnInput: "hdmi1",
		Inputs: map[string]int{
			"dp1":   0x0f,
			"hdmi1": 0x11,
			"hdmi2": 0x12,
		},
		USBWatch: USBWatch{
			VendorID:  "0000",
			ProductID: "0000",
			Enabled:   false,
		},
		PollIntervalMS: 400,
		Hotkeys: []Hotkey{
			{Modifiers: []string{"ctrl", "alt"}, Key: "1", Target: "dp1"},
			{Modifiers: []string{"ctrl", "alt"}, Key: "2", Target: "hdmi1"},
			{Modifiers: []string{"ctrl", "alt"}, Key: "3", Target: "hdmi2"},
		},
	}
}

// Path devuelve la ruta del config.yaml en el directorio de configuración
// del usuario del SO (%APPDATA% en Windows, ~/Library/Application Support
// en macOS).
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "AutoSwitchMonitor", "config.yaml"), nil
}

// Load lee config.yaml; si no existe, escribe y devuelve la configuración
// por defecto.
func Load() (Config, string, error) {
	path, err := Path()
	if err != nil {
		return Config{}, "", err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Default()
		if err := Save(cfg); err != nil {
			return Config{}, path, fmt.Errorf("creando config por defecto: %w", err)
		}
		return cfg, path, nil
	}
	if err != nil {
		return Config{}, path, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, path, fmt.Errorf("parseando %s: %w", path, err)
	}
	return cfg, path, nil
}

// Save escribe la configuración a disco, creando el directorio si hace falta.
func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
