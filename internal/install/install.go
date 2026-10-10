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
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	GitHubAPI    = "https://api.github.com/repos/Alia5/VIIPER"
	GitHubAssets = "https://github.com/Alia5/VIIPER/releases/download"
	apiTimeout   = 10 * time.Second
)

type Target struct {
	Source    string
	Exe       string
	Version   string
	Latest    string
	API       string
	Assets    string
	TempDir   string
	ConfigDir string
	Launch    func(*exec.Cmd) error
}

type step struct {
	label    string
	optional bool
	enabled  bool
	run      func(*Target) error
	err      error
}

type release struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease"`
}

//go:embed bundle
var bundle embed.FS

var (
	ErrGoRun     = errors.New("cannot install from 'go run'")
	ErrNoRelease = errors.New("no release found")

	client = &http.Client{
		Timeout: apiTimeout,
	}
	versionRe = regexp.MustCompile(`v?(\d+(?:\.\d+)*)(?:-(\d+)-g[0-9a-f]+)?`)
)

func (r release) version() string {
	if r.Prerelease {
		return r.Name
	}

	return r.TagName
}

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
	r, err := latest(*t, false)
	if err != nil {
		return err
	}
	t.Latest = r.TagName

	return nil
}

func downloadLatest(t *Target) error {
	if !ShouldUpdate(t.Version, t.Latest) {
		return nil
	}
	exe, err := download(*t, t.Latest)
	if err != nil {
		return err
	}
	t.Source = exe
	t.Version = t.Latest

	return nil
}

func latest(t Target, prerelease bool) (release, error) {
	url := t.API + "/releases/latest"
	if prerelease {
		url = t.API + "/releases?per_page=1"
	}
	resp, err := client.Get(url)
	if err != nil {
		return release{}, err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return release{}, fmt.Errorf("%s: status %d", url, resp.StatusCode)
	}
	if !prerelease {
		var r release
		err = json.NewDecoder(resp.Body).Decode(&r)

		return r, err
	}
	var releases []release
	err = json.NewDecoder(resp.Body).Decode(&releases)
	if err != nil {
		return release{}, err
	}
	if len(releases) == 0 {
		return release{}, ErrNoRelease
	}

	return releases[0], nil
}

func download(t Target, tag string) (string, error) {
	exe := filepath.Join(t.TempDir, exeName)

	return exe, downloadFile(fmt.Sprintf("%s/%s/viiper-%s-%s%s", t.Assets, tag, runtime.GOOS, runtime.GOARCH, filepath.Ext(exeName)), exe)
}

func parseVersion(s string) []int {
	m := versionRe.FindStringSubmatch(s)
	if m == nil {
		return nil
	}
	var parts []int
	for p := range strings.SplitSeq(m[1], ".") {
		n, _ := strconv.Atoi(p)
		parts = append(parts, n)
	}
	if m[2] != "" {
		n, _ := strconv.Atoi(m[2])
		parts = append(parts, n)
	}

	return parts
}

func ShouldUpdate(version, latest string) bool {
	if version == "" {
		return true
	}
	current := parseVersion(version)
	if current == nil {
		return false
	}
	available := parseVersion(latest)
	for i := range max(len(current), len(available)) {
		currentPart := 0
		if i < len(current) {
			currentPart = current[i]
		}
		availablePart := 0
		if i < len(available) {
			availablePart = available[i]
		}
		if currentPart != availablePart {
			return currentPart < availablePart
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
