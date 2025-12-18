//go:build linux

package api

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"

	"github.com/Alia5/VIIPER/usbip"
)

func attachLocalhostClientImpl(ctx context.Context, deviceExportMeta *usbip.ExportMeta, usbipServerPort uint16, _ bool, logger *slog.Logger) error {
	logger.Info("Auto-attaching localhost client", "busID", deviceExportMeta.BusId, "deviceID", deviceExportMeta.DevId)

	cmd := exec.CommandContext(
		ctx,
		"usbip",
		"--tcp-port",
		strconv.FormatUint(uint64(usbipServerPort), 10),
		"attach",
		"-r", "localhost",
		"-b", fmt.Sprintf("%d-%d", deviceExportMeta.BusId, deviceExportMeta.DevId),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to attach device",
			"error", err,
			"port", usbipServerPort,
			"output", string(output))
		return err
	}
	logger.Debug("usbip attach output", "output", string(output))

	return nil
}

// CheckAutoAttachPrerequisites checks if auto-attach prerequisites are met on Linux.
// Returns true if all requirements are satisfied, false otherwise with helpful log messages.
func CheckAutoAttachPrerequisites(_ bool, logger *slog.Logger) bool {
	allOk := true

	if _, err := exec.LookPath("usbip"); err != nil {
		logger.Warn("USB/IP tool 'usbip' not found in PATH")
		logger.Warn("Auto-attach requires the usbip command-line tool")
		logger.Info("Install usbip:")
		logger.Info("  Ubuntu/Debian: sudo apt install linux-tools-generic")
		logger.Info("  Arch Linux:    sudo pacman -S usbip")
		allOk = false
	} else {
		logger.Debug("usbip tool found in PATH")
	}

	data, err := os.ReadFile("/proc/modules")
	if err != nil {
		logger.Debug("Could not read /proc/modules", "error", err)
		// dont fail, try anyway
	} else if !bytes.Contains(data, []byte("vhci_hcd")) {
		logger.Warn("USB/IP kernel module 'vhci-hcd' is not loaded")
		logger.Warn("Auto-attach will not work until the module is loaded")
		logger.Info("To load the module now, run in another terminal:")
		logger.Info("  sudo modprobe vhci-hcd")
		logger.Info("")
		logger.Info("To automatically load at boot:")
		logger.Info("  echo 'vhci-hcd' | sudo tee /etc/modules-load.d/viiper.conf")
		allOk = false
	} else {
		logger.Debug("vhci-hcd kernel module is loaded")
	}

	return allOk
}
