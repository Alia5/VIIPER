package handler

import (
	"encoding/json"
	"log/slog"

	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/usb"
	"github.com/Alia5/VIIPER/viiperclient"
)

// BusList returns a handler that lists registered busses.
// Error logging is centralized in the API server.
func BusList(s *usb.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		buses := s.ListBuses()
		payload := viiperclient.BusListResponse{Buses: buses}
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		res.JSON = string(b)
		return nil
	}
}
