package autostart

import "testing"

// TestEnableDisable exercises the full cycle against the real OS (Windows
// registry / macOS LaunchAgent). There's no sensible way to mock these
// APIs, so this test writes and removes the real autostart entry for the
// test binary — it cleans up after itself even if it fails.
func TestEnableDisable(t *testing.T) {
	t.Cleanup(func() { _ = Disable() })

	if enabled, err := IsEnabled(); err != nil {
		t.Fatalf("IsEnabled before Enable: %v", err)
	} else if enabled {
		t.Fatal("expected disabled before Enable()")
	}

	if err := Enable(); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if enabled, err := IsEnabled(); err != nil {
		t.Fatalf("IsEnabled after Enable: %v", err)
	} else if !enabled {
		t.Fatal("expected enabled after Enable()")
	}

	if err := Disable(); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if enabled, err := IsEnabled(); err != nil {
		t.Fatalf("IsEnabled after Disable: %v", err)
	} else if enabled {
		t.Fatal("expected disabled after Disable()")
	}
}
