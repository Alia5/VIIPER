package handler

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"viiper/internal/server/api"
	"viiper/internal/server/usb"
	"viiper/pkg/apitypes"
	"viiper/pkg/virtualbus"
)

// BusCreate returns a handler that creates a new bus.
// Error logging is centralized in the API server; this handler only returns errors.
func BusCreate(s *usb.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		if len(req.Args) >= 1 {
			busId, err := strconv.ParseUint(req.Args[0], 10, 32)
			if err != nil {
				return err
			}
			b, err := virtualbus.NewWithBusId(uint32(busId))
			if err != nil {
				return err
			}
			if err := s.AddBus(b); err != nil {
				return err
			}
			out, err := json.Marshal(apitypes.BusCreateResponse{BusID: b.BusID()})
			if err != nil {
				return err
			}
			res.JSON = string(out)
			return nil
		}
		b := virtualbus.New()
		if err := s.AddBus(b); err != nil {
			return err
		}
		out, err := json.Marshal(apitypes.BusCreateResponse{BusID: b.BusID()})
		if err != nil {
			return err
		}
		res.JSON = string(out)
		return nil
	}
}
