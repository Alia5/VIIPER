package handler

import (
	"encoding/json"
	"log/slog"

	"github.com/Alia5/VIIPER/apitypes"
	"github.com/Alia5/VIIPER/internal/codegen/common"
	"github.com/Alia5/VIIPER/internal/server/api"
)

// Ping returns a handler for the "ping" endpoint.
// It provides a minimal identity + version response.
func Ping() api.HandlerFunc {
	return func(_ *api.Request, res *api.Response, logger *slog.Logger) error {
		ver, err := common.GetVersion()
		if err != nil {
			ver = common.Version
			if ver == "" {
				ver = "dev"
			}
			logger.Error("ping: invalid version format", "error", err, "version", ver)
		}

		payload := apitypes.PingResponse{Server: "VIIPER", Version: ver}
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		res.JSON = string(b)
		return nil
	}
}
