//go:build windows

package configpaths

// KeyFileDir returns the directory where the API key file should be stored.
func KeyFileDir() (string, error) {
	return DefaultConfigDir()
}
