package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/usb"
	"github.com/Alia5/VIIPER/viiperclient"
)

// BusDevicesList returns a handler that lists devices on a bus.
func BusDevicesList(s *usb.Server) api.HandlerFunc {
	return func(req *api.Request, res *api.Response, logger *slog.Logger) error {
		idStr, ok := req.Params["id"]
		if !ok {
			return viiperclient.ErrBadRequest("missing id parameter")
		}
		busID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return viiperclient.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err))
		}
		b := s.GetBus(uint32(busID))
		if b == nil {
			return viiperclient.ErrNotFound(fmt.Sprintf("bus %d not found", busID))
		}
		metas := b.GetAllDeviceMetas()
		out := make([]viiperclient.Device, 0, len(metas))
		for _, m := range metas {
			dtype := inferDeviceType(m.Dev)
			out = append(out, viiperclient.Device{
				BusID:          m.Meta.BusID,
				DevID:          fmt.Sprintf("%d", m.Meta.DevID),
				Vid:            fmt.Sprintf("0x%04x", m.Dev.GetDescriptor().Device.IDVendor),
				Pid:            fmt.Sprintf("0x%04x", m.Dev.GetDescriptor().Device.IDProduct),
				Type:           dtype,
				DeviceSpecific: m.Dev.GetDeviceSpecificArgs(),
			})
		}
		payload, err := json.Marshal(viiperclient.DevicesListResponse{Devices: out})
		if err != nil {
			return viiperclient.ErrInternal(fmt.Sprintf("failed to marshal response: %v", err))
		}
		res.JSON = string(payload)
		return nil
	}
}

// inferDeviceType attempts to derive a friendly device type name from the concrete type.
// For devices under /devices/<name>, we return the last path element (e.g., "xbox360").
// Fallback to the lowercased concrete type name if the package path is unavailable.
func inferDeviceType(dev any) string {
	if dev == nil {
		return ""
	}
	t := reflect.TypeOf(dev)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	pkg := t.PkgPath() // e.g., "github.com/Alia5/VIIPER/device/xbox360"
	if pkg != "" {
		base := filepath.Base(pkg)
		if base != "." && base != string(filepath.Separator) {
			return strings.ToLower(base)
		}
	}
	return strings.ToLower(t.Name())
}
