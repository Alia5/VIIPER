package steamdeck

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
	api.RegisterDevice("steamdeck", &handler{})
}

type handler struct{}

func (h *handler) CreateDevice(o *device.CreateOptions) usb.Device { return New(o) }

func (h *handler) StreamHandler() api.StreamHandlerFunc {
	return func(conn net.Conn, devPtr *usb.Device, logger *slog.Logger) error {
		if devPtr == nil || *devPtr == nil {
			return fmt.Errorf("nil device")
		}
		sdev, ok := (*devPtr).(*SteamDeck)
		if !ok {
			return fmt.Errorf("device is not steamdeck")
		}

		sdev.SetOutputCallback(func(r OutputState) {
			//	TODO: Uncomment
			// data, err := r.MarshalBinary()
			// if err != nil {
			// 	logger.Error("failed to marshal rumble", "error", err)
			// 	return
			// }
			// if _, err := conn.Write(data); err != nil {
			// 	logger.Error("failed to send rumble", "error", err)
			// }
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

			frame := make([]byte, InputStateSize)
			copy(frame, buf)
			var st InputState
			if err := st.UnmarshalBinary(frame); err != nil {
				return fmt.Errorf("decode input state: %w", err)
			}
			sdev.UpdateInputState(st)
		}
	}
}
