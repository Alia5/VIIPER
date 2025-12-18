package api

import (
	"context"
	"log/slog"

	"github.com/Alia5/VIIPER/usbip"
)

func AttachLocalhostClient(ctx context.Context, deviceExportMeta *usbip.ExportMeta, usbipServerPort uint16, useNativeIOCTL bool, logger *slog.Logger) error {
	return attachLocalhostClientImpl(ctx, deviceExportMeta, usbipServerPort, useNativeIOCTL, logger)
}
