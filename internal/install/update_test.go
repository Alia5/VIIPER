package install_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Alia5/VIIPER/internal/config"
	"github.com/Alia5/VIIPER/internal/install"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCheck(t *testing.T) {
	type testCase struct {
		name            string
		version         string
		comparedAgainst string
		expectUpdate    bool
	}

	cases := []testCase{
		{
			name:            "None vs Stable",
			version:         "",
			comparedAgainst: "v0.0.1",
			expectUpdate:    true,
		},
		{
			name:            "None vs Prerelease",
			version:         "",
			comparedAgainst: "v0.0.1-6-gabcdef0",
			expectUpdate:    true,
		},
		{
			name:            "Same (Stable)",
			version:         "v0.0.1",
			comparedAgainst: "v0.0.1",
			expectUpdate:    false,
		},
		{
			name:            "Same (Prerelease)",
			version:         "v0.0.1-6-gabcdef0",
			comparedAgainst: "v0.0.1-6-gabcdef0",
			expectUpdate:    false,
		},
		{
			name:            "Older Stable vs Stable",
			version:         "v0.0.1",
			comparedAgainst: "v0.0.2",
			expectUpdate:    true,
		},
		{
			name:            "Newer Stable vs Stable",
			version:         "v0.1.0",
			comparedAgainst: "v0.0.2",
			expectUpdate:    false,
		},
		{
			name:            "Stable vs its Prerelease",
			version:         "v0.0.1",
			comparedAgainst: "v0.0.1-6-gabcdef0",
			expectUpdate:    true,
		},
		{
			name:            "Prerelease vs its Stable",
			version:         "v0.0.1-6-gabcdef0",
			comparedAgainst: "v0.0.1",
			expectUpdate:    false,
		},
		{
			name:            "Older Prerelease vs Prerelease",
			version:         "v0.0.1-6-gabcdef0",
			comparedAgainst: "v0.0.1-8-g0123456",
			expectUpdate:    true,
		},
		{
			name:            "Prerelease vs next Stable",
			version:         "v0.0.1-8-g0123456",
			comparedAgainst: "v0.0.2",
			expectUpdate:    true,
		},
		{
			name:            "Stable vs Prerelease name",
			version:         "v0.0.1",
			comparedAgainst: "Dev Build v0.0.1-6-gabcdef0",
			expectUpdate:    true,
		},
		{
			name:            "Stable vs tag without version",
			version:         "v0.0.1",
			comparedAgainst: "dev-snapshot",
			expectUpdate:    false,
		},
		{
			name:            "Dev vs Stable",
			version:         "dev",
			comparedAgainst: "v999.0.0",
			expectUpdate:    false,
		},
		{
			name:            "Three vs four components",
			version:         "0.9.8",
			comparedAgainst: "0.9.8.1",
			expectUpdate:    true,
		},
		{
			name:            "Four vs three components",
			version:         "0.9.8.1",
			comparedAgainst: "0.9.8",
			expectUpdate:    false,
		},
	}

	for _, tc := range cases {
		shouldUpdate := install.ShouldUpdate(tc.version, tc.comparedAgainst)
		assert.Equal(t, tc.expectUpdate, shouldUpdate)
	}

}

func TestUpdateDialog(t *testing.T) {
	type testCase struct {
		name          string
		updateChannel config.UpdateNotify
		ownVersion    string
		cloudVersion  string
		setup         func(t *testing.T, target *install.Target)

		err          error
		expectDialog bool
		dialogChoice install.Action
		assert       func(t *testing.T, target *install.Target, updateLaunched bool)
	}

	cases := map[string]testCase{
		"CloudSameAsOwn": {
			name:         "ExpectNoDialog",
			ownVersion:   "v0.0.1",
			cloudVersion: "v0.0.1",
			expectDialog: false,
		},
		"CloudNewerThanOwn ChoiceRemindLater": {
			name:         "ExpectDialog ExpectVersionReminded",
			ownVersion:   "v0.0.1",
			cloudVersion: "v0.0.2",
			setup: func(_ *testing.T, target *install.Target) {
				target.Latest = "v0.0.2"
			},
			expectDialog: true,
			dialogChoice: install.ActionRemindLater,
			assert: func(t *testing.T, target *install.Target, _ bool) {
				assert.Equal(t, "v0.0.2", target.Latest)
			},
		},
		"CloudNewerThanOwn ChoiceSkipVersion": {
			name:         "ExpectDialog ExpectVersionSkipped",
			ownVersion:   "v0.0.1",
			cloudVersion: "v0.0.2",
			setup: func(t *testing.T, target *install.Target) {
				require.NoError(t, os.WriteFile(filepath.Join(target.ConfigDir, install.SkippedVersionFile), []byte("v0.0.2"), 0o644))
			},
			expectDialog: true,
			dialogChoice: install.ActionSkipVersion,
			assert: func(t *testing.T, target *install.Target, _ bool) {
				skippedVersion, err := os.ReadFile(filepath.Join(target.ConfigDir, install.SkippedVersionFile))
				require.NoError(t, err)
				assert.Equal(t, "v0.0.2", string(skippedVersion))
			},
		},
		"CloudNewerThanOwn ChoiceUpdateNow": {
			name:         "ExpectDialog ExpectUpdateLaunched",
			ownVersion:   "v0.0.1",
			cloudVersion: "v0.0.2",
			expectDialog: true,
			dialogChoice: install.ActionUpdateNow,
			assert: func(t *testing.T, _ *install.Target, updateLaunched bool) {
				assert.True(t, updateLaunched)
			},
		},
	}

	for channelName, updateChannel := range map[string]config.UpdateNotify{
		"ChannelStable":     config.UpdateNotifyStable,
		"ChannelPrerelease": config.UpdateNotifyPrerelease,
	} {
		t.Run(channelName, func(t *testing.T) {
			for baselineName, baseline := range cases {
				t.Run(baselineName, func(t *testing.T) {
					noPreviousChoice := baseline
					noPreviousChoice.name = "NoPreviousChoice " + baseline.name
					noPreviousChoice.setup = nil
					filledInCases := []testCase{noPreviousChoice}
					if baseline.setup != nil {
						previousChoiceOnSameCloud := baseline
						previousChoiceOnSameCloud.name = "PreviousChoiceOnSameCloud ExpectNoDialog"
						previousChoiceOnSameCloud.expectDialog = false
						previousChoiceOnSameCloud.assert = nil

						previousChoiceOnOlderCloud := baseline
						previousChoiceOnOlderCloud.name = "PreviousChoiceOnOlderCloud ExpectDialog"
						previousChoiceOnOlderCloud.cloudVersion = "v0.0.3"
						previousChoiceOnOlderCloud.expectDialog = true
						previousChoiceOnOlderCloud.assert = nil

						filledInCases = append(filledInCases, previousChoiceOnSameCloud, previousChoiceOnOlderCloud)
					}

					for _, tc := range filledInCases {
						tc.updateChannel = updateChannel
						t.Run(tc.name, func(t *testing.T) {
							releaseTag := tc.cloudVersion
							releaseRoute := "GET /releases/latest"
							releaseJSON := fmt.Sprintf(`{"tag_name":%q}`, releaseTag)
							if tc.updateChannel == config.UpdateNotifyPrerelease {
								// snapshots.yml: tag_name dev-snapshot, version in name
								releaseTag = "dev-snapshot"
								releaseRoute = "GET /releases"
								releaseJSON = fmt.Sprintf(`[{"tag_name":%q,"name":%q,"prerelease":true}]`, releaseTag, tc.cloudVersion)
							}
							github := http.NewServeMux()
							github.HandleFunc(releaseRoute, func(w http.ResponseWriter, _ *http.Request) {
								fmt.Fprint(w, releaseJSON)
							})
							github.HandleFunc("GET /download/"+releaseTag+"/", func(http.ResponseWriter, *http.Request) {})
							githubServer := httptest.NewServer(github)
							defer githubServer.Close()

							updateLaunched := false
							target := &install.Target{
								Version:   tc.ownVersion,
								API:       githubServer.URL,
								Assets:    githubServer.URL + "/download",
								TempDir:   t.TempDir(),
								ConfigDir: t.TempDir(),
								Launch: func(*exec.Cmd) error {
									updateLaunched = true

									return nil
								},
							}
							if tc.setup != nil {
								tc.setup(t, target)
							}

							dialogShown := false
							dialogMock := func(_, _ string) install.Action {
								dialogShown = true
								return tc.dialogChoice
							}
							_, err := install.CheckUpdate(
								target,
								tc.updateChannel == config.UpdateNotifyPrerelease,
								dialogMock,
							)
							assert.Equal(t, tc.expectDialog, dialogShown)
							assert.Equal(t, tc.err, err)

							if tc.assert != nil {
								tc.assert(t, target, updateLaunched)
							}
						})
					}
				})
			}
		})
	}
}

// func TestReal(t *testing.T) {

// 	target := install.Target{
// 		Source:    "viiper",
// 		Version:   "v0.0.1",
// 		API:       install.GitHubAPI,
// 		Assets:    install.GitHubAssets,
// 		TempDir:   os.TempDir(),
// 		ConfigDir: os.TempDir(),
// 		Launch: func(*exec.Cmd) error {
// 			slog.Debug("mock")
// 			return nil
// 		},
// 	}
// 	err := install.CheckUpdate(
// 		&target,
// 		false,
// 		install.ShowUpdateMessage,
// 	)
// 	assert.NoError(t, err)
// }
