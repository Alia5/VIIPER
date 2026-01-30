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
	"github.com/Alia5/VIIPER/virtualbus"
)

// BusCreate returns a handler that creates a new bus.
// Error logging is centralized in the API server; this handler only returns errors.
func BusCreate(s *usb.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		if req.Payload != "" {
			busId, err := strconv.ParseUint(req.Payload, 10, 32)
			if err != nil {
				return apierror.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err))
			}

			if busId == 0 {
				busId = uint64(s.NextFreeBusID())
			}

			b, err := virtualbus.NewWithBusId(uint32(busId))
			if err != nil {
				return apierror.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err))
			}
			if err := s.AddBus(b); err != nil {
				return apierror.ErrConflict(fmt.Sprintf("bus %d already exists", busId))
			}
			out, err := json.Marshal(apitypes.BusCreateResponse{BusID: b.BusID()})
			if err != nil {
				return apierror.ErrInternal(fmt.Sprintf("failed to marshal response: %v", err))
			}
			res.JSON = string(out)
			return nil
		}

		busId := s.NextFreeBusID()
		b := virtualbus.New(busId)
		if err := s.AddBus(b); err != nil {
			return apierror.ErrInternal(fmt.Sprintf("failed to add bus: %v", err))
		}
		out, err := json.Marshal(apitypes.BusCreateResponse{BusID: b.BusID()})
		if err != nil {
			return apierror.ErrInternal(fmt.Sprintf("failed to marshal response: %v", err))
		}
		res.JSON = string(out)
		return nil
	}
}
