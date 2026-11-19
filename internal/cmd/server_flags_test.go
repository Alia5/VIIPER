package cmd

import (
	"testing"

	"github.com/alecthomas/kong"
)

// Regression test for GitHub issue #30: the documented CLI flag
// --api.require-localhost-auth must be accepted and set the config.
func TestServerFlag_RequireLocalhostAuth(t *testing.T) {
	var s Server
	parser, err := kong.New(&s, kong.Name("viiper"))
	if err != nil {
		t.Fatalf("failed to build kong parser: %v", err)
	}

	if _, err := parser.Parse([]string{"--api.require-localhost-auth=true"}); err != nil {
		t.Fatalf("documented flag --api.require-localhost-auth rejected: %v", err)
	}

	if !s.APIServerConfig.RequireLocalhostAuth {
		t.Fatal("expected RequireLocalhostAuth to be true after parsing --api.require-localhost-auth=true")
	}
}
