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
			return api.ErrBadRequest("missing id parameter")
		}
		busID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return api.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err))
		}
		b := s.GetBus(uint32(busID))
		if b == nil {
			return api.ErrNotFound(fmt.Sprintf("bus %d not found", busID))
		}
		if req.Payload == "" {
			return api.ErrBadRequest("missing payload")
		}
		var deviceCreateReq apitypes.DeviceCreateRequest
		err = json.Unmarshal([]byte(req.Payload), &deviceCreateReq)
		if err != nil {
			return api.ErrBadRequest(fmt.Sprintf("invalid JSON payload: %v", err))
		}
		if deviceCreateReq.Type == nil {
			return api.ErrBadRequest("missing device type")
		}

		name := strings.ToLower(*deviceCreateReq.Type)

		reg := api.GetRegistration(name)
		if reg == nil {
			return api.ErrBadRequest(fmt.Sprintf("unknown device type: %s", name))
		}

		opts := device.CreateOptions{
			IdVendor:  deviceCreateReq.IdVendor,
			IdProduct: deviceCreateReq.IdProduct,
		}

		dev := reg.CreateDevice(&opts)
		devCtx, err := b.Add(dev)
		if err != nil {
			return api.ErrInternal(fmt.Sprintf("failed to add device to bus: %v", err))
		}

		exportMeta := device.GetDeviceMeta(devCtx)
		if exportMeta == nil {
			return api.ErrInternal("failed to get device metadata from context")
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

		if apiSrv.Config().AutoAttachLocalClient {
			err := api.AttachLocalhostClient(
				req.Ctx,
				exportMeta,
				s.GetListenPort(),
				logger,
			)
			if err != nil {
				logger.Error("failed to auto-attach localhost client", "error", err)
			}
		}

		payload, err := json.Marshal(apitypes.Device{
			BusID: uint32(busID),
			DevId: fmt.Sprintf("%d", exportMeta.DevId),
			Vid:   fmt.Sprintf("0x%04x", dev.GetDescriptor().Device.IDVendor),
			Pid:   fmt.Sprintf("0x%04x", dev.GetDescriptor().Device.IDProduct),
			Type:  name,
		})
		if err != nil {
			return api.ErrInternal(fmt.Sprintf("failed to marshal response: %v", err))
		}

		res.JSON = string(payload)
		return nil
	}
}
