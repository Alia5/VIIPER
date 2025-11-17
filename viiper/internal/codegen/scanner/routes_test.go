package scanner

import (
	"encoding/json"
	"testing"
)

func TestScanRoutes(t *testing.T) {
	// Scan the actual server.go file where routes are registered
	routes, err := ScanRoutes("../../cmd/server.go")
	if err != nil {
		t.Fatalf("ScanRoutes failed: %v", err)
	}

	if len(routes) == 0 {
		t.Fatal("expected at least one route, got none")
	}

	// Verify we found the expected routes
	expectedPaths := map[string]bool{
		"bus/list":               true,
		"bus/create":             true,
		"bus/remove":             true,
		"bus/{id}/list":          true,
		"bus/{id}/add":           true,
		"bus/{id}/remove":        true,
		"bus/{busId}/{deviceid}": true,
	}

	foundPaths := make(map[string]bool)
	for _, route := range routes {
		foundPaths[route.Path] = true
	}

	for expectedPath := range expectedPaths {
		if !foundPaths[expectedPath] {
			t.Errorf("expected to find route %q, but it was not discovered", expectedPath)
		}
	}

	// Print discovered routes as JSON for manual inspection
	t.Log("Discovered routes:")
	for _, route := range routes {
		data, _ := json.MarshalIndent(route, "", "  ")
		t.Logf("%s", data)
	}
}

func TestScanHandlerArgs(t *testing.T) {
	// Scan handler implementations
	handlerInfo, err := ScanHandlerArgs("../../server/api/handler")
	if err != nil {
		t.Fatalf("ScanHandlerArgs failed: %v", err)
	}

	t.Logf("Found %d handlers with argument usage", len(handlerInfo))

	// Print handler info for inspection
	for name, info := range handlerInfo {
		t.Logf("Handler: %s", name)
		data, _ := json.MarshalIndent(info, "  ", "  ")
		t.Logf("  %s", data)
	}

	// Verify BusCreate uses args
	if info, ok := handlerInfo["BusCreate"]; ok {
		if !info.UsesArgs {
			t.Errorf("expected BusCreate to use req.Args")
		}
	}
}

func TestEnrichRoutesWithHandlerInfo(t *testing.T) {
	// Scan routes
	routes, err := ScanRoutes("../../cmd/server.go")
	if err != nil {
		t.Fatalf("ScanRoutes failed: %v", err)
	}

	// Enrich with handler argument info
	enriched, err := EnrichRoutesWithHandlerInfo(routes, "../../server/api/handler")
	if err != nil {
		t.Fatalf("EnrichRoutesWithHandlerInfo failed: %v", err)
	}

	t.Log("Enriched routes:")
	for _, route := range enriched {
		data, _ := json.MarshalIndent(route, "", "  ")
		t.Logf("%s", data)
	}

	// Verify bus/create has argument info
	for _, route := range enriched {
		if route.Path == "bus/create" && len(route.Arguments) > 0 {
			t.Logf("bus/create correctly identified with %d argument(s)", len(route.Arguments))
		}
	}
}
