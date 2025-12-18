package xbox360

import (
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/usb"
)

func init() {
	api.RegisterDevice("xbox360", &handler{})
}

type handler struct{}

func (h *handler) CreateDevice(o *device.CreateOptions) usb.Device { return New(o) }

func (r *handler) StreamHandler() api.StreamHandlerFunc {
	return func(conn net.Conn, devPtr *usb.Device, logger *slog.Logger) error {
		if devPtr == nil || *devPtr == nil {
			return fmt.Errorf("nil device")
		}
		xdev, ok := (*devPtr).(*Xbox360)
		if !ok {
			return fmt.Errorf("device is not xbox360")
		}

		xdev.SetRumbleCallback(func(rumble XRumbleState) {
			data, err := rumble.MarshalBinary()
			if err != nil {
				logger.Error("failed to marshal rumble", "error", err)
				return
			}
			if _, err := conn.Write(data); err != nil {
				logger.Error("failed to send rumble", "error", err)
			}
		})

		buf := make([]byte, 14)
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
			xdev.UpdateInputState(state)
		}
	}
}
