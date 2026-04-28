package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef uintptr_t USBServerHandle;

typedef uintptr_t Xbox360DeviceHandle;

#define XBOX360_BUTTON_DPAD_UP     0x0001u
#define XBOX360_BUTTON_DPAD_DOWN   0x0002u
#define XBOX360_BUTTON_DPAD_LEFT   0x0004u
#define XBOX360_BUTTON_DPAD_RIGHT  0x0008u
#define XBOX360_BUTTON_START       0x0010u
#define XBOX360_BUTTON_BACK        0x0020u
#define XBOX360_BUTTON_LTHUMB      0x0040u
#define XBOX360_BUTTON_RTHUMB      0x0080u
#define XBOX360_BUTTON_LSHOULDER   0x0100u
#define XBOX360_BUTTON_RSHOULDER   0x0200u
#define XBOX360_BUTTON_GUIDE       0x0400u
#define XBOX360_BUTTON_A           0x1000u
#define XBOX360_BUTTON_B           0x2000u
#define XBOX360_BUTTON_X           0x4000u
#define XBOX360_BUTTON_Y           0x8000u

typedef struct {
	// Button bitfield (lower 16 bits used typically), higher bits reserved
	uint32_t Buttons;
	// Triggers: 0-255
	uint8_t LT;
	uint8_t RT;
	// Sticks: signed 16-bit little endian values
	int16_t LX;
	int16_t LY;
	int16_t RX;
	int16_t RY;
	uint8_t Reserved[6];
} Xbox360DeviceState;

typedef void (*Xbox360RumbleCallback)(Xbox360DeviceHandle handle, uint8_t leftMotor, uint8_t rightMotor);

static void viiper_call_rumble(Xbox360RumbleCallback fn, Xbox360DeviceHandle handle, uint8_t left, uint8_t right) {
	fn(handle, left, right);
}

*/
import "C"
import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/cgo"
	"slices"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/xbox360"
	"github.com/Alia5/VIIPER/internal/server/api"
)

// CreateXbox360Device creates a new Xbox360 device on the bus with the given ID on the server associated with the given handle.
// @param serverHandle Handle to the USB server.
// @param outDeviceHandle Output parameter for the created device handle.
// @param busID ID of the bus to add the device to.
// @param idVendor Optional USB vendor ID to set on the device.
// @param idProduct Optional USB product ID to set on the device.
// @param xinputSubType Optional XInput subtype to set on the device (e.g. 0x01 for gamepad, 0x02 for wheel, etc.). (Default gamepad)
// @param autoAttachLocalhost If true, the device will be automatically attached to a USBIP-Client/Driver running on THIS machine. (uses IOCTL on windows, USBIP binary on linux)
//
//export CreateXbox360Device
func CreateXbox360Device(
	serverHandle C.USBServerHandle,
	outDeviceHandle *C.Xbox360DeviceHandle,
	busID uint32,
	autoAttachLocalhost bool,
	idVendor uint16,
	idProduct uint16,
	xinputSubType uint8,
) bool {

	sh := cgo.Handle(serverHandle)
	shw, ok := sh.Value().(*usbServerHandleWrapper)
	if !ok {
		return false
	}
	bus := shw.s.GetBus(busID)
	if bus == nil {
		return false
	}

	opts := &device.CreateOptions{}
	if idVendor != 0 {
		opts.IdVendor = &idVendor
	}
	if idProduct != 0 {
		opts.IdProduct = &idProduct
	}
	if xinputSubType != 0 {
		subOpts := &xbox360.Xbox360CreateOptions{
			SubType: &xinputSubType,
		}
		str, err := json.Marshal(subOpts)
		if err != nil {
			return false
		}
		var deviceSpecific map[string]any
		err = json.Unmarshal(str, &deviceSpecific)
		if err != nil {
			return false
		}
		opts.DeviceSpecific = deviceSpecific
	}
	d, err := xbox360.New(opts)
	if err != nil {
		return false
	}
	devCtx, err := bus.Add(d)
	if err != nil {
		return false
	}
	exportMeta := device.GetDeviceMeta(devCtx)
	if exportMeta == nil {
		return false
	}

	if autoAttachLocalhost {
		err := api.AttachLocalhostClient(
			context.Background(),
			exportMeta,
			shw.s.GetListenPort(),
			true,
			slog.Default(),
		)
		if err != nil {
			slog.Error("failed to auto-attach localhost client", "error", err)
			return false
		}
	}

	handleWrapper := &deviceHandleWrapper{
		device:     d,
		exportMeta: exportMeta,
		usbServer:  shw,
	}
	*outDeviceHandle = C.Xbox360DeviceHandle(cgo.NewHandle(handleWrapper))

	shw.mtx.Lock()
	defer shw.mtx.Unlock()
	shw.deviceHandles[busID] = append(shw.deviceHandles[busID], deviceHandle(*outDeviceHandle))
	return true
}

// SetXbox360DeviceState updates the input state of the Xbox360 device associated with the given handle.
// @param deviceHandle Handle to the Xbox360 device to update.
// @param state New input state to set on the device.^
//
//export SetXbox360DeviceState
func SetXbox360DeviceState(handle C.Xbox360DeviceHandle, state C.Xbox360DeviceState) bool {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return false
	}
	xbox360device, ok := dhw.device.(*xbox360.Xbox360)
	if !ok {
		return false
	}
	deviceState := xbox360.InputState{
		Buttons: uint32(state.Buttons),
		LT:      uint8(state.LT),
		RT:      uint8(state.RT),
		LX:      int16(state.LX),
		LY:      int16(state.LY),
		RX:      int16(state.RX),
		RY:      int16(state.RY),
	}
	for i, v := range state.Reserved {
		deviceState.Reserved[i] = byte(v)
	}

	xbox360device.UpdateInputState(deviceState)

	return true
}

// RemoveXbox360Device removes the Xbox360 device associated with the given handle from the server.
// @param deviceHandle Handle to the Xbox360 device to remove.
//
//export RemoveXbox360Device
func RemoveXbox360Device(handle C.Xbox360DeviceHandle) bool {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return false
	}
	if err := dhw.usbServer.s.RemoveDeviceByID(dhw.exportMeta.BusId, fmt.Sprintf("%d", dhw.exportMeta.DevId)); err != nil {
		return false
	}

	shw := dhw.usbServer
	busID := dhw.exportMeta.BusId

	shw.mtx.Lock()
	defer shw.mtx.Unlock()
	shw.deviceHandles[busID] = slices.DeleteFunc(shw.deviceHandles[busID], func(h deviceHandle) bool {
		return h == deviceHandle(handle)
	})
	dh.Delete()

	return true
}

// SetXbox360RumbleCallback sets a callback to be invoked when the host sends rumble/motor commands to the device.
// @param handle Handle to the Xbox360 device.
// @param callback Callback function receiving the device handle and left/right motor intensities (0-255). Pass NULL to clear.
//
//export SetXbox360RumbleCallback
func SetXbox360RumbleCallback(handle C.Xbox360DeviceHandle, cb C.Xbox360RumbleCallback) bool {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return false
	}
	xbox360device, ok := dhw.device.(*xbox360.Xbox360)
	if !ok {
		return false
	}
	if cb == nil {
		xbox360device.SetRumbleCallback(nil)
		return true
	}
	xbox360device.SetRumbleCallback(func(rumble xbox360.XRumbleState) {
		C.viiper_call_rumble(cb, handle, C.uint8_t(rumble.LeftMotor), C.uint8_t(rumble.RightMotor))
	})
	return true
}
