BINARY := autoswitchmonitor

ifeq ($(OS),Windows_NT)
	EXT := .exe
else
	EXT :=
endif

.PHONY: build run scan icons vet clean

build:
	go build -o $(BINARY)$(EXT) ./cmd/autoswitchmonitor

run: build
	./$(BINARY)$(EXT)

scan:
	go run ./cmd/autoswitchmonitor -scan

# Regenera assets/icon.png + assets/icon.ico desde internal/appicon, y el
# recurso de ícono embebido en el .exe de Windows (cmd/autoswitchmonitor/rsrc_windows_*.syso).
# Solo hace falta correrlo si cambias el dibujo del ícono.
icons:
	go run ./tools/gen-icon
	cd cmd/autoswitchmonitor && go run github.com/tc-hib/go-winres@latest simply --icon ../../assets/icon.ico

vet:
	go vet ./...

clean:
	rm -f $(BINARY) $(BINARY).exe
