package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/internal/server/api"
	apierror "github.com/Alia5/VIIPER/internal/server/api/error"
	usbs "github.com/Alia5/VIIPER/internal/server/usb"
	"github.com/Alia5/VIIPER/viipertypes"
)

// BusDeviceAdd returns a handler to add devices to a bus.
func BusDeviceAdd(s *usbs.Server, apiSrv *api.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		idStr, ok := req.Params["id"]
		if !ok {
			return apierror.ErrBadRequest("missing id parameter")
		}
		busID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return apierror.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err))
		}
		b := s.GetBus(uint32(busID))
		if b == nil {
			return apierror.ErrNotFound(fmt.Sprintf("bus %d not found", busID))
		}
		if req.Payload == "" {
			return apierror.ErrBadRequest("missing payload")
		}
		var deviceCreateReq viipertypes.DeviceCreateRequest
		err = json.Unmarshal([]byte(req.Payload), &deviceCreateReq)
		if err != nil {
			return apierror.ErrBadRequest(fmt.Sprintf("invalid JSON payload: %v", err))
		}
		if deviceCreateReq.Type == nil {
			return apierror.ErrBadRequest("missing device type")
		}

		name := strings.ToLower(*deviceCreateReq.Type)

		reg := api.GetRegistration(name)
		if reg == nil {
			return apierror.ErrBadRequest(fmt.Sprintf("unknown device type: %s", name))
		}

		opts := device.CreateOptions{
			IDVendor:  deviceCreateReq.IDVendor,
			IDProduct: deviceCreateReq.IDProduct,
		}
		if deviceCreateReq.DeviceSpecific != nil {
			b, err := json.Marshal(deviceCreateReq.DeviceSpecific)
			if err != nil {
				return apierror.ErrBadRequest(fmt.Sprintf("invalid deviceSpecific JSON: %v", err))
			}
			opts.DeviceSpecific = string(b)
		}

		dev, err := reg.CreateDevice(&opts)
		if err != nil {
			return apierror.ErrBadRequest(fmt.Sprintf("failed to create device: %v", err))
		}
		devCtx, err := b.Add(dev)
		if err != nil {
			return apierror.ErrInternal(fmt.Sprintf("failed to add device to bus: %v", err))
		}

		exportMeta := device.GetDeviceMeta(devCtx)
		if exportMeta == nil {
			return apierror.ErrInternal("failed to get device metadata from context")
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
				deviceIDStr := fmt.Sprintf("%d", exportMeta.DevID)
				if err := s.RemoveDeviceByID(uint32(busID), deviceIDStr); err != nil {
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
				apiSrv.Config().AutoAttachWindowsNative,
				logger,
			)
			if err != nil {
				logger.Error("failed to auto-attach localhost client", "error", err)
				return apierror.ErrConflict(fmt.Sprintf(
					"Failed to auto-attach device: %v", err,
				))
			}
		}

		payload, err := json.Marshal(viipertypes.Device{
			BusID:          uint32(busID),
			DevID:          fmt.Sprintf("%d", exportMeta.DevID),
			Vid:            fmt.Sprintf("0x%04x", dev.GetDescriptor().Device.IDVendor),
			Pid:            fmt.Sprintf("0x%04x", dev.GetDescriptor().Device.IDProduct),
			Type:           name,
			DeviceSpecific: dev.GetDeviceSpecificArgs(),
		})
		if err != nil {
			return apierror.ErrInternal(fmt.Sprintf("failed to marshal response: %v", err))
		}

		res.JSON = string(payload)
		return nil
	}
}
