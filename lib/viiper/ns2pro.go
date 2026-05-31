package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef uintptr_t USBServerHandle;

typedef uintptr_t NS2ProDeviceHandle;

#define NS2PRO_BUTTON_B            0x00000001u
#define NS2PRO_BUTTON_A            0x00000002u
#define NS2PRO_BUTTON_Y            0x00000004u
#define NS2PRO_BUTTON_X            0x00000008u
#define NS2PRO_BUTTON_R            0x00000010u
#define NS2PRO_BUTTON_ZR           0x00000020u
#define NS2PRO_BUTTON_PLUS         0x00000040u
#define NS2PRO_BUTTON_RIGHT_STICK  0x00000080u
#define NS2PRO_BUTTON_DOWN         0x00000100u
#define NS2PRO_BUTTON_RIGHT        0x00000200u
#define NS2PRO_BUTTON_LEFT         0x00000400u
#define NS2PRO_BUTTON_UP           0x00000800u
#define NS2PRO_BUTTON_L            0x00001000u
#define NS2PRO_BUTTON_ZL           0x00002000u
#define NS2PRO_BUTTON_MINUS        0x00004000u
#define NS2PRO_BUTTON_LEFT_STICK   0x00008000u
#define NS2PRO_BUTTON_HOME         0x00010000u
#define NS2PRO_BUTTON_CAPTURE      0x00020000u
#define NS2PRO_BUTTON_GR           0x00040000u
#define NS2PRO_BUTTON_GL           0x00080000u
#define NS2PRO_BUTTON_C            0x00100000u
#define NS2PRO_BUTTON_HEADSET      0x00200000u

#define NS2PRO_STICK_MIN    0x0000u
#define NS2PRO_STICK_CENTER 0x0800u
#define NS2PRO_STICK_MAX    0x0FFFu

#define NS2PRO_FEATURE_BUTTONS 0x01u
#define NS2PRO_FEATURE_STICKS  0x02u
#define NS2PRO_FEATURE_IMU     0x04u
#define NS2PRO_FEATURE_MOUSE   0x10u
#define NS2PRO_FEATURE_RUMBLE  0x20u

typedef struct {
	uint32_t Buttons;
	uint16_t LX;
	uint16_t LY;
	uint16_t RX;
	uint16_t RY;
	int16_t  AccelX;
	int16_t  AccelY;
	int16_t  AccelZ;
	int16_t  GyroX;
	int16_t  GyroY;
	int16_t  GyroZ;
} NS2ProDeviceState;

typedef struct {
	const char* SerialNumber;  // NULL = use default
	uint8_t     BatteryLevel;  // 0-9; 0 = use default (9 = full)
	uint8_t     Charging;      // 0 = not charging
	uint8_t     ExternalPower; // 0 = battery only
	uint16_t    BatteryVolts;  // mV; 0 = use default (3800)
} NS2ProMetaState;

typedef struct {
	uint8_t LeftRumble[16];
	uint8_t RightRumble[16];
	uint8_t Flags;
	uint8_t PlayerLedMask;
} NS2ProOutputState;

typedef void (*NS2ProOutputCallback)(NS2ProDeviceHandle handle, NS2ProOutputState output);

static void viiper_call_ns2pro_output(NS2ProOutputCallback fn, NS2ProDeviceHandle handle, NS2ProOutputState output) {
	fn(handle, output);
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
	"github.com/Alia5/VIIPER/device/ns2pro"
	"github.com/Alia5/VIIPER/internal/server/api"
)

// CreateNS2ProDevice creates a new Nintendo Switch 2 Pro Controller device on the bus with the given ID on the server associated with the given handle.
// @param serverHandle Handle to the USB server.
// @param outDeviceHandle Output parameter for the created device handle.
// @param busID ID of the bus to add the device to.
// @param autoAttachLocalhost If true, the device will be automatically attached to a USBIP-Client/Driver running on THIS machine.
// @param idVendor Optional USB vendor ID (0 = default).
// @param idProduct Optional USB product ID (0 = default).
// @param meta Optional pointer to initial device metadata. Pass NULL to use defaults.
//
//export CreateNS2ProDevice
func CreateNS2ProDevice(
	serverHandle C.USBServerHandle,
	outDeviceHandle *C.NS2ProDeviceHandle,
	busID uint32,
	autoAttachLocalhost bool,
	idVendor uint16,
	idProduct uint16,
	meta *C.NS2ProMetaState,
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
		opts.IDVendor = &idVendor
	}
	if idProduct != 0 {
		opts.IDProduct = &idProduct
	}
	if meta != nil {
		goMeta := ns2pro.MetaState{
			SerialNumber:  goStringOrEmpty(meta.SerialNumber),
			BatteryLevel:  uint8(meta.BatteryLevel),
			Charging:      meta.Charging != 0,
			ExternalPower: meta.ExternalPower != 0,
			BatteryVolts:  uint16(meta.BatteryVolts),
		}
		b, err := json.Marshal(goMeta)
		if err != nil {
			return false
		}
		opts.DeviceSpecific = string(b)
	}

	d, err := ns2pro.New(opts)
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
	*outDeviceHandle = C.NS2ProDeviceHandle(cgo.NewHandle(handleWrapper))

	shw.mtx.Lock()
	defer shw.mtx.Unlock()
	shw.deviceHandles[busID] = append(shw.deviceHandles[busID], deviceHandle(*outDeviceHandle))
	return true
}

// SetNS2ProDeviceState updates the input state of the NS2Pro device associated with the given handle.
// @param handle Handle to the NS2Pro device.
// @param state New input state to set on the device.
//
//export SetNS2ProDeviceState
func SetNS2ProDeviceState(handle C.NS2ProDeviceHandle, state C.NS2ProDeviceState) bool {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return false
	}
	ns2device, ok := dhw.device.(*ns2pro.NS2Pro)
	if !ok {
		return false
	}
	s := ns2pro.InputState{
		Buttons: uint32(state.Buttons),
		LX:      uint16(state.LX),
		LY:      uint16(state.LY),
		RX:      uint16(state.RX),
		RY:      uint16(state.RY),
		AccelX:  int16(state.AccelX),
		AccelY:  int16(state.AccelY),
		AccelZ:  int16(state.AccelZ),
		GyroX:   int16(state.GyroX),
		GyroY:   int16(state.GyroY),
		GyroZ:   int16(state.GyroZ),
	}
	ns2device.UpdateInputState(s)
	return true
}

// SetNS2ProOutputCallback sets a callback to be invoked when the host sends output (rumble/LED) commands to the device.
// @param handle Handle to the NS2Pro device.
// @param callback Callback receiving the full output state (HD rumble data, flags, player LED mask). Pass NULL to clear.
//
//export SetNS2ProOutputCallback
func SetNS2ProOutputCallback(handle C.NS2ProDeviceHandle, cb C.NS2ProOutputCallback) bool {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return false
	}
	ns2device, ok := dhw.device.(*ns2pro.NS2Pro)
	if !ok {
		return false
	}
	if cb == nil {
		ns2device.SetOutputCallback(nil)
		return true
	}
	ns2device.SetOutputCallback(func(out ns2pro.OutputState) {
		var cOut C.NS2ProOutputState
		for i := 0; i < 16; i++ {
			cOut.LeftRumble[i] = C.uint8_t(out.LeftRumble[i])
			cOut.RightRumble[i] = C.uint8_t(out.RightRumble[i])
		}
		cOut.Flags = C.uint8_t(out.Flags)
		cOut.PlayerLedMask = C.uint8_t(out.PlayerLedMask)
		C.viiper_call_ns2pro_output(cb, handle, cOut)
	})
	return true
}

// RemoveNS2ProDevice removes the NS2Pro device associated with the given handle from the server.
// @param handle Handle to the NS2Pro device to remove.
//
//export RemoveNS2ProDevice
func RemoveNS2ProDevice(handle C.NS2ProDeviceHandle) bool {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return false
	}
	if err := dhw.usbServer.s.RemoveDeviceByID(dhw.exportMeta.BusID, fmt.Sprintf("%d", dhw.exportMeta.DevID)); err != nil {
		return false
	}

	shw := dhw.usbServer
	busID := dhw.exportMeta.BusID

	shw.mtx.Lock()
	defer shw.mtx.Unlock()
	shw.deviceHandles[busID] = slices.DeleteFunc(shw.deviceHandles[busID], func(h deviceHandle) bool {
		return h == deviceHandle(handle)
	})
	dh.Delete()

	return true
}
