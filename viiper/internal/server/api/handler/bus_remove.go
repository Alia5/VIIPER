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
		if req.Payload == "" {
			return api.ErrBadRequest("missing busId")
		}
		busID, err := strconv.ParseUint(req.Payload, 10, 32)
		if err != nil {
			return api.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err))
		}
		if err := s.RemoveBus(uint32(busID)); err != nil {
			return api.ErrNotFound(fmt.Sprintf("bus %d not found", busID))
		}
		out, err := json.Marshal(apitypes.BusRemoveResponse{BusID: uint32(busID)})
		if err != nil {
			return api.ErrInternal(fmt.Sprintf("failed to marshal response: %v", err))
		}
		res.JSON = string(out)
		return nil
	}
}
