package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"viiper/internal/server/api"
	"viiper/internal/server/usb"
	"viiper/pkg/apitypes"
)

// BusRemove returns a handler that removes a bus.
func BusRemove(s *usb.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		if len(req.Args) < 1 {
			return fmt.Errorf("missing busId")
		}
		busID, err := strconv.ParseUint(req.Args[0], 10, 32)
		if err != nil {
			return err
		}
		if err := s.RemoveBus(uint32(busID)); err != nil {
			return err
		}
		out, err := json.Marshal(apitypes.BusRemoveResponse{BusID: uint32(busID)})
		if err != nil {
			return err
		}
		res.JSON = string(out)
		return nil
	}
}
