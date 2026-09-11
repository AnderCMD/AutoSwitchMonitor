# Contributing to AutoSwitchMonitor

Thanks for your interest! This project is intentionally small — the goal
is to keep it lightweight and free of heavy dependencies. A few guidelines
before submitting a PR:

## Local development

```bash
git clone https://github.com/AnderCMD/AutoSwitchMonitor.git
cd AutoSwitchMonitor
make build   # or: go build -o AutoSwitchMonitor(.exe) ./cmd/autoswitchmonitor
make vet
```

On macOS, to test the real equivalent of the Windows `.exe` (double-click,
no console), use `make app` instead of `make build` — it builds
`AutoSwitchMonitor.app` (see `packaging/darwin/Info.plist`).

Requires Go 1.21+ and `CGO_ENABLED=1` (needed by `getlantern/systray` and
`golang.design/x/hotkey`).

## Structure

```
cmd/autoswitchmonitor/   entry point + CLI
internal/config/         config.yaml
internal/ddc/            DDC/CI control (native on Windows, m1ddc/ddcctl on macOS)
internal/usbwatch/       USB enumeration by polling (SetupAPI / system_profiler)
internal/hotkeys/        global hotkeys
internal/autostart/      enable/disable start with system
internal/trayapp/        tray icon + orchestration
internal/appicon/        color icon (from embedded icon_master.png) and macOS silhouette template
assets/icon.svg          source icon design (edit logo changes here)
tools/render-icon-master/ rasterizes assets/icon.svg to internal/appicon/icon_master.png (needs Chrome/Chromium)
tools/gen-icon/          regenerates assets/icon.png, assets/icon.ico and assets/icon.icns from icon_master.png
packaging/darwin/        Info.plist for the AutoSwitchMonitor.app bundle (make app)
```

## Design principles (read these before proposing a new dependency)

- **No libusb/cgo for USB**: we use the OS's native APIs (SetupAPI on
  Windows, `system_profiler` on macOS) so nothing extra needs to be
  distributed or installed.
- **Native DDC/CI only on Windows** (`Dxva2.dll`). On macOS we delegate to
  `m1ddc`/`ddcctl` because Apple doesn't publish a supported API for this.
- Before adding a new dependency, ask yourself whether it can be solved
  with the standard library or a native OS API — the goal is for the
  binary to stay small and start up instantly.

## Reporting bugs / requesting features

Open an issue. For bugs, include: OS and version, monitor brand/model, KVM
brand/model, and the console log if the app shows one (run it from a
terminal, not from the shortcut, to see the `log.Printf` output).

## Pull requests

- One PR = one focused change. Avoid mixing refactors with features.
- Run `make vet` before opening the PR.
- If you change the logo design, edit `assets/icon.svg` and run
  `go run ./tools/render-icon-master` (needs Chrome/Chromium installed) to
  regenerate `internal/appicon/icon_master.png`. Then run `make icons` to
  regenerate `assets/icon.png`, `assets/icon.ico`, `assets/icon.icns`
  (macOS only, needs `iconutil`) and the Windows `.syso` files, and commit
  all of those regenerated files.
- Most of the code consists of bindings to OS APIs, which are hard to test
  without real hardware (KVM, real monitor); when something can be tested
  against the OS without external hardware (e.g. `internal/autostart`
  against the real registry/LaunchAgent), add a test — run `make test`.
  For everything else, describe in the PR how you tested it manually.
