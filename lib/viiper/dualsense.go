package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef uintptr_t USBServerHandle;

typedef uintptr_t DSDeviceHandle;

#define DS_BUTTON_SQUARE    0x00000010u
#define DS_BUTTON_CROSS     0x00000020u
#define DS_BUTTON_CIRCLE    0x00000040u
#define DS_BUTTON_TRIANGLE  0x00000080u
#define DS_BUTTON_L1        0x00000100u
#define DS_BUTTON_R1        0x00000200u
#define DS_BUTTON_L2        0x00000400u
#define DS_BUTTON_R2        0x00000800u
#define DS_BUTTON_CREATE    0x00001000u
#define DS_BUTTON_OPTIONS   0x00002000u
#define DS_BUTTON_L3        0x00004000u
#define DS_BUTTON_R3        0x00008000u
#define DS_BUTTON_PS        0x00010000u
#define DS_BUTTON_TOUCHPAD  0x00020000u
#define DS_BUTTON_MIC_MUTE  0x00040000u
#define DS_BUTTON_RFN       0x00080000u
#define DS_BUTTON_LFN       0x00100000u
#define DS_BUTTON_R4        0x00200000u
#define DS_BUTTON_L4        0x00400000u

#define DS_DPAD_UP     0x01u
#define DS_DPAD_DOWN   0x02u
#define DS_DPAD_LEFT   0x04u
#define DS_DPAD_RIGHT  0x08u

#define DS_SHELL_COLOR_WHITE                    "00"
#define DS_SHELL_COLOR_BLACK                    "01"
#define DS_SHELL_COLOR_COSMIC_RED               "02"
#define DS_SHELL_COLOR_NOVA_PINK                "03"
#define DS_SHELL_COLOR_GALACTIC_PURPLE          "04"
#define DS_SHELL_COLOR_STARLIGHT_BLUE           "05"
#define DS_SHELL_COLOR_GREY_CAMOUFLAGE          "06"
#define DS_SHELL_COLOR_VOLCANIC_RED             "07"
#define DS_SHELL_COLOR_STERLING_SILVER          "08"
#define DS_SHELL_COLOR_COBALT_BLUE              "09"
#define DS_SHELL_COLOR_CHROMA_TEAL              "10"
#define DS_SHELL_COLOR_CHROMA_INDIGO            "11"
#define DS_SHELL_COLOR_CHROMA_PEARL             "12"
#define DS_SHELL_COLOR_ANNIVERSARY_30TH         "30"
#define DS_SHELL_COLOR_GOD_OF_WAR_RAGNAROK      "Z1"
#define DS_SHELL_COLOR_SPIDER_MAN_2             "Z2"
#define DS_SHELL_COLOR_ASTRO_BOT                "Z3"
#define DS_SHELL_COLOR_FORTNITE                 "Z4"
#define DS_SHELL_COLOR_MONSTER_HUNTER_WILDS     "Z5"
#define DS_SHELL_COLOR_THE_LAST_OF_US           "Z6"
#define DS_SHELL_COLOR_GHOST_OF_YOTEI           "Z7"
#define DS_SHELL_COLOR_ICON_BLUE_LIMITED_EDITION "ZB"
#define DS_SHELL_COLOR_ASTRO_BOT_JOYFUL_EDITION "ZC"
#define DS_SHELL_COLOR_GENSHIN_IMPACT           "ZE"

typedef struct {
	int8_t   LX;
	int8_t   LY;
	int8_t   RX;
	int8_t   RY;
	uint32_t Buttons;
	uint8_t  DPad;
	uint8_t  L2;
	uint8_t  R2;
	uint16_t Touch1X;
	uint16_t Touch1Y;
	uint8_t  Touch1Active;
	uint16_t Touch2X;
	uint16_t Touch2Y;
	uint8_t  Touch2Active;
	int16_t  GyroX;
	int16_t  GyroY;
	int16_t  GyroZ;
	int16_t  AccelX;
	int16_t  AccelY;
	int16_t  AccelZ;
} DSDeviceState;

typedef struct {
	const char* SerialNumber;       // NULL = use default
	const char* MACAddress;         // NULL = use default
	const char* Board;              // NULL = use default
	uint8_t     BatteryStatus;      // 0 = use default
	double      TemperatureCelsius; // 0 = use default
	double      BatteryVoltage;     // 0 = use default
	const char* ShellColor;     // NULL = use default (2-char code, e.g. "00", "Z1")
} DSMetaState;

typedef void (*DSOutputCallback)(DSDeviceHandle handle, uint8_t rumbleSmall, uint8_t rumbleLarge, uint8_t ledRed, uint8_t ledGreen, uint8_t ledBlue, uint8_t playerLeds);

static void viiper_call_ds_output(DSOutputCallback fn, DSDeviceHandle handle, uint8_t rumbleSmall, uint8_t rumbleLarge, uint8_t ledRed, uint8_t ledGreen, uint8_t ledBlue, uint8_t playerLeds) {
	fn(handle, rumbleSmall, rumbleLarge, ledRed, ledGreen, ledBlue, playerLeds);
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
	"github.com/Alia5/VIIPER/device/dualsense"
	"github.com/Alia5/VIIPER/internal/server/api"
)

// CreateDualSenseDevice creates a new DualSense (non-edge) device on the bus with the given ID on the server associated with the given handle.
// @param serverHandle Handle to the USB server.
// @param outDeviceHandle Output parameter for the created device handle.
// @param busID ID of the bus to add the device to.
// @param autoAttachLocalhost If true, the device will be automatically attached to a USBIP-Client/Driver running on THIS machine.
// @param idVendor Optional USB vendor ID (0 = default).
// @param idProduct Optional USB product ID (0 = default).
// @param meta Optional pointer to initial device metadata. Pass NULL to use defaults.
//
//export CreateDualSenseDevice
func CreateDualSenseDevice(
	serverHandle C.USBServerHandle,
	outDeviceHandle *C.DSDeviceHandle,
	busID uint32,
	autoAttachLocalhost bool,
	idVendor uint16,
	idProduct uint16,
	meta *C.DSMetaState,
) bool {
	return createDualSenseDevice(serverHandle, outDeviceHandle, busID, autoAttachLocalhost, idVendor, idProduct, meta, false)
}

// CreateDualSenseEdgeDevice creates a new DualSense Edge device on the bus with the given ID on the server associated with the given handle.
// @param serverHandle Handle to the USB server.
// @param outDeviceHandle Output parameter for the created device handle.
// @param busID ID of the bus to add the device to.
// @param autoAttachLocalhost If true, the device will be automatically attached to a USBIP-Client/Driver running on THIS machine.
// @param idVendor Optional USB vendor ID (0 = default).
// @param idProduct Optional USB product ID (0 = default).
// @param meta Optional pointer to initial device metadata. Pass NULL to use defaults.
//
//export CreateDualSenseEdgeDevice
func CreateDualSenseEdgeDevice(
	serverHandle C.USBServerHandle,
	outDeviceHandle *C.DSDeviceHandle,
	busID uint32,
	autoAttachLocalhost bool,
	idVendor uint16,
	idProduct uint16,
	meta *C.DSMetaState,
) bool {
	return createDualSenseDevice(serverHandle, outDeviceHandle, busID, autoAttachLocalhost, idVendor, idProduct, meta, true)
}

func createDualSenseDevice(
	serverHandle C.USBServerHandle,
	outDeviceHandle *C.DSDeviceHandle,
	busID uint32,
	autoAttachLocalhost bool,
	idVendor uint16,
	idProduct uint16,
	meta *C.DSMetaState,
	edge bool,
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
		goMeta := dualsense.MetaState{
			SerialNumber:       goStringOrEmpty(meta.SerialNumber),
			MACAddress:         goStringOrEmpty(meta.MACAddress),
			Board:              goStringOrEmpty(meta.Board),
			BatteryStatus:      uint8(meta.BatteryStatus),
			TemperatureCelsius: float64(meta.TemperatureCelsius),
			BatteryVoltage:     float64(meta.BatteryVoltage),
			ShellColor:         goStringOrEmpty(meta.ShellColor),
		}
		b, err := json.Marshal(goMeta)
		if err != nil {
			return false
		}
		opts.DeviceSpecific = string(b)
	}

	ctor := dualsense.New
	if edge {
		ctor = dualsense.NewEdge
	}
	d, err := ctor(opts)
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
	*outDeviceHandle = C.DSDeviceHandle(cgo.NewHandle(handleWrapper))

	shw.mtx.Lock()
	defer shw.mtx.Unlock()
	shw.deviceHandles[busID] = append(shw.deviceHandles[busID], deviceHandle(*outDeviceHandle))
	return true
}

// SetDualSenseDeviceState updates the input state of the DualSense device associated with the given handle.
// @param handle Handle to the DualSense device.
// @param state New input state to set on the device.
//
//export SetDualSenseDeviceState
func SetDualSenseDeviceState(handle C.DSDeviceHandle, state C.DSDeviceState) bool {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return false
	}
	dsDevice, ok := dhw.device.(*dualsense.DualSense)
	if !ok {
		return false
	}
	s := &dualsense.InputState{
		LX:           int8(state.LX),
		LY:           int8(state.LY),
		RX:           int8(state.RX),
		RY:           int8(state.RY),
		Buttons:      uint32(state.Buttons),
		DPad:         uint8(state.DPad),
		L2:           uint8(state.L2),
		R2:           uint8(state.R2),
		Touch1X:      uint16(state.Touch1X),
		Touch1Y:      uint16(state.Touch1Y),
		Touch1Active: state.Touch1Active != 0,
		Touch2X:      uint16(state.Touch2X),
		Touch2Y:      uint16(state.Touch2Y),
		Touch2Active: state.Touch2Active != 0,
		GyroX:        int16(state.GyroX),
		GyroY:        int16(state.GyroY),
		GyroZ:        int16(state.GyroZ),
		AccelX:       int16(state.AccelX),
		AccelY:       int16(state.AccelY),
		AccelZ:       int16(state.AccelZ),
	}
	dsDevice.UpdateInputState(s)
	return true
}

// SetDualSenseOutputCallback sets a callback to be invoked when the host sends output (rumble/LED) commands to the device.
// @param handle Handle to the DualSense device.
// @param callback Callback receiving rumbleSmall, rumbleLarge, ledRed, ledGreen, ledBlue, playerLeds. Pass NULL to clear.
//
//export SetDualSenseOutputCallback
func SetDualSenseOutputCallback(handle C.DSDeviceHandle, cb C.DSOutputCallback) bool {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return false
	}
	dsDevice, ok := dhw.device.(*dualsense.DualSense)
	if !ok {
		return false
	}
	if cb == nil {
		dsDevice.SetOutputCallback(nil)
		return true
	}
	dsDevice.SetOutputCallback(func(out dualsense.OutputState) {
		C.viiper_call_ds_output(cb, handle,
			C.uint8_t(out.RumbleSmall),
			C.uint8_t(out.RumbleLarge),
			C.uint8_t(out.LedRed),
			C.uint8_t(out.LedGreen),
			C.uint8_t(out.LedBlue),
			C.uint8_t(out.PlayerLeds),
		)
	})
	return true
}

// RemoveDualSenseDevice removes the DualSense device associated with the given handle from the server.
// @param handle Handle to the DualSense device to remove.
//
//export RemoveDualSenseDevice
func RemoveDualSenseDevice(handle C.DSDeviceHandle) bool {
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
