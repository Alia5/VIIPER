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
			return fmt.Errorf("missing id")
		}
		busID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return err
		}
		if len(req.Args) < 1 {
			return fmt.Errorf("missing device number")
		}
		deviceID := req.Args[0]
		if err := s.RemoveDeviceByID(uint32(busID), deviceID); err != nil {
			return err
		}
		j, err := json.Marshal(apitypes.DeviceRemoveResponse{BusID: uint32(busID), DevId: deviceID})
		if err != nil {
			return err
		}
		res.JSON = string(j)
		return nil
	}
}
