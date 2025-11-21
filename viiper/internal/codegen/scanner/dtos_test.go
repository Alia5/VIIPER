package scanner

import (
	"encoding/json"
	"testing"
)

func TestScanDTOs(t *testing.T) {
	// Scan the apitypes package
	schemas, err := ScanDTOsInPackage("../../../pkg/apitypes")
	if err != nil {
		t.Fatalf("ScanDTOsInPackage failed: %v", err)
	}

	if len(schemas) == 0 {
		t.Fatal("expected at least one DTO schema, got none")
	}

	t.Logf("Found %d DTO schemas", len(schemas))

	// Expected DTOs
	expectedDTOs := map[string]bool{
		"ApiError":             true,
		"BusListResponse":      true,
		"BusCreateResponse":    true,
		"BusRemoveResponse":    true,
		"Device":               true,
		"DevicesListResponse":  true,
		"DeviceRemoveResponse": true,
	}

	foundDTOs := make(map[string]bool)
	for _, schema := range schemas {
		foundDTOs[schema.Name] = true
	}

	for expectedDTO := range expectedDTOs {
		if !foundDTOs[expectedDTO] {
			t.Errorf("expected to find DTO %q, but it was not discovered", expectedDTO)
		}
	}

	// Print all discovered DTOs as JSON for inspection
	t.Log("Discovered DTO schemas:")
	for _, schema := range schemas {
		data, _ := json.MarshalIndent(schema, "", "  ")
		t.Logf("%s", data)
	}

	// Verify BusCreateResponse has correct structure
	for _, schema := range schemas {
		if schema.Name == "BusCreateResponse" {
			if len(schema.Fields) != 1 {
				t.Errorf("BusCreateResponse: expected 1 field, got %d", len(schema.Fields))
			}
			if len(schema.Fields) > 0 {
				field := schema.Fields[0]
				if field.Name != "BusID" {
					t.Errorf("BusCreateResponse: expected field name 'BusID', got %q", field.Name)
				}
				if field.JSONName != "busId" {
					t.Errorf("BusCreateResponse: expected JSON name 'busId', got %q", field.JSONName)
				}
				if field.Type != "uint32" {
					t.Errorf("BusCreateResponse: expected type 'uint32', got %q", field.Type)
				}
				if field.TypeKind != "primitive" {
					t.Errorf("BusCreateResponse: expected typeKind 'primitive', got %q", field.TypeKind)
				}
			}
		}

		// Verify DevicesListResponse has array field
		if schema.Name == "DevicesListResponse" {
			if len(schema.Fields) != 1 {
				t.Errorf("DevicesListResponse: expected 1 field, got %d", len(schema.Fields))
			}
			if len(schema.Fields) > 0 {
				field := schema.Fields[0]
				if field.Name != "Devices" {
					t.Errorf("DevicesListResponse: expected field name 'Devices', got %q", field.Name)
				}
				if field.Type != "[]Device" {
					t.Errorf("DevicesListResponse: expected type '[]Device', got %q", field.Type)
				}
				if field.TypeKind != "slice" {
					t.Errorf("DevicesListResponse: expected typeKind 'slice', got %q", field.TypeKind)
				}
			}
		}
	}
}
