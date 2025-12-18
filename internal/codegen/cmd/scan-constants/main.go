package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

func main() {
	viiperRoot := "."
	if len(os.Args) > 1 {
		viiperRoot = os.Args[1]
	}

	// Discover device packages automatically
	deviceBaseDir := filepath.Join(viiperRoot, "pkg", "device")
	entries, err := os.ReadDir(deviceBaseDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading device directory: %v\n", err)
		os.Exit(1)
	}

	var devices []string
	for _, entry := range entries {
		if entry.IsDir() {
			devices = append(devices, entry.Name())
		}
	}

	for _, device := range devices {
		devicePath := filepath.Join(deviceBaseDir, device)

		result, err := scanner.ScanDeviceConstants(devicePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning %s: %v\n", device, err)
			continue
		}

		fmt.Printf("\n=== %s ===\n", result.DeviceType)
		fmt.Printf("Constants: %d\n", len(result.Constants))
		fmt.Printf("Maps: %d\n", len(result.Maps))

		if len(result.Constants) > 0 {
			fmt.Printf("\nFirst 10 constants:\n")
			for i, c := range result.Constants {
				if i >= 10 {
					break
				}
				fmt.Printf("  %s = %v (%s)\n", c.Name, c.Value, c.Type)
			}
			if len(result.Constants) > 10 {
				fmt.Printf("  ... and %d more\n", len(result.Constants)-10)
			}
		}

		if len(result.Maps) > 0 {
			fmt.Printf("\nMaps:\n")
			for _, m := range result.Maps {
				fmt.Printf("  %s: map[%s]%s (%d entries)\n",
					m.Name, m.KeyType, m.ValueType, len(m.Entries))
			}
		}
	}
}
