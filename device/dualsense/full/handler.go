package dualsensefull

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/dualsense"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/viiperclient"
)

var TypeName = viiperclient.DeviceType[CreateOptions]{
	Name: "dualsense_full",
}

var TypeNameEdge = viiperclient.DeviceType[CreateOptions]{
	Name: "dualsenseedge_full",
}

func init() {
	api.RegisterDevice(TypeName.Name, &handler{
		parent: &dualsense.Handler{},
	})
	api.RegisterDevice(TypeNameEdge.Name, &handler{
		parent: &dualsense.EdgeHandler{},
	})
}

type handler struct {
	parent api.DeviceHandler
}

func (h *handler) CreateDevice(ctx context.Context, o *device.CreateOptions) (usb.Device, error) {
	_, edge := h.parent.(*dualsense.EdgeHandler)
	return new(
		ctx,
		o,
		func(ctx context.Context, o *device.CreateOptions) (*dualsense.DualSense, error) {
			parentDevice, err := h.parent.CreateDevice(ctx, o)
			if err != nil {
				return nil, err
			}
			ds, ok := parentDevice.(*dualsense.DualSense)
			if !ok {
				return nil, fmt.Errorf("%w: expected DualSense", device.ErrWrongDeviceType)
			}
			return ds, nil
		},
		edge,
	)
}

func (h *handler) StreamHandler() api.StreamHandlerFunc {
	return func(conn net.Conn, devPtr *usb.Device, logger *slog.Logger) error {
		if devPtr == nil || *devPtr == nil {
			return fmt.Errorf("nil device")
		}
		d, ok := (*devPtr).(*Device)
		if !ok {
			_, edge := h.parent.(*dualsense.EdgeHandler)
			if edge {
				return fmt.Errorf("%w: expected DualSenseEdge full", device.ErrWrongDeviceType)
			}
			return fmt.Errorf("%w: expected DualSense full", device.ErrWrongDeviceType)
		}
		defer d.ReleaseSerialAndMAC()

		d.SetOutputCallback(func(output OutputState) {
			data, err := output.MarshalBinary()
			if err != nil {
				logger.Error("failed to marshal feedback", "error", err)
				return
			}
			_, err = conn.Write(data)
			if err != nil {
				logger.Error("failed to send feedback", "error", err)
			}
		})

		buf := make([]byte, dualsense.InputStateSize)
		for {
			_, err := io.ReadFull(conn, buf)
			if err == io.EOF {
				logger.Info("client disconnected")
				return nil
			}
			if err != nil {
				return fmt.Errorf("read input state: %w", err)
			}
			var state dualsense.InputState
			err = state.UnmarshalBinary(buf)
			if err != nil {
				return fmt.Errorf("unmarshal input state: %w", err)
			}
			d.UpdateInputState(&state)
		}
	}
}

func (h *handler) UpdateMetaState(meta string, dev *usb.Device) error {
	d, ok := (*dev).(*Device)
	if !ok {
		_, edge := h.parent.(*dualsense.EdgeHandler)
		if edge {
			return fmt.Errorf("%w: expected DualSenseEdge full", device.ErrWrongDeviceType)
		}
		return fmt.Errorf("%w: expected DualSense full", device.ErrWrongDeviceType)
	}
	var parentDevice usb.Device = d.DualSense
	return h.parent.UpdateMetaState(meta, &parentDevice)
}
