package scanner

import (
	"encoding/json"
	"testing"
)

type testCase struct {
	name string
	run  func(t *testing.T)
}

func TestScannerSuite(t *testing.T) {
	cases := []testCase{
		{
			name: "ScanRoutes discovers expected paths",
			run: func(t *testing.T) {
				routes, err := ScanRoutes("../../cmd/server.go")
				if err != nil {
					t.Fatalf("ScanRoutes failed: %v", err)
				}
				if len(routes) == 0 {
					t.Fatal("expected at least one route, got none")
				}
				expected := map[string]bool{
					"bus/list":               true,
					"bus/create":             true,
					"bus/remove":             true,
					"bus/{id}/list":          true,
					"bus/{id}/add":           true,
					"bus/{id}/remove":        true,
					"bus/{busId}/{deviceid}": true,
				}
				found := make(map[string]bool)
				for _, r := range routes {
					found[r.Path] = true
				}
				for p := range expected {
					if !found[p] {
						t.Errorf("expected route %s not found", p)
					}
				}
				t.Log("Discovered routes (raw):")
				for _, r := range routes {
					data, _ := json.MarshalIndent(r, "", "  ")
					t.Logf("%s", data)
				}
			},
		},
		{
			name: "EnrichRoutes classifies payload kinds correctly",
			run: func(t *testing.T) {
				routes, err := ScanRoutes("../../cmd/server.go")
				if err != nil {
					t.Fatalf("ScanRoutes failed: %v", err)
				}
				enriched, err := EnrichRoutesWithHandlerInfo(routes, "../../server/api/handler")
				if err != nil {
					t.Fatalf("EnrichRoutesWithHandlerInfo failed: %v", err)
				}
				seen := map[string]RouteInfo{}
				for _, r := range enriched {
					seen[r.Path] = r
				}
				assertPayload := func(path string, kind PayloadKind, required bool) {
					v, ok := seen[path]
					if !ok {
						t.Errorf("%s missing", path)
						return
					}
					if v.Payload.Kind != kind || v.Payload.Required != required {
						t.Errorf("%s expected kind=%s required=%v got %+v", path, kind, required, v.Payload)
					}
				}
				assertPayload("bus/{id}/add", PayloadJSON, true)
				assertPayload("bus/create", PayloadNumeric, false)
				assertPayload("bus/remove", PayloadNumeric, true)
				assertPayload("bus/{id}/remove", PayloadString, true)
				assertPayload("bus/list", PayloadNone, false)
				assertPayload("bus/{id}/list", PayloadNone, false)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { tc.run(t) })
	}
}
