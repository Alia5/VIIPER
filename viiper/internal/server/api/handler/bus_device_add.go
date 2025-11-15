package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"viiper/internal/server/api"
	usbs "viiper/internal/server/usb"
	"viiper/pkg/apitypes"
	"viiper/pkg/device"
)

// BusDeviceAdd returns a handler to add devices to a bus.
func BusDeviceAdd(s *usbs.Server, apiSrv *api.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		idStr, ok := req.Params["id"]
		if !ok {
			return fmt.Errorf("missing id")
		}
		busID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return err
		}
		b := s.GetBus(uint32(busID))
		if b == nil {
			return fmt.Errorf("unknown bus")
		}
		if len(req.Args) < 1 {
			return fmt.Errorf("missing device type")
		}
		name := strings.ToLower(req.Args[0])

		reg := api.GetRegistration(name)
		if reg == nil {
			return fmt.Errorf("unknown device type: %s", name)
		}

		dev := reg.CreateDevice()
		devCtx, err := b.Add(dev)
		if err != nil {
			return err
		}

		exportMeta := device.GetDeviceMeta(devCtx)
		if exportMeta == nil {
			return fmt.Errorf("failed to get device metadata from context")
		}

		connTimer := device.GetConnTimer(devCtx)
		if connTimer != nil {
			connTimer.Reset(apiSrv.Config().DeviceHandlerConnectTimeout)
		}
		go func() {
			select {
			case <-devCtx.Done():
				connTimer.Stop()
				return
			case <-connTimer.C:
				deviceIDStr := fmt.Sprintf("%d", exportMeta.DevId)
				if err := b.RemoveDeviceByID(deviceIDStr); err != nil {
					logger.Error("timeout: failed to remove device", "busID", busID, "deviceID", deviceIDStr, "error", err)
				} else {
					logger.Info("timeout: removed device (no connection)", "busID", busID, "deviceID", deviceIDStr)
				}
			}
		}()

		payload, err := json.Marshal(apitypes.DeviceAddResponse{
			ID: fmt.Sprintf("%d-%d", busID, exportMeta.DevId),
		})
		if err != nil {
			return err
		}

		res.JSON = string(payload)
		return nil
	}
}
