//go:build !windows

package util

func IsRunFromGUI() bool {
	// On non-Windows, always return false.
	// We only use this to spawn the viiper server without going through hoops...
	// On Linux you can use nohup, systemd, and bazillion other ways...
	// I'd like to also assume Linux users are familiar with a CLI
	// if not... they should learn!
	return false
}

func HideConsoleWindow() {
	// No-op on non-Windows platforms
}
