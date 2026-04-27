package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef uintptr_t USBServerHandle;

typedef uintptr_t MouseDeviceHandle;

#define MOUSE_BTN_LEFT    0x01u
#define MOUSE_BTN_RIGHT   0x02u
#define MOUSE_BTN_MIDDLE  0x04u
#define MOUSE_BTN_BACK    0x08u
#define MOUSE_BTN_FORWARD 0x10u

typedef struct {
	uint8_t Buttons;
	int16_t DX;
	int16_t DY;
	int16_t Wheel;
	int16_t Pan;
} MouseDeviceState;

*/
import "C"
import (
	"context"
	"fmt"
	"log/slog"
	"runtime/cgo"
	"slices"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/mouse"
	"github.com/Alia5/VIIPER/internal/server/api"
)

// CreateMouseDevice creates a new HID mouse device on the bus with the given ID on the server associated with the given handle.
// @param serverHandle Handle to the USB server.
// @param outDeviceHandle Output parameter for the created device handle.
// @param busID ID of the bus to add the device to.
// @param autoAttachLocalhost If true, the device will be automatically attached to a USBIP-Client/Driver running on THIS machine.
// @param idVendor Optional USB vendor ID (0 = default).
// @param idProduct Optional USB product ID (0 = default).
//
//export CreateMouseDevice
func CreateMouseDevice(
	serverHandle C.USBServerHandle,
	outDeviceHandle *C.MouseDeviceHandle,
	busID uint32,
	autoAttachLocalhost bool,
	idVendor uint16,
	idProduct uint16,
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

	d, err := mouse.New(opts)
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
	*outDeviceHandle = C.MouseDeviceHandle(cgo.NewHandle(handleWrapper))

	shw.mtx.Lock()
	defer shw.mtx.Unlock()
	shw.deviceHandles[busID] = append(shw.deviceHandles[busID], deviceHandle(*outDeviceHandle))
	return true
}

// SetMouseDeviceState updates the input state of the mouse device associated with the given handle.
// @param handle Handle to the mouse device.
// @param state New input state. DX/DY/Wheel/Pan are relative and consumed each poll cycle.
//
//export SetMouseDeviceState
func SetMouseDeviceState(handle C.MouseDeviceHandle, state C.MouseDeviceState) bool {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return false
	}
	mouseDevice, ok := dhw.device.(*mouse.Mouse)
	if !ok {
		return false
	}
	mouseDevice.UpdateInputState(mouse.InputState{
		Buttons: uint8(state.Buttons),
		DX:      int16(state.DX),
		DY:      int16(state.DY),
		Wheel:   int16(state.Wheel),
		Pan:     int16(state.Pan),
	})
	return true
}

// RemoveMouseDevice removes the mouse device associated with the given handle from the server.
// @param handle Handle to the mouse device to remove.
//
//export RemoveMouseDevice
func RemoveMouseDevice(handle C.MouseDeviceHandle) bool {
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
