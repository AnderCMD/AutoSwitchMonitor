<p align="center">
  <img src="assets/icon.png" width="96" height="96" alt="AutoSwitchMonitor">
</p>

<h1 align="center">AutoSwitchMonitor</h1>

<p align="center">
  <a href="https://github.com/AnderCMD/AutoSwitchMonitor/actions/workflows/build.yml"><img src="https://github.com/AnderCMD/AutoSwitchMonitor/actions/workflows/build.yml/badge.svg" alt="build status"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT license"></a>
  <a href="https://github.com/AnderCMD/AutoSwitchMonitor/releases"><img src="https://img.shields.io/github/v/release/AnderCMD/AutoSwitchMonitor?include_prereleases" alt="latest release"></a>
</p>

Automatically switches your monitor's video input (DDC/CI) when you move
your USB KVM switch between PCs, and exposes configurable global hotkeys to
switch inputs manually. Goal: stop using the monitor's physical buttons.

Single binary, no runtime (Go), ~7 MB, runs in the system tray. Windows and
macOS.

## How it works (important to understand before configuring)

Your KVM (as you described it) **only shares keyboard/mouse between 2
PCs**; the video from each PC is connected directly to a different monitor
input (DP, HDMI1, HDMI2). That's why you need to run **one instance of this
app on each of the 2 PCs connected to the KVM**, and optionally a third
instance on the PC connected directly to HDMI2 (only for the manual
hotkey, without autodetection).

Each instance:

1. Watches a specific USB device (e.g. the keyboard/mouse that passes
   through the KVM). When that device **appears** on that PC, it means the
   KVM has just selected you → the app sends a DDC/CI command "switch the
   monitor to my input".
2. It also registers global hotkeys to force a switch to any input at any
   time (useful especially on the PC that isn't on the KVM).

**Real DDC/CI limitation to keep in mind:** many monitors only respond to
DDC/CI commands over the cable of the input that's currently *active*.
That is, to jump from HDMI1 to HDMI2, the command normally has to come
from the PC that's currently being displayed (HDMI1), not from the
"target" PC. This is exactly what the flow above achieves: whoever has
control at that moment is the one who sends the switch command.

## Installation / build

You need Go 1.21+ and `CGO_ENABLED=1` (enabled by default if you have a C
compiler installed; on Windows, `winget install GoLang.Go` already brings
what's needed in most cases — if the build fails, install
[TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or use
`winget install -e --id BrechtSanders.WinLibs.POSIX.UCRT`).

```bash
make build
```

This generates `AutoSwitchMonitor.exe` on Windows or `AutoSwitchMonitor` on
macOS/Linux (the `Makefile` takes care of the correct extension per OS). If
you don't have `make`, the manual equivalent is:

```bash
# Windows — the name MUST explicitly include ".exe": if you pass -o
# without an extension, Go creates the file literally without ".exe" and
# you won't be able to run it by double-clicking or find it in Explorer.
# -ldflags -H=windowsgui prevents a console window from opening alongside
# the tray icon every time you run the .exe.
go build -ldflags="-H=windowsgui" -o AutoSwitchMonitor.exe ./cmd/autoswitchmonitor

# macOS / Linux
go build -o AutoSwitchMonitor ./cmd/autoswitchmonitor
```

### Debugging (viewing the logs)

The normal Windows build (`make build`) doesn't show any console, so the
`log.Printf` output isn't visible at a glance. To debug:

```bash
# Option 1: separate console build, doesn't affect the normal build
make build-debug
./AutoSwitchMonitor-debug.exe

# Option 2: redirect the normal build's output to a file
.\AutoSwitchMonitor.exe 2> debug.log
```

### macOS

macOS **has no public Apple API for DDC/CI**. This app delegates the input
switch to a community-proven command-line tool — install **one** of these
with Homebrew:

```bash
# Apple Silicon (M1/M2/M3...), monitors via USB-C/DP Alt Mode:
brew install waydabber/m1ddc/m1ddc

# Intel Mac:
brew install ddcctl
```

> Note: `m1ddc` sends DDC/CI commands over the AUX channel of
> **USB-C/DisplayPort Alt Mode** — it needs that channel end to end. This
> fails (typically with
> `DDC communication failure: (iokit/?) unknown subsystem error`) if at any
> point along the way there's a conversion to HDMI: the built-in HDMI port
> on base M1/M2 models, or a **USB-C→HDMI adapter/cable**, almost never
> pass that channel through even though the video itself looks fine.
> **Apple Silicon MacBook Air/Pro models have no physical HDMI port** — any
> HDMI output on them is, by definition, already a USB-C→HDMI cable/adapter,
> so on those machines this isn't a rare edge case but the typical
> scenario. Try the standalone command first to rule out your app:
> `m1ddc display list` and then `m1ddc display <N> set input <code>` — if
> that also fails, it's the physical connection, not AutoSwitchMonitor. The
> most reliable way to get DDC/CI working on Apple Silicon is a genuine
> end-to-end DisplayPort path (USB-C→DisplayPort cable, or USB-C→USB-C if
> the monitor has a USB-C input with video) — there Apple doesn't expose
> any official API, but at least the AUX channel reaches the monitor
> intact.
>
> If you only have HDMI available, **it's not 100% impossible, but it
> depends on the chip in your adapter/cable/hub** — many cheap single-port
> adapters don't forward the DDC channel, but several multi-port USB-C hubs
> do. Community reports from
> [MonitorControl](https://github.com/MonitorControl/MonitorControl/discussions/1247):
>
> | Work | Don't work |
> |---|---|
> | Anker USB-C Hub 7-in-1 (with SD-Card) | Syntech USB-C to HDMI Adapter 4K |
> | Anker USB-C Hub 7-in-1 (with LAN) | Atvoiti USB-C to HDMI Adapter |
> | | Baseus Typ-C Hub 4K HDMI RJ45 TF 100W |
>
> If you have a USB-C→HDMI adapter/hub and want to test yours: `m1ddc
> display list` and then `m1ddc display <N> set input <code>` (typical VCP
> codes: `0x0f`=DP1, `0x11`=HDMI1, `0x12`=HDMI2) — if that responds without
> error, your setup does support DDC and AutoSwitchMonitor should work. The
> app retries each input switch 3 times (the DDC channel is prone to
> transient failures even on setups that do work), so a consistent failure
> (not occasional) usually indicates the adapter doesn't forward the
> channel. If you tested yours, open an
> [issue](https://github.com/AnderCMD/AutoSwitchMonitor/issues) letting us
> know whether it worked — the idea is for this table to grow with the
> community.

> **macOS status:** the macOS code (`_darwin.go`) follows the same API as
> Windows, compiles cleanly, and has already been tested running for real
> on a Mac (Apple Silicon). If something fails on yours, open an
> [issue](https://github.com/AnderCMD/AutoSwitchMonitor/issues) — reports
> and PRs from Mac users are appreciated.

#### Tray app (equivalent of the Windows .exe)

`go build -o AutoSwitchMonitor ./cmd/autoswitchmonitor` generates a plain
Unix binary: if you double-click it in Finder, macOS opens it inside a
Terminal window (it's not a real app). To get the real equivalent of the
Windows `.exe` — double-click, no console, no Dock icon, only the tray
icon — build the `.app`:

```bash
make app
```

This compiles the binary and assembles `AutoSwitchMonitor.app` (uses
`packaging/darwin/Info.plist`, which marks the app as `LSUIElement`, and
`assets/icon.icns`, generated by `make icons`). Double-click to open it, or
run `open AutoSwitchMonitor.app`.

`AutoSwitchMonitor.app` is ad-hoc signed (`codesign -s -`) so macOS lets it
run — since it isn't signed with a Developer ID nor notarized, the first
time you may need to right-click → Open instead of a normal double-click,
so Gatekeeper lets you through.

## Configuration

The first time you run the app it creates a `config.yaml` with default
values at:

- Windows: `%APPDATA%\AutoSwitchMonitor\config.yaml`
- macOS: `~/Library/Application Support/AutoSwitchMonitor/config.yaml`

You can see the exact path with:

```bash
AutoSwitchMonitor -config-path
```

### 1. Check your monitor's DDC/CI codes

`config.yaml` ships with typical codes (`dp1=0x0f`, `hdmi1=0x11`,
`hdmi2=0x12`), but they **vary by manufacturer**. If switching inputs does
nothing or switches to the wrong input, look up the "Input Source" table
(VCP 0x60) in your monitor's manual, or try other common values
(`0x01`=VGA, `0x03`=DVI, `0x0f`=DisplayPort1, `0x10`=DisplayPort2,
`0x11`=HDMI1, `0x12`=HDMI2).

### 2. Identify the USB device the KVM watches

Run, on each of the 2 PCs connected to the KVM:

```bash
AutoSwitchMonitor -scan
```

Leave the command running and switch the KVM a couple of times between the
two PCs. You'll see `+ CONNECTED` / `- DISCONNECTED` lines; the
`vendor_id:product_id` that appears/disappears exactly when the KVM
selects/deselects you is the one you need. Copy it into `usb_watch` in that
PC's `config.yaml`:

```yaml
usb_watch:
  vendor_id: "046d"
  product_id: "c547"
  enabled: true
```

On the PC that's **not** on the KVM (the one connected directly to
HDMI2), leave `usb_watch.enabled: false` — it will only use the hotkeys.

### 3. Set your own input per PC

In each `config.yaml`, `own_input` must be the input for *that* PC:

- PC A (on the KVM, wired to DP): `own_input: dp1`
- PC B (on the KVM, wired to HDMI1): `own_input: hdmi1`
- PC C (direct to HDMI2, no KVM): `own_input: hdmi2`, `usb_watch.enabled: false`

### 4. Hotkeys

By default: `Ctrl+Alt+1` → DP1, `Ctrl+Alt+2` → HDMI1, `Ctrl+Alt+3` →
HDMI2, the same on all 3 PCs. Edit them freely in `config.yaml`:

```yaml
hotkeys:
  - modifiers: ["ctrl", "alt"]
    key: "3"
    target: hdmi2
```

Valid modifiers: `ctrl`, `shift`, `alt` (or `option`), `win` (or `cmd`) —
`win`/`cmd` and `alt`/`option` are aliases of each other so the same
config.yaml works on both Windows and macOS. Valid keys: `0`-`9`, `a`-`z`.

After editing `config.yaml`, restart the app for the changes to take
effect (right-click the tray icon → Quit, and open it again).

## Daily use

Run the binary; a system tray icon appears with:

- One item per configured input, to switch manually with the mouse.
- "Start with system" (checkbox, see below).
- "Open config folder".
- "Quit".

### Starting automatically with the system

Enable/disable it directly from the tray icon → **"Start with system"**
(checkbox). No manual steps needed:

- **Windows:** saved in `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
  (per-user, no administrator permissions required).
- **macOS:** creates a LaunchAgent at
  `~/Library/LaunchAgents/dev.andercmd.autoswitchmonitor.plist`.

On macOS, the first time the system will ask for **Accessibility**
permission (Settings → Privacy & Security → Accessibility) so the global
hotkeys work — this is a requirement of `golang.design/x/hotkey`, not
something this app can avoid.

## Project structure

```
cmd/autoswitchmonitor/   entry point + CLI (-scan, -config-path)
internal/config/         loads/saves config.yaml
internal/ddc/            DDC/CI: native via Win32 API on Windows,
                          shells out to m1ddc/ddcctl on macOS
internal/usbwatch/       USB enumeration by polling: SetupAPI on Windows
                          (no libusb/cgo), system_profiler on macOS
internal/hotkeys/        global hotkeys (golang.design/x/hotkey)
internal/autostart/      enable/disable start with system
internal/trayapp/        tray icon + orchestration
internal/appicon/        color icon (from embedded icon_master.png) and silhouette
                          template for the macOS menu bar
tools/render-icon-master/ rasterizes assets/icon.svg to internal/appicon/icon_master.png
tools/gen-icon/          regenerates assets/icon.png, assets/icon.ico and assets/icon.icns
assets/icon.svg          source icon design (edit logo changes here)
assets/                  icon.png / icon.ico / icon.icns used by the tray, the
                          Windows .exe (embedded via go-winres), the
                          macOS .app and this README
packaging/darwin/        Info.plist for the AutoSwitchMonitor.app bundle (make app)
```

## Publishing a release

Every push to a `v*` tag (e.g. `v1.0.0`) triggers the GitHub Actions
workflow, which builds the Windows and macOS binaries and publishes them
automatically as a
[Release](https://github.com/AnderCMD/AutoSwitchMonitor/releases) with
auto-generated notes — no need to upload or build anything by hand.

**From VS Code (without using the terminal):**

1. Make sure your latest commit is already synced (the "Sync Changes"
   button / the cloud icon in the bottom bar).
2. `Ctrl+Shift+P` → type **"Git: Create Tag"** → type the name, e.g.
   `v1.0.0` → Enter (you can leave the message empty).
3. `Ctrl+Shift+P` → **"Git: Push Tags"** → this pushes the tag to GitHub.
4. Within a few minutes, the Release appears on the repo's *Releases* tab
   with the Windows `.exe` and the macOS binary attached.

## Design decisions (why it's built this way)

- **No libusb/cgo for USB**: uses SetupAPI on Windows and
  `system_profiler` on macOS — both ship with the OS, so nothing extra
  needs to be distributed or installed for USB detection.
- **Native DDC/CI only on Windows** (`Dxva2.dll`): it's Microsoft's public,
  stable API, with no external dependencies.
- **DDC/CI via `m1ddc`/`ddcctl` on macOS**: Apple doesn't publish a
  supported API for this; DDC/CI itself on Apple Silicon depends on
  private frameworks that the community already keeps up to date in these
  tools. Reimplementing it here would be fragile and high-maintenance.

## Contributing

PRs and issues are welcome — read [CONTRIBUTING.md](CONTRIBUTING.md)
before you start.

## License

[MIT](LICENSE) © [AnderCMD](https://andercmd.dev)
