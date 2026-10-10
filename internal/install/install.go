package install

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	latestReleaseURL = "https://api.github.com/repos/Alia5/VIIPER/releases/latest"
	releaseAssetURL  = "https://github.com/Alia5/VIIPER/releases/download/%s/viiper-%s-%s%s"
)

type Target struct {
	Source  string
	Exe     string
	Version string
	Latest  string
}

type step struct {
	label    string
	optional bool
	enabled  bool
	run      func(*Target) error
	err      error
}

//go:embed bundle
var bundle embed.FS

var ErrGoRun = errors.New("cannot install from 'go run'")

func Run(t Target, uninstall bool) error {
	if strings.Contains(t.Source, "go-build") {
		return ErrGoRun
	}
	steps := installSteps
	if uninstall {
		steps = uninstallSteps
	}
	for _, step := range steps {
		if step.optional && !step.enabled {
			continue
		}
		err := step.run(&t)
		if err != nil && step.optional {
			slog.Warn(step.label, "error", err)

			continue
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func checkLatest(t *Target) error {
	resp, err := http.Get(latestReleaseURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	var release struct {
		TagName string `json:"tag_name"`
	}
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return err
	}
	t.Latest = release.TagName

	return nil
}

func downloadLatest(t *Target) error {
	if !isOlder(t.Version, t.Latest) {
		return nil
	}
	download := filepath.Join(os.TempDir(), exeName)
	err := downloadFile(fmt.Sprintf(releaseAssetURL, t.Latest, runtime.GOOS, runtime.GOARCH, filepath.Ext(exeName)), download)
	if err != nil {
		return err
	}
	t.Source = download
	t.Version = t.Latest

	return nil
}

func isOlder(version, than string) bool {
	if version == "" {
		return true
	}
	have := strings.Split(strings.TrimPrefix(version, "v"), ".")
	need := strings.Split(strings.TrimPrefix(than, "v"), ".")
	for i := range need {
		h := 0
		if i < len(have) {
			h, _ = strconv.Atoi(have[i])
		}
		n, _ := strconv.Atoi(need[i])
		if h != n {
			return h < n
		}
	}

	return false
}

func writeFile(path string, content []byte) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return err
	}

	return os.WriteFile(path, content, 0o755)
}

func copyFile(source, path string) error {
	if source == path {
		return nil
	}
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	return writeFile(path, content)
}

func downloadFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return writeFile(path, content)
}
