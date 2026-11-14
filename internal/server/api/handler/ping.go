package handler

import (
	"encoding/json"
	"log/slog"

	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/meta"
	"github.com/Alia5/VIIPER/viiperclient"
)

// Ping returns a handler for the "ping" endpoint.
// It provides a minimal identity + version response.
func Ping() api.HandlerFunc {
	return func(_ *api.Request, res *api.Response, logger *slog.Logger) error {
		ver := meta.Version
		if ver == "" {
			ver = "dev"
		}

		payload := viiperclient.PingResponse{
			Server:  "VIIPER",
			Version: ver,
		}
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		res.JSON = string(b)
		return nil
	}
}
