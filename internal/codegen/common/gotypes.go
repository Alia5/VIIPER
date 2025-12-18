package common

import "strings"

// NormalizeGoType strips pointer and slice prefixes from a Go type string
// and reports whether the original type was a slice or pointer.
// Examples: "*MyType" -> ("MyType", false, true), "[]uint8" -> ("uint8", true, false)
func NormalizeGoType(goType string) (base string, isSlice bool, isPointer bool) {
	base = goType
	if strings.HasPrefix(base, "*") {
		base = strings.TrimPrefix(base, "*")
		isPointer = true
	}
	if strings.HasPrefix(base, "[]") {
		base = strings.TrimPrefix(base, "[]")
		isSlice = true
	}
	return
}
