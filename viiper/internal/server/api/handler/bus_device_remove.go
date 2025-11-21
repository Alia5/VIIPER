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

// BusDeviceRemove returns a handler that removes a device by device number.
func BusDeviceRemove(s *usb.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		idStr, ok := req.Params["id"]
		if !ok {
			return api.ErrBadRequest("missing id parameter")
		}
		busID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return api.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err))
		}
		if req.Payload == "" {
			return api.ErrBadRequest("missing device number")
		}
		deviceID := req.Payload

		b := s.GetBus(uint32(busID))
		if b == nil {
			return api.ErrNotFound(fmt.Sprintf("bus %d not found", busID))
		}
		if err := b.RemoveDeviceByID(deviceID); err != nil {
			return api.ErrNotFound(fmt.Sprintf("device %s not found on bus %d", deviceID, busID))
		}

		j, err := json.Marshal(apitypes.DeviceRemoveResponse{BusID: uint32(busID), DevId: deviceID})
		if err != nil {
			return api.ErrInternal(fmt.Sprintf("failed to marshal response: %v", err))
		}
		res.JSON = string(j)
		return nil
	}
}
