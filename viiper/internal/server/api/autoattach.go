package api

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"viiper/pkg/usbip"
)

func AttachLocalhostClient(ctx context.Context, deviceExportMeta *usbip.ExportMeta, usbipServerPort uint16, logger *slog.Logger) error {

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
