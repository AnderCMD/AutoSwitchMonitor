BINARY := AutoSwitchMonitor

ifeq ($(OS),Windows_NT)
	EXT := .exe
	# GUI subsystem: without this, Windows opens a console window alongside
	# the tray icon every time the .exe runs.
	LDFLAGS := -H=windowsgui
else
	EXT :=
	LDFLAGS :=
endif

.PHONY: build build-debug run scan icons app vet test clean

build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY)$(EXT) ./cmd/autoswitchmonitor

# Build with a visible console (console subsystem), to see log.Printf
# output live while debugging. See also: make run, or `2> debug.log`.
build-debug:
	go build -o $(BINARY)-debug$(EXT) ./cmd/autoswitchmonitor

run: build-debug
	./$(BINARY)-debug$(EXT)

scan:
	go run ./cmd/autoswitchmonitor -scan

# Regenerates assets/icon.png + assets/icon.ico from internal/appicon, and
# the icon resource embedded in the Windows .exe (cmd/autoswitchmonitor/rsrc_windows_*.syso).
# Only needs to be run if you change the icon artwork.
icons:
	go run ./tools/gen-icon
	cd cmd/autoswitchmonitor && go run github.com/tc-hib/go-winres@latest simply --icon ../../assets/icon.ico

# Packages AutoSwitchMonitor.app: on macOS, double-clicking the bare
# binary opens it inside a Terminal window (it's not a .app). This
# target builds the real bundle, equivalent to Windows' .exe with
# -H=windowsgui: LSUIElement in Info.plist declares it a tray app (no
# Dock icon, no console window).
app: build
	rm -rf $(BINARY).app
	mkdir -p $(BINARY).app/Contents/MacOS $(BINARY).app/Contents/Resources
	cp $(BINARY) $(BINARY).app/Contents/MacOS/$(BINARY)
	cp packaging/darwin/Info.plist $(BINARY).app/Contents/Info.plist
	if [ -f assets/icon.icns ]; then cp assets/icon.icns $(BINARY).app/Contents/Resources/icon.icns; fi
	codesign --force --deep -s - $(BINARY).app 2>/dev/null || true
	@echo "Bundle created: $(BINARY).app (double-click to open)"

vet:
	go vet ./...

# internal/autostart writes/removes the real OS autostart entry
# (registry on Windows, LaunchAgent on macOS) and cleans up after itself.
test:
	go test ./...

clean:
	rm -f $(BINARY) $(BINARY).exe $(BINARY)-debug $(BINARY)-debug.exe
	rm -rf $(BINARY).app
