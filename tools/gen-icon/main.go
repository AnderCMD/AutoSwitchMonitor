// Command gen-icon regenerates assets/icon.png, assets/icon.ico, and (on
// macOS) assets/icon.icns from the same drawing used by the system tray
// icon (internal/appicon), so the icon looks identical everywhere: tray,
// Windows .exe, macOS .app, and the README.
//
// Usage: go run ./tools/gen-icon
package main

import (
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/AnderCMD/AutoSwitchMonitor/internal/appicon"
)

func main() {
	root, err := repoRoot()
	if err != nil {
		log.Fatal(err)
	}
	assetsDir := filepath.Join(root, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		log.Fatal(err)
	}

	master := appicon.Draw(256)
	writePNG(filepath.Join(assetsDir, "icon.png"), master)

	sizes := []int{16, 24, 32, 48, 64, 128, 256}
	images := make([]*image.NRGBA, 0, len(sizes))
	for _, s := range sizes {
		images = append(images, appicon.Draw(s))
	}
	icoPath := filepath.Join(assetsDir, "icon.ico")
	if err := os.WriteFile(icoPath, appicon.EncodeMultiICO(images), 0o644); err != nil {
		log.Fatal(err)
	}

	log.Println("generated:", filepath.Join("assets", "icon.png"))
	log.Println("generated:", filepath.Join("assets", "icon.ico"))

	if runtime.GOOS == "darwin" {
		genICNS(assetsDir)
	}
}

// genICNS builds an .iconset with each size Apple requires (rendered
// directly with appicon.Draw, without rescaling bitmaps) and packages it
// with iconutil, the macOS command-line tool for this.
func genICNS(assetsDir string) {
	if _, err := exec.LookPath("iconutil"); err != nil {
		log.Println("iconutil not found, skipping assets/icon.icns (only available on macOS)")
		return
	}

	iconsetDir := filepath.Join(assetsDir, "icon.iconset")
	if err := os.RemoveAll(iconsetDir); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(iconsetDir, 0o755); err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(iconsetDir)

	// name -> size in pixels to render.
	entries := map[string]int{
		"icon_16x16.png":      16,
		"icon_16x16@2x.png":   32,
		"icon_32x32.png":      32,
		"icon_32x32@2x.png":   64,
		"icon_128x128.png":    128,
		"icon_128x128@2x.png": 256,
		"icon_256x256.png":    256,
		"icon_256x256@2x.png": 512,
		"icon_512x512.png":    512,
		"icon_512x512@2x.png": 1024,
	}
	for name, size := range entries {
		writePNG(filepath.Join(iconsetDir, name), appicon.Draw(size))
	}

	icnsPath := filepath.Join(assetsDir, "icon.icns")
	cmd := exec.Command("iconutil", "-c", "icns", iconsetDir, "-o", icnsPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("iconutil failed: %v (%s)", err, out)
	}
	log.Println("generated:", filepath.Join("assets", "icon.icns"))
}

func writePNG(path string, img *image.NRGBA) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
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
