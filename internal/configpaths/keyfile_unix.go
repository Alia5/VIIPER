//go:build !windows

package configpaths

import (
	"os"
	"path/filepath"
)

// KeyFileDir returns the directory where the API key file should be stored.
// On Unix, root services use /etc/viiper.
func KeyFileDir() (string, error) {
	if os.Geteuid() == 0 {
		return filepath.Join(string(os.PathSeparator), "etc", "viiper"), nil
	}
	return DefaultConfigDir()
}
