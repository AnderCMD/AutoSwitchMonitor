// Command render-icon-master rasteriza assets/icon.svg (el diseño fuente:
// anillo de cristal, monitor y toggle "ON") a un PNG maestro de alta
// resolución, usando Chrome/Chromium en modo headless —el propio dibujo usa
// <text> y filtros (feGaussianBlur, feDropShadow) que los rasterizadores
// SVG puros de Go no soportan, así que este paso necesita un motor de
// render real—.
//
// Solo hace falta correrlo cuando cambia el diseño del logo (assets/icon.svg):
// escribe internal/appicon/icon_master.png, que gen-icon (100% Go, sin
// dependencias externas) reescala luego a cada tamaño para
// assets/icon.png/.ico/.icns y el ícono de la bandeja en tiempo de
// ejecución. No hace falta Chrome para compilar ni para correr la app.
//
// Uso: go run ./tools/render-icon-master
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
		log.Fatalf("no se encontró %s: %v", svgPath, err)
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
		log.Fatalf("chrome headless falló: %v\n%s", err, out)
	}

	if err := verifySquarePNG(outPath, masterSize); err != nil {
		log.Fatal(err)
	}

	log.Println("generado:", filepath.Join("internal", "appicon", "icon_master.png"))
	log.Println("corré 'go run ./tools/gen-icon' para regenerar assets/icon.png/.ico/.icns")
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
	return "", fmt.Errorf("no se encontró Chrome/Chromium instalado; instalalo o ajustá tools/render-icon-master")
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
		return fmt.Errorf("%s midió %dx%d, se esperaba %dx%d (revisá la versión de chrome headless)", path, cfg.Width, cfg.Height, want, want)
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
