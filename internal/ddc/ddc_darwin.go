//go:build darwin

package ddc

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

// El canal DDC/CI (AUX de DisplayPort, o I2C reenviado por un hub/adaptador
// USB-C→HDMI) es propenso a fallos transitorios de comunicación incluso en
// setups que sí funcionan en general. Reintentar unas pocas veces con una
// pausa corta evita reportar error por un solo glitch pasajero.
const (
	setInputAttempts = 3
	setInputRetryGap = 300 * time.Millisecond
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
		return runWithRetry(path, "set", "input", strconv.Itoa(vcpCode))
	}

	if path, err := exec.LookPath("ddcctl"); err == nil {
		return runWithRetry(path, "-d", "1", "-i", strconv.Itoa(vcpCode))
	}

	return fmt.Errorf("no se encontró m1ddc ni ddcctl en el PATH; instala uno con Homebrew " +
		"(brew install waydabber/m1ddc/m1ddc, o brew install ddcctl) para controlar la entrada del monitor")
}

func runWithRetry(path string, args ...string) error {
	var lastErr error
	var lastOut []byte
	for attempt := 1; attempt <= setInputAttempts; attempt++ {
		out, err := exec.Command(path, args...).CombinedOutput()
		if err == nil {
			return nil
		}
		lastErr, lastOut = err, out
		if attempt < setInputAttempts {
			time.Sleep(setInputRetryGap)
		}
	}
	return fmt.Errorf("%s falló tras %d intentos: %w (%s)%s",
		path, setInputAttempts, lastErr, lastOut, adapterHint)
}

// adapterHint se agrega al error cuando falla el DDC/CI: en Apple Silicon
// (sin puerto HDMI físico) el canal DDC solo funciona si el hub/cable
// USB-C→HDMI lo reenvía, y muchos adaptadores baratos de un solo puerto no
// lo hacen aunque el video sí se vea bien. Ver README.md § macOS.
const adapterHint = " — si esto falla siempre (no solo a veces), puede ser que tu " +
	"cable/adaptador USB-C→HDMI no reenvíe el canal DDC/CI: prueba con un hub " +
	"USB-C multipuerto en vez de un adaptador simple, o una conexión DisplayPort " +
	"real si tu monitor la tiene (ver README.md)"
