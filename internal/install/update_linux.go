package install

import (
	"os"
	"os/exec"
)

func update(t Target, tag string) error {
	exe, err := download(t, tag)
	if err != nil {
		return err
	}
	if os.Geteuid() != 0 {
		return t.Launch(exec.Command("pkexec", exe, "install", "--quiet"))
	}

	return t.Launch(exec.Command("systemd-run", "--collect", exe, "install", "--quiet"))
}
