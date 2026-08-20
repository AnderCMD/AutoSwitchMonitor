// Command gen-icon regenera assets/icon.png y assets/icon.ico a partir del
// mismo dibujo que usa el ícono de la bandeja del sistema
// (internal/appicon), para que el ícono se vea idéntico en todas partes:
// bandeja, .exe de Windows y README.
//
// Uso: go run ./tools/gen-icon
package main

import (
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

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

	log.Println("generado:", filepath.Join("assets", "icon.png"))
	log.Println("generado:", filepath.Join("assets", "icon.ico"))
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
