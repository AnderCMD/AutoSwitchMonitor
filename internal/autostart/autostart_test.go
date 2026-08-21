package autostart

import "testing"

// TestEnableDisable ejercita el ciclo completo contra el SO real (registro
// de Windows / LaunchAgent de macOS). No hay forma sensata de mockear estas
// APIs, así que esta prueba escribe y borra la entrada real de autoarranque
// para el binario de test — se limpia sola al final incluso si falla.
func TestEnableDisable(t *testing.T) {
	t.Cleanup(func() { _ = Disable() })

	if enabled, err := IsEnabled(); err != nil {
		t.Fatalf("IsEnabled antes de Enable: %v", err)
	} else if enabled {
		t.Fatal("esperaba deshabilitado antes de Enable()")
	}

	if err := Enable(); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if enabled, err := IsEnabled(); err != nil {
		t.Fatalf("IsEnabled después de Enable: %v", err)
	} else if !enabled {
		t.Fatal("esperaba habilitado después de Enable()")
	}

	if err := Disable(); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if enabled, err := IsEnabled(); err != nil {
		t.Fatalf("IsEnabled después de Disable: %v", err)
	} else if enabled {
		t.Fatal("esperaba deshabilitado después de Disable()")
	}
}
