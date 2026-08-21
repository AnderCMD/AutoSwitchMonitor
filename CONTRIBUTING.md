# Contribuir a AutoSwitchMonitor

¡Gracias por tu interés! Este proyecto es pequeño a propósito — la idea es
mantenerlo ligero y sin dependencias pesadas. Algunas guías antes de mandar
un PR:

## Desarrollo local

```bash
git clone https://github.com/AnderCMD/AutoSwitchMonitor.git
cd AutoSwitchMonitor
make build   # o: go build -o AutoSwitchMonitor(.exe) ./cmd/autoswitchmonitor
make vet
```

En macOS, para probar el equivalente real al `.exe` de Windows (doble clic,
sin consola), usa `make app` en vez de `make build` — arma
`AutoSwitchMonitor.app` (ver `packaging/darwin/Info.plist`).

Requiere Go 1.21+ y `CGO_ENABLED=1` (lo necesitan `getlantern/systray` y
`golang.design/x/hotkey`).

## Estructura

```
cmd/autoswitchmonitor/   entry point + CLI
internal/config/         config.yaml
internal/ddc/            control DDC/CI (nativo en Windows, m1ddc/ddcctl en macOS)
internal/usbwatch/       enumeración USB por sondeo (SetupAPI / system_profiler)
internal/hotkeys/        hotkeys globales
internal/autostart/      activar/desactivar inicio con el sistema
internal/trayapp/        ícono de bandeja + orquestación
internal/appicon/        ícono a color (desde icon_master.png embebido) y silueta template de macOS
assets/icon.svg          diseño fuente del ícono (editar acá los cambios de logo)
tools/render-icon-master/ rasteriza assets/icon.svg a internal/appicon/icon_master.png (necesita Chrome/Chromium)
tools/gen-icon/          regenera assets/icon.png, assets/icon.ico y assets/icon.icns desde icon_master.png
packaging/darwin/        Info.plist del bundle AutoSwitchMonitor.app (make app)
```

## Principios de diseño (léelos antes de proponer una dependencia nueva)

- **Sin libusb/cgo para USB**: usamos las APIs nativas del SO (SetupAPI en
  Windows, `system_profiler` en macOS) para no distribuir/instalar nada
  aparte.
- **DDC/CI nativo solo en Windows** (`Dxva2.dll`). En macOS delegamos en
  `m1ddc`/`ddcctl` porque Apple no publica una API soportada para esto.
- Antes de agregar una dependencia nueva, pregúntate si se puede resolver
  con la librería estándar o una API nativa del SO — el objetivo es que el
  binario siga siendo pequeño y de arranque instantáneo.

## Reportar bugs / pedir features

Abre un issue. Para bugs, incluye: SO y versión, marca/modelo de monitor,
marca/modelo del KVM, y el log de consola si la app lo muestra (corre desde
una terminal, no desde el acceso directo, para ver los `log.Printf`).

## Pull requests

- Un PR = un cambio enfocado. Evita mezclar refactors con features.
- Corre `make vet` antes de abrir el PR.
- Si cambiás el diseño del logo, editá `assets/icon.svg` y corré
  `go run ./tools/render-icon-master` (necesita Chrome/Chromium instalado)
  para regenerar `internal/appicon/icon_master.png`. Después corré
  `make icons` para regenerar `assets/icon.png`, `assets/icon.ico`,
  `assets/icon.icns` (solo en macOS, necesita `iconutil`) y los `.syso` de
  Windows, y commiteá todos esos archivos regenerados.
- La mayor parte del código son bindings a APIs del SO, difíciles de
  testear sin el hardware real (KVM, monitor real); cuando algo sí se
  puede probar contra el SO sin hardware externo (ej. `internal/autostart`
  contra el registro/LaunchAgent real), agrega un test — corre `make test`.
  Para todo lo demás, describe en el PR cómo lo probaste manualmente.
