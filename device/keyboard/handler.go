package keyboard

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
	api.RegisterDevice("keyboard", &handler{})
}

type handler struct{}

func (h *handler) CreateDevice(o *device.CreateOptions) (usb.Device, error) { return New(o) }

func (h *handler) StreamHandler() api.StreamHandlerFunc {
	return func(conn net.Conn, devPtr *usb.Device, logger *slog.Logger) error {
		if devPtr == nil || *devPtr == nil {
			return fmt.Errorf("nil device")
		}
		kdev, ok := (*devPtr).(*Keyboard)
		if !ok {
			return fmt.Errorf("device is not keyboard")
		}

		// Set LED callback to write LED state to client
		kdev.SetLEDCallback(func(led LEDState) {
			ledByte := uint8(0)
			if led.NumLock {
				ledByte |= LEDNumLock
			}
			if led.CapsLock {
				ledByte |= LEDCapsLock
			}
			if led.ScrollLock {
				ledByte |= LEDScrollLock
			}
			if led.Compose {
				ledByte |= LEDCompose
			}
			if led.Kana {
				ledByte |= LEDKana
			}
			if _, err := conn.Write([]byte{ledByte}); err != nil {
				logger.Warn("failed to write LED state", "error", err)
			}
		})

		// Read loop: Client → Device (key presses)
		for {
			// Read header (2 bytes minimum: modifiers + key count)
			header := make([]byte, 2)
			if _, err := io.ReadFull(conn, header); err != nil {
				if err == io.EOF {
					logger.Info("client disconnected")
					return nil
				}
				return fmt.Errorf("read header: %w", err)
			}

			keyCount := header[1]

			// Read key codes
			keys := make([]byte, keyCount)
			if keyCount > 0 {
				if _, err := io.ReadFull(conn, keys); err != nil {
					return fmt.Errorf("read keys: %w", err)
				}
			}

			// Build full packet and unmarshal
			fullPacket := append(header, keys...)
			var state InputState
			if err := state.UnmarshalBinary(fullPacket); err != nil {
				return fmt.Errorf("unmarshal input state: %w", err)
			}

			kdev.UpdateInputState(state)
		}
	}
}
