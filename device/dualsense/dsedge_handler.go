package dualsense

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/usb/hid"
	"github.com/Alia5/VIIPER/viiperclient"
)

const (
	OutputReportCountEdge         = 63
	FeatureF2CountEdge            = 52
	ReportCountOffsetFromReportID = 2
)

var EdgeOnlyFeatureReports = []hid.Item{
	hid.ReportID{ID: 0x60}, hid.Usage{Usage: 0x41}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x61}, hid.Usage{Usage: 0x42}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x62}, hid.Usage{Usage: 0x43}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x63}, hid.Usage{Usage: 0x44}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x64}, hid.Usage{Usage: 0x45}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x65}, hid.Usage{Usage: 0x46}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x68}, hid.Usage{Usage: 0x47}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x70}, hid.Usage{Usage: 0x48}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x71}, hid.Usage{Usage: 0x49}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x72}, hid.Usage{Usage: 0x4A}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x73}, hid.Usage{Usage: 0x4B}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x74}, hid.Usage{Usage: 0x4C}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x75}, hid.Usage{Usage: 0x4D}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x76}, hid.Usage{Usage: 0x4E}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x77}, hid.Usage{Usage: 0x4F}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x78}, hid.Usage{Usage: 0x50}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x79}, hid.Usage{Usage: 0x51}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x7A}, hid.Usage{Usage: 0x52}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
	hid.ReportID{ID: 0x7B}, hid.Usage{Usage: 0x53}, hid.ReportCount{Count: 63}, hid.Feature{Flags: hid.MainData | hid.MainVar | hid.MainAbs},
}

var TypeNameEdge = viiperclient.DeviceType[MetaState]{
	Name: "dualsenseedge",
}

func init() {
	api.RegisterDevice(TypeNameEdge.Name, &dsedgehandler{})
}

type dsedgehandler struct{}

func (h *EdgeHandler) CreateDevice(ctx context.Context, o *device.CreateOptions) (usb.Device, error) {
	if o == nil {
		o = &device.CreateOptions{}
	}

	metaState := MetaState{
		ShellColor: DefaultShellColor,
	}
	if o.DeviceSpecific != "" {
		if err := json.Unmarshal([]byte(o.DeviceSpecific), &metaState); err != nil {
			return nil, fmt.Errorf("invalid device specific JSON: %w", err)
		}
	}

	serial := metaState.SerialNumber
	if serial == "" {
		serial = DefaultSerialNumberDSEdge
	}
	if metaState.ShellColor != "" && len(serial) >= 6 {
		code := strings.ToUpper(metaState.ShellColor)
		if len(code) >= 2 {
			serial = serial[:4] + code[:2] + serial[6:]
		}
	}
	if _, ok := serials[serial]; ok {
		for i := 1; i < 16; i++ {
			newSerial := fmt.Sprintf("%s%02X", serial[:len(serial)-2], i)
			if _, exists := serials[newSerial]; !exists {
				serial = newSerial
				break
			}
		}
	}
	metaState.SerialNumber = serial
	serials[serial] = struct{}{}

	mac := metaState.MACAddress
	if mac == "" {
		mac = DefaultMACAddressDSEdge
	}
	if _, ok := macs[mac]; ok {
		prefix := mac[:len(mac)-2]
		for i := 1; i <= 16; i++ {
			candidate := fmt.Sprintf("%s%02X", prefix, i)
			if _, exists := macs[candidate]; !exists {
				mac = candidate
				break
			}
		}
	}
	metaState.MACAddress = mac
	macs[mac] = struct{}{}

	b, err := json.Marshal(metaState)
	if err != nil {
		return nil, fmt.Errorf("marshal meta state: %w", err)
	}
	o.DeviceSpecific = string(b)

	return new(ctx, o, true)
}

func (h *dsedgehandler) StreamHandler() api.StreamHandlerFunc {
	return func(conn net.Conn, devPtr *usb.Device, logger *slog.Logger) error {
		defer func() {
			if devPtr == nil || *devPtr == nil {
				return
			}
			dse, ok := (*devPtr).(*DualSense)
			if !ok {
				slog.Warn("device is not DualSenseEdge on disconnect")
				return
			}
			dse.mtx.Lock()
			serial := dse.metaState.SerialNumber
			mac := dse.metaState.MACAddress
			dse.mtx.Unlock()
			delete(serials, serial)
			delete(macs, mac)
			slog.Debug("DualSenseEdge disconnected, serial/mac released", "serial", serial, "mac", mac)
		}()

		if devPtr == nil || *devPtr == nil {
			return fmt.Errorf("nil device")
		}
		dse, ok := (*devPtr).(*DualSense)
		if !ok {
			return fmt.Errorf("%w: expected DualSenseEdge", device.ErrWrongDeviceType)
		}

		dse.SetOutputCallback(func(feedback OutputState) {
			data, err := feedback.MarshalBinary()
			if err != nil {
				logger.Error("failed to marshal feedback", "error", err)
				return
			}
			if _, err := conn.Write(data); err != nil {
				logger.Error("failed to send feedback", "error", err)
			}
		})

		buf := make([]byte, InputStateSize)
		for {
			if _, err := io.ReadFull(conn, buf); err != nil {
				if err == io.EOF {
					logger.Info("client disconnected")
					return nil
				}
				return fmt.Errorf("read input state: %w", err)
			}

			var state InputState
			if err := state.UnmarshalBinary(buf); err != nil {
				return fmt.Errorf("unmarshal input state: %w", err)
			}
			dse.UpdateInputState(&state)
		}
	}
}

func (h *dsedgehandler) UpdateMetaState(meta string, dev *usb.Device) error {
	dse, ok := (*dev).(*DualSense)
	if !ok {
		return fmt.Errorf("%w: expected DualSenseEdge", device.ErrWrongDeviceType)
	}
	dse.mtx.Lock()
	current := *dse.metaState
	dse.mtx.Unlock()
	if err := json.Unmarshal([]byte(meta), &current); err != nil {
		return fmt.Errorf("unmarshal meta state: %w", err)
	}
	dse.SetMetaState(current)
	return nil
}
