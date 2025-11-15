package handler

import (
	"encoding/json"
	"log/slog"
	"viiper/internal/server/api"
	"viiper/internal/server/usb"
	"viiper/pkg/apitypes"
)

// BusList returns a handler that lists registered busses.
// Error logging is centralized in the API server.
func BusList(s *usb.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		buses := s.ListBuses()
		payload := apitypes.BusListResponse{Buses: buses}
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		res.JSON = string(b)
		return nil
	}
}
