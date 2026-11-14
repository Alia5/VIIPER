package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef uintptr_t USBServerHandle;

typedef uintptr_t DSDeviceHandle;

#define DS_VENDOR_ID       0x054Cu
#define DS_PRODUCT_ID      0x0CE6u
#define DS_PRODUCT_ID_EDGE 0x0DF2u

#define DS_TOUCHPAD_MAX_X 1920
#define DS_TOUCHPAD_MAX_Y 1080

#define DS_GYRO_COUNTS_PER_DPS  16.384
#define DS_ACCEL_COUNTS_PER_MS2 835.07

#define DS_OUTPUT_REPORT_SIZE 64

#define DS_OUTPUT_REPORT_OFFSET_UPDATE_FLAGS1     1
#define DS_OUTPUT_REPORT_OFFSET_UPDATE_FLAGS2     2
#define DS_OUTPUT_REPORT_OFFSET_RUMBLE_SMALL      3
#define DS_OUTPUT_REPORT_OFFSET_RUMBLE_LARGE      4
#define DS_OUTPUT_REPORT_OFFSET_TRIGGER_R2_MODE   11
#define DS_OUTPUT_REPORT_OFFSET_TRIGGER_R2_PARAMS 12
#define DS_OUTPUT_REPORT_OFFSET_TRIGGER_L2_MODE   22
#define DS_OUTPUT_REPORT_OFFSET_TRIGGER_L2_PARAMS 23
#define DS_OUTPUT_REPORT_OFFSET_UPDATE_FLAGS3     39
#define DS_OUTPUT_REPORT_OFFSET_PLAYER_LEDS       44
#define DS_OUTPUT_REPORT_OFFSET_LED_RED           45
#define DS_OUTPUT_REPORT_OFFSET_LED_GREEN         46
#define DS_OUTPUT_REPORT_OFFSET_LED_BLUE          47

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
#define DS_BUTTON_RFN       0x00200000u
#define DS_BUTTON_LFN       0x00100000u
#define DS_BUTTON_R4        0x00800000u
#define DS_BUTTON_L4        0x00400000u

#define DS_DPAD_UP     0x01u
#define DS_DPAD_DOWN   0x02u
#define DS_DPAD_LEFT   0x04u
#define DS_DPAD_RIGHT  0x08u

#define DS_OUTPUT_UPDATE1_RUMBLE         0x01u
#define DS_OUTPUT_UPDATE1_HAPTICS_SELECT 0x02u
#define DS_OUTPUT_UPDATE1_TRIGGER_R2     0x04u
#define DS_OUTPUT_UPDATE1_TRIGGER_L2     0x08u
#define DS_OUTPUT_UPDATE1_SPEAKER_VOLUME    0x20u
#define DS_OUTPUT_UPDATE1_MICROPHONE_VOLUME 0x40u
#define DS_OUTPUT_UPDATE1_AUDIO_CONTROL     0x80u

#define DS_OUTPUT_UPDATE2_MIC_MUTE_LED       0x01u
#define DS_OUTPUT_UPDATE2_POWER_SAVE_CONTROL 0x02u
#define DS_OUTPUT_UPDATE2_LED            0x04u
#define DS_OUTPUT_UPDATE2_PLAYER_LEDS    0x10u
#define DS_OUTPUT_UPDATE2_AUDIO_CONTROL2 0x80u

#define DS_OUTPUT_UPDATE3_IMPROVED_RUMBLE 0x04u

#define DS_TRIGGER_PARAMS_SIZE 10
#define DS_TRIGGER_ZONE_COUNT  10

#define DS_TRIGGER_MODE_SIMPLE_FEEDBACK  0x01u
#define DS_TRIGGER_MODE_SIMPLE_WEAPON    0x02u
#define DS_TRIGGER_MODE_OFF              0x05u
#define DS_TRIGGER_MODE_SIMPLE_VIBRATION 0x06u
#define DS_TRIGGER_MODE_LIMITED_FEEDBACK 0x11u
#define DS_TRIGGER_MODE_LIMITED_WEAPON   0x12u
#define DS_TRIGGER_MODE_FEEDBACK         0x21u
#define DS_TRIGGER_MODE_BOW              0x22u
#define DS_TRIGGER_MODE_GALLOPING        0x23u
#define DS_TRIGGER_MODE_WEAPON           0x25u
#define DS_TRIGGER_MODE_VIBRATION        0x26u
#define DS_TRIGGER_MODE_MACHINE          0x27u
#define DS_TRIGGER_MODE_CALIBRATION      0xFCu
#define DS_TRIGGER_MODE_DEBUG_FD         0xFDu
#define DS_TRIGGER_MODE_DEBUG_FE         0xFEu

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
	uint8_t     HeadphonesConnected; // 0 = no headphones in the audio jack
	uint8_t     MicrophoneConnected; // 0 = no headset microphone in the audio jack
} DSMetaState;

typedef struct {
	uint8_t Mode;                                 // DS_TRIGGER_MODE_*
	uint8_t Params[DS_TRIGGER_PARAMS_SIZE];       // raw effect parameters
	uint8_t ZoneStrengths[DS_TRIGGER_ZONE_COUNT]; // feedback, vibration: 0 = zone not active, 1-8
	uint8_t StartZone;                            // weapon, bow, galloping, machine: 0-9
	uint8_t EndZone;                              // weapon, bow, galloping, machine: 0-9
	uint8_t StartPosition;                        // simple and limited modes: raw position
	uint8_t EndPosition;                          // simple and limited weapon: raw position
	uint8_t Strength;                             // weapon, bow: 1-8; simple and limited modes: raw strength
	uint8_t SnapForce;                            // bow: 1-8
	uint8_t AmplitudeA;                           // machine: 0-7
	uint8_t AmplitudeB;                           // machine: 0-7
	uint8_t FirstFoot;                            // galloping: 0-6
	uint8_t SecondFoot;                           // galloping: 1-7
	uint8_t Frequency;                            // vibration, galloping, machine, simple vibration: Hz
	uint8_t Period;                               // machine: tenths of a second
} DSTriggerEffect;

typedef struct {
	uint8_t         UpdateFlags1;     // DS_OUTPUT_UPDATE1_*
	uint8_t         UpdateFlags2;     // DS_OUTPUT_UPDATE2_*
	uint8_t         UpdateFlags3;     // DS_OUTPUT_UPDATE3_*
	uint8_t         RumbleSmall;
	uint8_t         RumbleLarge;
	uint8_t         LedRed;
	uint8_t         LedGreen;
	uint8_t         LedBlue;
	uint8_t         PlayerLeds;
	DSTriggerEffect TriggerR2;
	DSTriggerEffect TriggerL2;
} DSOutputState;

typedef void (*DSOutputCallback)(DSDeviceHandle handle, DSOutputState state);

static void viiper_call_ds_output(DSOutputCallback fn, DSDeviceHandle handle, DSOutputState state) {
	fn(handle, state);
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
	"unsafe"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/dualsense"
	"github.com/Alia5/VIIPER/internal/server/api"
)

type (
	dsDeviceState = C.DSDeviceState
	dsMetaState   = C.DSMetaState
	dsOutputState = C.DSOutputState
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
			SerialNumber:        goStringOrEmpty(meta.SerialNumber),
			MACAddress:          goStringOrEmpty(meta.MACAddress),
			Board:               goStringOrEmpty(meta.Board),
			BatteryStatus:       uint8(meta.BatteryStatus),
			TemperatureCelsius:  float64(meta.TemperatureCelsius),
			BatteryVoltage:      float64(meta.BatteryVoltage),
			ShellColor:          goStringOrEmpty(meta.ShellColor),
			HeadphonesConnected: meta.HeadphonesConnected != 0,
			MicrophoneConnected: meta.MicrophoneConnected != 0,
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
	ctx, cancel := context.WithCancel(context.Background())
	d, err := ctor(ctx, opts)
	if err != nil {
		cancel()
		return false
	}
	devCtx, err := bus.Add(d, cancel)
	if err != nil {
		cancel()
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

// SetDualSenseOutputCallback sets a callback to be invoked when the host sends output (rumble/LED/adaptive trigger) commands to the device.
// @param handle Handle to the DualSense device.
// @param callback Callback receiving the output state. Pass NULL to clear.
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
		C.viiper_call_ds_output(cb, handle, C.DSOutputState{
			UpdateFlags1: C.uint8_t(out.UpdateFlags1),
			UpdateFlags2: C.uint8_t(out.UpdateFlags2),
			UpdateFlags3: C.uint8_t(out.UpdateFlags3),
			RumbleSmall:  C.uint8_t(out.RumbleSmall),
			RumbleLarge:  C.uint8_t(out.RumbleLarge),
			LedRed:       C.uint8_t(out.LedRed),
			LedGreen:     C.uint8_t(out.LedGreen),
			LedBlue:      C.uint8_t(out.LedBlue),
			PlayerLeds:   C.uint8_t(out.PlayerLeds),
			TriggerR2:    dsTriggerEffect(out.TriggerR2()),
			TriggerL2:    dsTriggerEffect(out.TriggerL2()),
		})
	})
	return true
}

// DualSenseOutputReport builds the USB output report (report ID 0x02) of a DualSense from an output state.
// @param state Output state, as received by the output callback.
// @param report Output buffer of DS_OUTPUT_REPORT_SIZE bytes.
//
//export DualSenseOutputReport
func DualSenseOutputReport(state C.DSOutputState, report *C.uint8_t) {
	goState := goDSOutputState(state)
	out := goState.OutputReport()
	copy(unsafe.Slice((*byte)(unsafe.Pointer(report)), len(out)), out[:])
}

// DualSenseGyroDpsToRaw converts an angular velocity in degrees per second to the raw gyroscope value of the device state.
//
//export DualSenseGyroDpsToRaw
func DualSenseGyroDpsToRaw(dps float64) int16 {
	return dualsense.GyroDpsToRaw(dps)
}

// DualSenseAccelMS2ToRaw converts an acceleration in m/s² to the raw accelerometer value of the device state.
//
//export DualSenseAccelMS2ToRaw
func DualSenseAccelMS2ToRaw(ms2 float64) int16 {
	return dualsense.AccelMS2ToRaw(ms2)
}

func goDSOutputState(state dsOutputState) dualsense.OutputState {
	out := dualsense.OutputState{
		UpdateFlags1:  uint8(state.UpdateFlags1),
		UpdateFlags2:  uint8(state.UpdateFlags2),
		UpdateFlags3:  uint8(state.UpdateFlags3),
		RumbleSmall:   uint8(state.RumbleSmall),
		RumbleLarge:   uint8(state.RumbleLarge),
		LedRed:        uint8(state.LedRed),
		LedGreen:      uint8(state.LedGreen),
		LedBlue:       uint8(state.LedBlue),
		PlayerLeds:    uint8(state.PlayerLeds),
		TriggerR2Mode: uint8(state.TriggerR2.Mode),
		TriggerL2Mode: uint8(state.TriggerL2.Mode),
	}
	for i := range out.TriggerR2Params {
		out.TriggerR2Params[i] = uint8(state.TriggerR2.Params[i])
		out.TriggerL2Params[i] = uint8(state.TriggerL2.Params[i])
	}
	return out
}

func dsTriggerEffect(effect dualsense.TriggerEffect) C.DSTriggerEffect {
	out := C.DSTriggerEffect{
		Mode:          C.uint8_t(effect.Mode),
		StartZone:     C.uint8_t(effect.StartZone),
		EndZone:       C.uint8_t(effect.EndZone),
		StartPosition: C.uint8_t(effect.StartPosition),
		EndPosition:   C.uint8_t(effect.EndPosition),
		Strength:      C.uint8_t(effect.Strength),
		SnapForce:     C.uint8_t(effect.SnapForce),
		AmplitudeA:    C.uint8_t(effect.AmplitudeA),
		AmplitudeB:    C.uint8_t(effect.AmplitudeB),
		FirstFoot:     C.uint8_t(effect.FirstFoot),
		SecondFoot:    C.uint8_t(effect.SecondFoot),
		Frequency:     C.uint8_t(effect.Frequency),
		Period:        C.uint8_t(effect.Period),
	}
	for i, param := range effect.Params {
		out.Params[i] = C.uint8_t(param)
	}
	for i, strength := range effect.ZoneStrengths {
		out.ZoneStrengths[i] = C.uint8_t(strength)
	}
	return out
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
