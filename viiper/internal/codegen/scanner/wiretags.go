package scanner

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// WireField represents a single field in a wire protocol struct
type WireField struct {
	Name string `json:"name"` // Field name (e.g., "modifiers", "keys")
	Type string `json:"type"` // Wire type token (e.g., "u8", "i16", may include array marker like "u8*count")
	Spec string `json:"spec"` // Full spec from tag (e.g., "keys:u8*count")
}

// WireTag represents a parsed viiper:wire comment
type WireTag struct {
	Device    string      `json:"device"`    // "keyboard", "mouse", "xbox360"
	Direction string      `json:"direction"` // "c2s" or "s2c"
	Fields    []WireField `json:"fields"`
}

// WireTags holds all wire tags for all devices
type WireTags struct {
	Tags map[string]map[string]*WireTag // device -> direction -> tag
}

// wireTagPattern matches: viiper:wire <device> <direction> field:type ...
var wireTagPattern = regexp.MustCompile(`viiper:wire\s+(\w+)\s+(c2s|s2c)\s+(.+)`)

// ScanWireTags scans all device packages for viiper:wire comments
func ScanWireTags(devicePkgPaths []string) (*WireTags, error) {
	result := &WireTags{
		Tags: make(map[string]map[string]*WireTag),
	}

	for _, pkgPath := range devicePkgPaths {
		entries, err := os.ReadDir(pkgPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %w", pkgPath, err)
		}

		fset := token.NewFileSet()
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}

			filePath := filepath.Join(pkgPath, entry.Name())
			file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
			if err != nil {
				continue
			}

			for _, commentGroup := range file.Comments {
				for _, comment := range commentGroup.List {
					if tag := parseWireTag(comment.Text); tag != nil {
						if result.Tags[tag.Device] == nil {
							result.Tags[tag.Device] = make(map[string]*WireTag)
						}
						result.Tags[tag.Device][tag.Direction] = tag
					}
				}
			}
		}
	}

	return result, nil
}

// parseWireTag parses a single viiper:wire comment line
func parseWireTag(comment string) *WireTag {
	text := strings.TrimSpace(strings.TrimPrefix(comment, "//"))
	text = strings.TrimSpace(strings.TrimPrefix(text, "/*"))
	text = strings.TrimSpace(strings.TrimSuffix(text, "*/"))

	matches := wireTagPattern.FindStringSubmatch(text)
	if matches == nil {
		return nil
	}

	device := matches[1]
	direction := matches[2]
	fieldSpecs := strings.Fields(matches[3])

	tag := &WireTag{
		Device:    device,
		Direction: direction,
		Fields:    []WireField{},
	}

	for _, spec := range fieldSpecs {
		if field := parseWireField(spec); field != nil {
			tag.Fields = append(tag.Fields, *field)
		}
	}

	return tag
}

func parseWireField(spec string) *WireField {
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	name := parts[0]
	typeSpec := parts[1]

	return &WireField{
		Name: name,
		Type: typeSpec,
		Spec: spec,
	}
}

// HasDirection checks if a device has a wire tag for the given direction
func (wt *WireTags) HasDirection(device, direction string) bool {
	if deviceTags, ok := wt.Tags[device]; ok {
		_, exists := deviceTags[direction]
		return exists
	}
	return false
}

// GetTag retrieves the wire tag for a device and direction
func (wt *WireTags) GetTag(device, direction string) *WireTag {
	if deviceTags, ok := wt.Tags[device]; ok {
		return deviceTags[direction]
	}
	return nil
}
