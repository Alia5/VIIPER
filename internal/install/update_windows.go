package install

import (
	"os/exec"
	"syscall"
)

func update(t Target, tag string) error {
	exe, err := download(t, tag)
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "install", "--quiet")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	return t.Launch(cmd)
}
