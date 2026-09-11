// Package config loads and saves AutoSwitchMonitor's configuration.
//
// Each PC connected to the KVM runs its own instance of the app with its
// own config.yaml: each instance only knows "what is my own video input"
// and "which USB device I must watch to know when the KVM selected me".
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Hotkey defines a key combination that forces the monitor to switch to a
// specific input, regardless of the KVM's state.
type Hotkey struct {
	// Modifiers: combination of "ctrl", "alt", "shift", "win"/"cmd".
	Modifiers []string `yaml:"modifiers"`
	// Key: final key, e.g. "1", "k", "f13".
	Key string `yaml:"key"`
	// Target: logical input name (must exist in Inputs).
	Target string `yaml:"target"`
}

// USBWatch identifies the USB device whose presence indicates that the KVM
// selected this PC.
type USBWatch struct {
	VendorID  string `yaml:"vendor_id"`  // e.g. "046d"
	ProductID string `yaml:"product_id"` // e.g. "c52b"
	// Enabled allows disabling autodetection (e.g. on the PC that isn't
	// connected to the KVM) leaving only the manual hotkeys.
	Enabled bool `yaml:"enabled"`
}

// Config is the full config.yaml file.
type Config struct {
	// OwnInput is this PC's logical input (must exist in Inputs).
	// It's applied automatically when USBWatch detects that the KVM
	// switched to this PC.
	OwnInput string `yaml:"own_input"`

	// Inputs maps logical names ("dp1", "hdmi1", "hdmi2") to the DDC/CI
	// input VCP code (VCP 0x60) used by YOUR monitor. These codes vary by
	// manufacturer; use `autoswitchmonitor -list-inputs` or the monitor's
	// manual to confirm them. Typical values: DisplayPort1=0x0f,
	// HDMI1=0x11, HDMI2=0x12.
	Inputs map[string]int `yaml:"inputs"`

	USBWatch USBWatch `yaml:"usb_watch"`

	// PollIntervalMS: how often the USB device list is checked.
	PollIntervalMS int `yaml:"poll_interval_ms"`

	Hotkeys []Hotkey `yaml:"hotkeys"`
}

func Default() Config {
	// "alt" and "option" are aliases for the same modifier (see
	// internal/hotkeys/mods_*.go), but showing the OS-native name in the
	// freshly created config.yaml avoids it looking like a "Windows"
	// config on macOS when it isn't.
	altKey := "alt"
	if runtime.GOOS == "darwin" {
		altKey = "option"
	}
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
			{Modifiers: []string{"ctrl", altKey}, Key: "1", Target: "dp1"},
			{Modifiers: []string{"ctrl", altKey}, Key: "2", Target: "hdmi1"},
			{Modifiers: []string{"ctrl", altKey}, Key: "3", Target: "hdmi2"},
		},
	}
}

// Path returns the path to config.yaml in the OS user config directory
// (%APPDATA% on Windows, ~/Library/Application Support on macOS).
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "AutoSwitchMonitor", "config.yaml"), nil
}

// Load reads config.yaml; if it doesn't exist, it writes and returns the
// default configuration.
func Load() (Config, string, error) {
	path, err := Path()
	if err != nil {
		return Config{}, "", err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Default()
		if err := Save(cfg); err != nil {
			return Config{}, path, fmt.Errorf("creating default config: %w", err)
		}
		return cfg, path, nil
	}
	if err != nil {
		return Config{}, path, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, path, fmt.Errorf("parsing %s: %w", path, err)
	}
	return cfg, path, nil
}

// Save writes the configuration to disk, creating the directory if needed.
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
