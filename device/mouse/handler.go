package mouse

import (
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/usb"
	"github.com/Alia5/VIIPER/viiperclient"
)

var TypeName = viiperclient.DeviceType[viiperclient.NoOptions]{
	Name: "mouse",
}

func init() {
	api.RegisterDevice(TypeName.Name, &handler{})
}

type handler struct{}

func (h *handler) CreateDevice(o *device.CreateOptions) (usb.Device, error) { return New(o) }

func (h *handler) StreamHandler() api.StreamHandlerFunc {
	return func(conn net.Conn, devPtr *usb.Device, logger *slog.Logger) error {
		if devPtr == nil || *devPtr == nil {
			return fmt.Errorf("nil device")
		}
		mdev, ok := (*devPtr).(*Mouse)
		if !ok {
			return fmt.Errorf("device is not mouse")
		}

		buf := make([]byte, 9)
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
			mdev.UpdateInputState(state)
		}
	}
}

func (h *handler) UpdateMetaState(meta string, dev *usb.Device) error {
	return nil
}
