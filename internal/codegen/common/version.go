package common

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is set via ldflags at build time: -ldflags "-X viiper/internal/codegen/common.Version=x.y.z"
var Version = ""

// GetVersion returns the version string that was set at build time via ldflags.
// Returns "0.0.1-dev" if Version is empty (development builds only).
// For production releases, Version MUST be set via: go build -ldflags "-X viiper/internal/codegen/common.Version=x.y.z"
func GetVersion() (string, error) {
	if Version == "" {
		return "0.0.1-dev", nil
	}

	version := strings.TrimPrefix(Version, "v")
	baseVersion := strings.SplitN(version, "-", 2)[0]
	if !strings.Contains(baseVersion, ".") {
		return "", fmt.Errorf("invalid version format: %s (expected x.y.z)", Version)
	}

	return version, nil
}

// ParseVersion extracts major, minor, patch from version string like "1.2.3" or "1.2.3-dirty"
// Returns major, minor, patch as integers.
func ParseVersion(version string) (major, minor, patch int) {
	parts := strings.SplitN(version, "-", 2)
	version = parts[0]

	nums := strings.Split(version, ".")
	if len(nums) >= 1 {
		major, _ = strconv.Atoi(nums[0])
	}
	if len(nums) >= 2 {
		minor, _ = strconv.Atoi(nums[1])
	}
	if len(nums) >= 3 {
		patch, _ = strconv.Atoi(nums[2])
	}
	return
}
