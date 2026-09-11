// Command render-icon-master rasterizes assets/icon.svg (the source
// design: glass ring, monitor, and "ON" toggle) into a high-resolution
// master PNG, using headless Chrome/Chromium —the drawing itself uses
// <text> and filters (feGaussianBlur, feDropShadow) that Go's pure SVG
// rasterizers don't support, so this step needs a real render engine—.
//
// Only needs to be run when the logo design changes (assets/icon.svg):
// it writes internal/appicon/icon_master.png, which gen-icon (100% Go, no
// external dependencies) then rescales to every size for
// assets/icon.png/.ico/.icns and the tray icon at runtime. Chrome isn't
// needed to build or run the app.
//
// Usage: go run ./tools/render-icon-master
package main

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const masterSize = 1024

func main() {
	root, err := repoRoot()
	if err != nil {
		log.Fatal(err)
	}

	svgPath := filepath.Join(root, "assets", "icon.svg")
	if _, err := os.Stat(svgPath); err != nil {
		log.Fatalf("could not find %s: %v", svgPath, err)
	}

	chrome, err := findChrome()
	if err != nil {
		log.Fatal(err)
	}

	outPath := filepath.Join(root, "internal", "appicon", "icon_master.png")

	args := []string{
		"--headless",
		"--disable-gpu",
		"--default-background-color=00000000",
		fmt.Sprintf("--screenshot=%s", outPath),
		fmt.Sprintf("--window-size=%d,%d", masterSize, masterSize),
		"file://" + svgPath,
	}
	cmd := exec.Command(chrome, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("headless chrome failed: %v\n%s", err, out)
	}

	if err := verifySquarePNG(outPath, masterSize); err != nil {
		log.Fatal(err)
	}

	log.Println("generated:", filepath.Join("internal", "appicon", "icon_master.png"))
	log.Println("run 'go run ./tools/gen-icon' to regenerate assets/icon.png/.ico/.icns")
}

func findChrome() (string, error) {
	candidates := map[string][]string{
		"darwin": {
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		},
		"linux": {"google-chrome", "chromium", "chromium-browser"},
		"windows": {
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		},
	}
	for _, c := range candidates[runtime.GOOS] {
		if filepath.IsAbs(c) {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
			continue
		}
		if p, err := exec.LookPath(c); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("could not find an installed Chrome/Chromium; install it or adjust tools/render-icon-master")
}

func verifySquarePNG(path string, want int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	cfg, err := png.DecodeConfig(f)
	if err != nil {
		return err
	}
	if cfg.Width != want || cfg.Height != want {
		return fmt.Errorf("%s measured %dx%d, expected %dx%d (check the headless chrome version)", path, cfg.Width, cfg.Height, want, want)
	}
	return nil
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
