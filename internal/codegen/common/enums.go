package common

import (
	"sort"
	"strings"
)

// SanitizeLeadingDigit prefixes names that start with a digit with "Num"
// to keep identifiers valid in target languages.
func SanitizeLeadingDigit(name string) string {
	if name == "" {
		return ""
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "Num" + name
	}
	return name
}

// TrimPrefixAndSanitize splits a constant full name into its enum prefix and member,
// and sanitizes the member (e.g., leading digits -> "Num...").
// Example: "Key_1" => ("Key_", "Num1"), "ModifierShift" => ("Modifier", "Shift").
func TrimPrefixAndSanitize(full string) (prefix, member string) {
	prefix = ExtractPrefix(full)
	member = strings.TrimPrefix(full, prefix)
	member = SanitizeLeadingDigit(member)
	return
}

// SortedStringKeys returns the sorted keys of a map[string]any.
func SortedStringKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
