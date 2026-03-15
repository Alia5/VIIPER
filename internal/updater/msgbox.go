package updater

import (
	"runtime"

	"github.com/ncruces/zenity"
)

func showMessageBox(title, message string) int {
	options := []string{"Update Now", "View on GitHub", "Remind Me Later", "Skip This Version"}
	if runtime.GOOS == "linux" {
		options = []string{"View on GitHub", "Remind Me Later", "Skip This Version"}
	}
	choice, err := zenity.List(
		message,
		options,
		zenity.Title(title),
	)
	if err != nil {
		return ActionRemindLater
	}
	switch choice {
	case "Update Now":
		return ActionUpdateNow
	case "View on GitHub":
		return ActionViewGitHub
	case "Skip This Version":
		return ActionDismiss
	default:
		return ActionRemindLater
	}
}
