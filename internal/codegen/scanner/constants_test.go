package scanner

import (
	"path/filepath"
	"testing"
)

func TestScanKeyboardConstants(t *testing.T) {
	// Relative path from scanner package to keyboard device
	keyboardPath := filepath.Join("..", "..", "..", "device", "keyboard")

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
	xbox360Path := filepath.Join("..", "..", "..", "device", "xbox360")

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

func TestScanNs2ProConstants(t *testing.T) {
	ns2proPath := filepath.Join("..", "..", "..", "device", "ns2pro")

	result, err := ScanDeviceConstants(ns2proPath)
	if err != nil {
		t.Fatalf("Failed to scan ns2pro constants: %v", err)
	}

	constants := make(map[string]ConstantInfo, len(result.Constants))
	for _, c := range result.Constants {
		constants[c.Name] = c
	}

	if got := constants["ButtonB"].Value; got != uint64(1) {
		t.Fatalf("expected ButtonB to equal 1, got %#v", got)
	}
	if got := constants["ButtonA"].Value; got != uint64(2) {
		t.Fatalf("expected ButtonA to equal 2, got %#v", got)
	}
	if got := constants["ButtonHeadset"].Value; got != uint64(1<<21) {
		t.Fatalf("expected ButtonHeadset to equal 1<<21, got %#v", got)
	}
	if got := constants["DefaultSerialEnding"].Value; got != "00" {
		t.Fatalf("expected DefaultSerialEnding to equal 00, got %#v", got)
	}
	if got := constants["DefaultSerial"].Value; got != "VIIPER-NS2PRO-00" {
		t.Fatalf("expected DefaultSerial to equal VIIPER-NS2PRO-00, got %#v", got)
	}
}
