package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/usb"
	"github.com/Alia5/VIIPER/viiperclient"
)

// BusRemove returns a handler that removes a bus.
func BusRemove(s *usb.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		if req.Payload == "" {
			return viiperclient.ErrBadRequest("missing busId")
		}
		busID, err := strconv.ParseUint(req.Payload, 10, 32)
		if err != nil {
			return viiperclient.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err))
		}
		if err := s.RemoveBus(uint32(busID)); err != nil {
			return viiperclient.ErrNotFound(fmt.Sprintf("bus %d not found", busID))
		}
		out, err := json.Marshal(viiperclient.BusRemoveResponse{BusID: uint32(busID)})
		if err != nil {
			return viiperclient.ErrInternal(fmt.Sprintf("failed to marshal response: %v", err))
		}
		res.JSON = string(out)
		return nil
	}
}
