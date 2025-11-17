package scanner

import (
	"path/filepath"
	"testing"
)

func TestScanKeyboardConstants(t *testing.T) {
	// Relative path from scanner package to keyboard device
	keyboardPath := filepath.Join("..", "..", "..", "pkg", "device", "keyboard")

	result, err := ScanDeviceConstants(keyboardPath)
	if err != nil {
		t.Fatalf("Failed to scan keyboard constants: %v", err)
	}

	if result.DeviceType != "keyboard" {
		t.Errorf("Expected deviceType 'keyboard', got '%s'", result.DeviceType)
	}

	if len(result.Constants) == 0 {
		t.Errorf("Expected to find keyboard constants, got none")
	}

	// Should find 3 maps: KeyName, CharToKey, ShiftChars
	if len(result.Maps) != 3 {
		t.Errorf("Expected 3 maps, got %d", len(result.Maps))
		for i, m := range result.Maps {
			t.Logf("Map %d: %s (keyType: %s, valueType: %s, entries: %d)",
				i, m.Name, m.KeyType, m.ValueType, len(m.Entries))
		}
	}

	t.Logf("Found %d constants and %d maps", len(result.Constants), len(result.Maps))
}

func TestScanXbox360Constants(t *testing.T) {
	xbox360Path := filepath.Join("..", "..", "..", "pkg", "device", "xbox360")

	result, err := ScanDeviceConstants(xbox360Path)
	if err != nil {
		t.Fatalf("Failed to scan xbox360 constants: %v", err)
	}

	// Should find 15 button constants
	if len(result.Constants) != 15 {
		t.Errorf("Expected 15 button constants, got %d", len(result.Constants))
	}

	// Xbox360 has no maps
	if len(result.Maps) != 0 {
		t.Errorf("Expected 0 maps, got %d", len(result.Maps))
	}

	t.Logf("Found %d constants", len(result.Constants))
}
