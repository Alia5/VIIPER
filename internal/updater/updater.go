package updater

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Alia5/VIIPER/internal/config"
	"github.com/Alia5/VIIPER/internal/configpaths"
)

const (
	ActionRemindLater = 0
	ActionViewGitHub  = 1
	ActionUpdateNow   = 2
	ActionDismiss     = 3
)

var (
	client             = &http.Client{Timeout: 10 * time.Second}
	remindLaterVersion string
	versionRe          = regexp.MustCompile(`v(\d+)\.(\d+)\.(\d+)(?:-(\d+)-g[0-9a-f]+)?`)
)

type version struct {
	Major, Minor, Patch, Commits int
}

func parseVersion(s string) (version, bool) {
	m := versionRe.FindStringSubmatch(s)
	if m == nil {
		return version{}, false
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	commits := 0
	if m[4] != "" {
		commits, _ = strconv.Atoi(m[4])
	}
	return version{major, minor, patch, commits}, true
}

func dismissedFilePath() string {
	dir, err := configpaths.DefaultConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "update-dismissed")
}

func isDismissed(ver string) bool {
	p := dismissedFilePath()
	if p == "" {
		return false
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == ver
}

func writeDismissed(ver string) {
	p := dismissedFilePath()
	if p == "" {
		return
	}
	_ = configpaths.EnsureDir(p)
	if err := os.WriteFile(p, []byte(ver), 0o644); err != nil {
		slog.Error("failed to write update-dismissed", "error", err)
	}
}

type release struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease"`
	HTMLURL    string `json:"html_url"`
}

func CheckUpdate(currentVersion string, notify config.UpdateNotify) {
	cur, ok := parseVersion(currentVersion)
	if !ok && currentVersion != "dev" {
		slog.Error("failed to parse current version", "version", currentVersion)
		return
	}

	var r release
	if notify == config.UpdateNotifyPrerelease {
		resp, err := client.Get("https://api.github.com/repos/Alia5/VIIPER/releases?per_page=1")
		if err != nil {
			slog.Error("failed to fetch releases", "error", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			slog.Error("unexpected status from GitHub API", "status", resp.StatusCode)
			return
		}
		var releases []release
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			slog.Error("failed to decode releases", "error", err)
			return
		}
		if len(releases) == 0 {
			return
		}
		r = releases[0]
	} else {
		resp, err := client.Get("https://api.github.com/repos/Alia5/VIIPER/releases/latest")
		if err != nil {
			slog.Error("failed to fetch latest release", "error", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			slog.Error("unexpected status from GitHub API", "status", resp.StatusCode)
			return
		}
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			slog.Error("failed to decode latest release", "error", err)
			return
		}
	}

	versionSource := r.TagName
	if r.Prerelease {
		versionSource = r.Name
	}

	remote, ok := parseVersion(versionSource)
	if !ok {
		slog.Error("failed to parse remote version", "version", versionSource)
		return
	}

	newer := remote.Major > cur.Major ||
		(remote.Major == cur.Major && remote.Minor > cur.Minor) ||
		(remote.Major == cur.Major && remote.Minor == cur.Minor && remote.Patch > cur.Patch) ||
		(remote.Major == cur.Major && remote.Minor == cur.Minor && remote.Patch == cur.Patch && remote.Commits > cur.Commits)

	if !newer {
		return
	}

	matched := versionRe.FindString(versionSource)

	if isDismissed(matched) || remindLaterVersion == matched {
		return
	}

	slog.Info("update available", "current", currentVersion, "available", matched)
	installChannel := "stable"
	if notify == config.UpdateNotifyPrerelease {
		installChannel = "main"
	}

	action := showMessageBox(
		"VIIPER Update Available",
		fmt.Sprintf("A new version of VIIPER is available: %s", matched),
	)

	switch action {
	case ActionRemindLater:
		remindLaterVersion = matched
	case ActionViewGitHub:
		openBrowser(r.HTMLURL)
	case ActionUpdateNow:
		writeDismissed(matched)
		runInstallScript(installChannel)
	case ActionDismiss:
		writeDismissed(matched)
	}
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("cmd", "/c", "start", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		slog.Error("failed to open browser", "error", err)
	}
}

func runInstallScript(channel string) {
	baseURL := "https://alia5.github.io/VIIPER/" + channel + "/install"
	switch runtime.GOOS {
	case "windows":
		url := baseURL + ".ps1"
		cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command",
			"Start-Process powershell -ArgumentList '-NoExit -NoProfile -ExecutionPolicy Bypass -Command \"iwr -useb "+url+" | iex\"' -Verb RunAs")
		if err := cmd.Run(); err != nil {
			slog.Error("failed to run install script", "error", err)
		}
	case "linux":
		cmd := exec.Command("sh", "-c", fmt.Sprintf("curl -fsSL '%s.sh' | sh", baseURL))
		if err := cmd.Start(); err != nil {
			slog.Error("failed to run install script", "error", err)
		}
	}
}
