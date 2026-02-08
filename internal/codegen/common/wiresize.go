package common

import (
	"sort"
	"strconv"
	"strings"

	"github.com/Alia5/VIIPER/internal/codegen/meta"
	"github.com/Alia5/VIIPER/internal/codegen/scanner"
)

// WireTypeSize returns the size in bytes of a wire protocol type.
func WireTypeSize(wireType string) int {
	switch wireType {
	case "u8", "i8":
		return 1
	case "u16", "i16":
		return 2
	case "u32", "i32":
		return 4
	case "u64", "i64":
		return 8
	default:
		return 1
	}
}

// CalculateOutputSize computes the exact size in bytes of a device's output (s2c) message.
// Returns 0 if the tag is nil or device has no output.
// For variable-length fields (e.g., "u8*count"), returns 0 to indicate dynamic size.
func CalculateOutputSize(tag *scanner.WireTag) int {
	if tag == nil {
		return 0
	}

	total := 0
	for _, field := range tag.Fields {
		wireType := field.Type
		if idx := strings.Index(wireType, "*"); idx >= 0 {
			baseType := wireType[:idx]
			countToken := wireType[idx+1:]
			if n, err := strconv.Atoi(countToken); err == nil {
				total += WireTypeSize(baseType) * n
				continue
			}
			return 0
		}
		total += WireTypeSize(wireType)
	}

	return total
}

// GetWireTag returns the wire tag for a device and direction from metadata.
// Direction can be "input"/"c2s" or "output"/"s2c".
func GetWireTag(md *meta.Metadata, deviceName, direction string) *scanner.WireTag {
	if md.WireTags == nil {
		return nil
	}
	dir := direction
	if direction == "input" {
		dir = "c2s"
	} else if direction == "output" {
		dir = "s2c"
	}
	return md.WireTags.GetTag(deviceName, dir)
}

// GetWireFields returns the wire fields for a device and direction from metadata.
func GetWireFields(md *meta.Metadata, deviceName, direction string) []scanner.WireField {
	tag := GetWireTag(md, deviceName, direction)
	if tag == nil {
		return nil
	}
	return tag.Fields
}

// HasWireTag returns true if a wire tag exists for the device and direction.
func HasWireTag(md *meta.Metadata, deviceName, direction string) bool {
	return GetWireTag(md, deviceName, direction) != nil
}

// ExtractPathParams parses a route pattern like "bus/{id}/list" and returns
// the parameter names in order (e.g., ["id"]).
func ExtractPathParams(path string) []string {
	var params []string
	for _, part := range strings.Split(path, "/") {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			params = append(params, part[1:len(part)-1])
		}
	}
	return params
}

// MapEntry is used for sorted iteration over map entries in templates.
type MapEntry struct {
	Key   string
	Value any
}

// SortedMapEntries returns map entries sorted by key for deterministic output.
func SortedMapEntries(entries map[string]interface{}) []MapEntry {
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]MapEntry, 0, len(keys))
	for _, k := range keys {
		result = append(result, MapEntry{Key: k, Value: entries[k]})
	}
	return result
}
