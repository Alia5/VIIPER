package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/Alia5/VIIPER/apitypes"
	"github.com/Alia5/VIIPER/internal/server/api"
	apierror "github.com/Alia5/VIIPER/internal/server/api/error"
	"github.com/Alia5/VIIPER/internal/server/usb"
)

// BusRemove returns a handler that removes a bus.
func BusRemove(s *usb.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		if req.Payload == "" {
			return apierror.ErrBadRequest("missing busId")
		}
		busID, err := strconv.ParseUint(req.Payload, 10, 32)
		if err != nil {
			return apierror.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err))
		}
		if err := s.RemoveBus(uint32(busID)); err != nil {
			return apierror.ErrNotFound(fmt.Sprintf("bus %d not found", busID))
		}
		out, err := json.Marshal(apitypes.BusRemoveResponse{BusID: uint32(busID)})
		if err != nil {
			return apierror.ErrInternal(fmt.Sprintf("failed to marshal response: %v", err))
		}
		res.JSON = string(out)
		return nil
	}
}
