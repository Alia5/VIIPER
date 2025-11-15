package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"viiper/internal/server/api"
	"viiper/internal/server/usb"
	"viiper/pkg/apitypes"
)

// BusDevicesList returns a handler that lists devices on a bus.
func BusDevicesList(s *usb.Server) api.HandlerFunc {
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
		metas := b.GetAllDeviceMetas()
		out := make([]apitypes.Device, 0, len(metas))
		for _, m := range metas {
			dtype := inferDeviceType(m.Dev)
			out = append(out, apitypes.Device{
				BusID: m.Meta.BusId,
				DevId: fmt.Sprintf("%d", m.Meta.DevId),
				Vid:   fmt.Sprintf("0x%04x", m.Desc.Device.IDVendor),
				Pid:   fmt.Sprintf("0x%04x", m.Desc.Device.IDProduct),
				Type:  dtype,
			})
		}
		payload, err := json.Marshal(apitypes.DevicesListResponse{Devices: out})
		if err != nil {
			return err
		}
		res.JSON = string(payload)
		return nil
	}
}

// inferDeviceType attempts to derive a friendly device type name from the concrete type.
// For devices under pkg/devices/<name>, we return the last path element (e.g., "xbox360").
// Fallback to the lowercased concrete type name if the package path is unavailable.
func inferDeviceType(dev any) string {
	if dev == nil {
		return ""
	}
	t := reflect.TypeOf(dev)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	pkg := t.PkgPath() // e.g., "viiper/pkg/device/xbox360"
	if pkg != "" {
		base := filepath.Base(pkg)
		if base != "." && base != string(filepath.Separator) {
			return strings.ToLower(base)
		}
	}
	return strings.ToLower(t.Name())
}
