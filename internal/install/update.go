package install

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/ncruces/zenity"
)

type Action int

const (
	ActionRemindLater Action = 0
	ActionUpdateNow   Action = 1
	ActionSkipVersion Action = 2
)

const SkippedVersionFile = "update-dismissed"

func CheckUpdate(t *Target, prerelease bool, dialog func(title, message string) Action) (availableVersion string, err error) {
	r, err := latest(*t, prerelease)
	if err != nil {
		return "", err
	}
	available := r.version()
	if !ShouldUpdate(t.Version, available) {
		return "", nil
	}
	if isDismissed(*t, available) || t.Latest == available {
		return available, nil
	}
	slog.Info("update available", "current", t.Version, "available", available)
	switch dialog(
		"VIIPER Update Available",
		fmt.Sprintf("A new version of VIIPER is available: %s", available),
	) {
	case ActionRemindLater:
		t.Latest = available
	case ActionUpdateNow:
		writeDismissed(*t, available)

		return available, update(*t, r.TagName)
	case ActionSkipVersion:
		writeDismissed(*t, available)
	}

	return available, nil
}

func ForgetUpdateChoice(t *Target) {
	t.Latest = ""
	if t.ConfigDir == "" {
		return
	}
	_ = os.Remove(filepath.Join(t.ConfigDir, SkippedVersionFile))
}

func ShowUpdateMessage(title, message string) Action {
	err := zenity.Question(
		message,
		zenity.Title(title),
		zenity.OKLabel("Update now"),
		zenity.CancelLabel("Later"),
		zenity.ExtraButton("Skip version"),
	)
	switch {
	case err == nil:
		return ActionUpdateNow
	case errors.Is(err, zenity.ErrExtraButton):
		return ActionSkipVersion
	default:
		return ActionRemindLater
	}
}

func isDismissed(t Target, version string) bool {
	if t.ConfigDir == "" {
		return false
	}
	content, err := os.ReadFile(filepath.Join(t.ConfigDir, SkippedVersionFile))
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(content)) == version
}

func writeDismissed(t Target, version string) {
	if t.ConfigDir == "" {
		return
	}
	_ = os.MkdirAll(t.ConfigDir, 0o755)
	err := os.WriteFile(filepath.Join(t.ConfigDir, SkippedVersionFile), []byte(version), 0o644)
	if err != nil {
		slog.Error("failed to write update-dismissed", "error", err)
	}
}
