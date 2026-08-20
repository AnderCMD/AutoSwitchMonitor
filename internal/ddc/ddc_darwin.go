//go:build darwin

package ddc

import (
	"fmt"
	"os/exec"
	"strconv"
)

// macOS no tiene una API pública de Apple para DDC/CI: el propio DDC/CI en
// Apple Silicon depende de frameworks privados (DisplayServices/AVService)
// que Apple puede cambiar sin aviso entre versiones de macOS. En vez de
// reimplementar esos bindings privados aquí (frágil y de alto mantenimiento),
// delegamos en herramientas de línea de comandos ya probadas por la
// comunidad:
//
//   - m1ddc (Apple Silicon, vía USB-C/DisplayPort Alt Mode):
//     brew install m1ddc   (https://github.com/waydabber/m1ddc)
//   - ddcctl (Intel Mac):
//     brew install ddcctl  (https://github.com/kfix/ddcctl)
//
// setInput intenta m1ddc primero y cae a ddcctl si no está instalado.
func setInput(vcpCode int) error {
	if path, err := exec.LookPath("m1ddc"); err == nil {
		cmd := exec.Command(path, "set", "input", strconv.Itoa(vcpCode))
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("m1ddc falló: %w (%s)", err, out)
		}
		return nil
	}

	if path, err := exec.LookPath("ddcctl"); err == nil {
		cmd := exec.Command(path, "-d", "1", "-i", strconv.Itoa(vcpCode))
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("ddcctl falló: %w (%s)", err, out)
		}
		return nil
	}

	return fmt.Errorf("no se encontró m1ddc ni ddcctl en el PATH; instala uno con Homebrew " +
		"(brew install waydabber/m1ddc/m1ddc, o brew install ddcctl) para controlar la entrada del monitor")
}
