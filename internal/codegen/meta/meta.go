package meta

import "github.com/Alia5/VIIPER/internal/codegen/scanner"

// Metadata holds all scanned information needed for code generation
// Shared between generator orchestrator and language-specific generators.
type Metadata struct {
	Routes         []scanner.RouteInfo
	DTOs           []scanner.DTOSchema
	DevicePackages map[string]*scanner.DeviceConstants // device name -> constants/maps
	WireTags       *scanner.WireTags                   // parsed viiper:wire comments
	CTypeNames     map[string]string                   // DTO name -> C typedef name (e.g., "Device" -> "device_info")
}
