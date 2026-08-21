BINARY := AutoSwitchMonitor

ifeq ($(OS),Windows_NT)
	EXT := .exe
	# Subsistema GUI: sin esto, Windows abre una ventana de consola junto
	# con el ícono de bandeja cada vez que se ejecuta el .exe.
	LDFLAGS := -H=windowsgui
else
	EXT :=
	LDFLAGS :=
endif

.PHONY: build build-debug run scan icons app vet test clean

build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY)$(EXT) ./cmd/autoswitchmonitor

# Build con consola visible (subsistema console), para ver los log.Printf
# en vivo mientras depuras. Ver también: make run, o `2> debug.log`.
build-debug:
	go build -o $(BINARY)-debug$(EXT) ./cmd/autoswitchmonitor

run: build-debug
	./$(BINARY)-debug$(EXT)

scan:
	go run ./cmd/autoswitchmonitor -scan

# Regenera assets/icon.png + assets/icon.ico desde internal/appicon, y el
# recurso de ícono embebido en el .exe de Windows (cmd/autoswitchmonitor/rsrc_windows_*.syso).
# Solo hace falta correrlo si cambias el dibujo del ícono.
icons:
	go run ./tools/gen-icon
	cd cmd/autoswitchmonitor && go run github.com/tc-hib/go-winres@latest simply --icon ../../assets/icon.ico

# Empaqueta AutoSwitchMonitor.app: en macOS, doble clic en el binario
# suelto lo abre dentro de una ventana de Terminal (no es un .app). Este
# target arma el bundle real, equivalente al .exe con -H=windowsgui de
# Windows: LSUIElement en Info.plist lo declara app de bandeja (sin ícono
# en el Dock ni ventana de consola).
app: build
	rm -rf $(BINARY).app
	mkdir -p $(BINARY).app/Contents/MacOS $(BINARY).app/Contents/Resources
	cp $(BINARY) $(BINARY).app/Contents/MacOS/$(BINARY)
	cp packaging/darwin/Info.plist $(BINARY).app/Contents/Info.plist
	if [ -f assets/icon.icns ]; then cp assets/icon.icns $(BINARY).app/Contents/Resources/icon.icns; fi
	codesign --force --deep -s - $(BINARY).app 2>/dev/null || true
	@echo "Bundle creado: $(BINARY).app (doble clic para abrir)"

vet:
	go vet ./...

# internal/autostart escribe/borra la entrada real de autoarranque del SO
# (registro en Windows, LaunchAgent en macOS) y la limpia sola al terminar.
test:
	go test ./...

clean:
	rm -f $(BINARY) $(BINARY).exe $(BINARY)-debug $(BINARY)-debug.exe
	rm -rf $(BINARY).app
