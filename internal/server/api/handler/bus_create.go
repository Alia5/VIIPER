package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/usb"
	"github.com/Alia5/VIIPER/viiperclient"
	"github.com/Alia5/VIIPER/virtualbus"
)

// BusCreate returns a handler that creates a new bus.
func BusCreate(s *usb.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		if req.Payload != "" {
			busID, err := strconv.ParseUint(req.Payload, 10, 32)
			if err != nil {
				return viiperclient.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err))
			}

			if busID == 0 {
				busID = uint64(s.NextFreeBusID())
			}

			b, err := virtualbus.NewWithBusID(uint32(busID))
			if err != nil {
				return viiperclient.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err))
			}
			if err := s.AddBus(b); err != nil {
				return viiperclient.ErrConflict(fmt.Sprintf("bus %d already exists", busID))
			}
			out, err := json.Marshal(viiperclient.BusCreateResponse{BusID: b.BusID()})
			if err != nil {
				return viiperclient.ErrInternal(fmt.Sprintf("failed to marshal response: %v", err))
			}
			res.JSON = string(out)
			return nil
		}

		busID := s.NextFreeBusID()
		b := virtualbus.New(busID)
		if err := s.AddBus(b); err != nil {
			return viiperclient.ErrInternal(fmt.Sprintf("failed to add bus: %v", err))
		}
		out, err := json.Marshal(viiperclient.BusCreateResponse{BusID: b.BusID()})
		if err != nil {
			return viiperclient.ErrInternal(fmt.Sprintf("failed to marshal response: %v", err))
		}
		res.JSON = string(out)
		return nil
	}
}
