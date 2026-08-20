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

.PHONY: build build-debug run scan icons vet clean

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
# --manifest none evita un conflicto de linkeo ("multiple non-default
# manifests") con el manifest que ya trae webview/WebView2 vía cgo.
# Solo hace falta correr esto si cambias el dibujo del ícono.
icons:
	go run ./tools/gen-icon
	cd cmd/autoswitchmonitor && go run github.com/tc-hib/go-winres@latest simply --manifest none --icon ../../assets/icon.ico

vet:
	go vet ./...

clean:
	rm -f $(BINARY) $(BINARY).exe $(BINARY)-debug $(BINARY)-debug.exe
