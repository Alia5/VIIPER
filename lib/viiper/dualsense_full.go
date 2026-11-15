package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef uintptr_t USBServerHandle;

typedef uintptr_t DSDeviceHandle;

#define DS_POWER_SAVE_CONTROL_MICROPHONE_MUTE 0x10u
#define DS_AUDIO_CONTROL_OUTPUT_PATH          0x30u

#define DS_OUTPUT_REPORT_OFFSET_SPEAKER_VOLUME     6
#define DS_OUTPUT_REPORT_OFFSET_MICROPHONE_VOLUME  7
#define DS_OUTPUT_REPORT_OFFSET_AUDIO_CONTROL      8
#define DS_OUTPUT_REPORT_OFFSET_MIC_MUTE_LED       9
#define DS_OUTPUT_REPORT_OFFSET_POWER_SAVE_CONTROL 10
#define DS_OUTPUT_REPORT_OFFSET_AUDIO_CONTROL2     38

#define DS_AUDIO_SAMPLE_RATE            48000
#define DS_AUDIO_BYTES_PER_SAMPLE       2
#define DS_AUDIO_OUTPUT_CHANNELS        4
#define DS_AUDIO_MICROPHONE_CHANNELS    2
#define DS_AUDIO_OUTPUT_PACKET_SIZE     (DS_AUDIO_SAMPLE_RATE / 1000 * DS_AUDIO_OUTPUT_CHANNELS * DS_AUDIO_BYTES_PER_SAMPLE)
#define DS_AUDIO_MICROPHONE_PACKET_SIZE (DS_AUDIO_SAMPLE_RATE / 1000 * DS_AUDIO_MICROPHONE_CHANNELS * DS_AUDIO_BYTES_PER_SAMPLE)

typedef struct {
	uint8_t UpdateFlags1;     // DS_OUTPUT_UPDATE1_*
	uint8_t UpdateFlags2;     // DS_OUTPUT_UPDATE2_*
	uint8_t SpeakerVolume;
	uint8_t MicrophoneVolume;
	uint8_t AudioControl;     // DS_AUDIO_CONTROL_*
	uint8_t MicMuteLed;
	uint8_t PowerSaveControl; // DS_POWER_SAVE_CONTROL_*
	uint8_t AudioControl2;    // bits 0-2: gain of the speaker pre-amplifier
} DSAudioControlState;

typedef void (*DSAudioControlCallback)(DSDeviceHandle handle, DSAudioControlState state);

// pcm: DS_AUDIO_OUTPUT_CHANNELS channels, DS_AUDIO_SAMPLE_RATE Hz, S16LE. Only valid during the call.
typedef void (*DSAudioOutputCallback)(DSDeviceHandle handle, const uint8_t* pcm, uint32_t size);

// pcm: buffer to fill with DS_AUDIO_MICROPHONE_CHANNELS channels, DS_AUDIO_SAMPLE_RATE Hz, S16LE. Only valid during the call.
typedef void (*DSMicrophoneCallback)(DSDeviceHandle handle, uint8_t* pcm, uint32_t size);

static void viiper_call_ds_audio_control(DSAudioControlCallback fn, DSDeviceHandle handle, DSAudioControlState state) {
	fn(handle, state);
}

static void viiper_call_ds_audio_output(DSAudioOutputCallback fn, DSDeviceHandle handle, const uint8_t* pcm, uint32_t size) {
	fn(handle, pcm, size);
}

static void viiper_call_ds_microphone(DSMicrophoneCallback fn, DSDeviceHandle handle, uint8_t* pcm, uint32_t size) {
	fn(handle, pcm, size);
}

*/
import "C"
import (
	"context"
	"encoding/json"
	"log/slog"
	"runtime/cgo"
	"unsafe"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/dualsense"
	dualsensefull "github.com/Alia5/VIIPER/device/dualsense/full"
	"github.com/Alia5/VIIPER/internal/server/api"
)

// CreateDualSenseFull creates a new DualSense (non-edge) device with audio on the bus with the given ID on the server associated with the given handle.
// The returned handle is used with SetDualSenseDeviceState, SetDualSenseOutputCallback, SetDualSenseFullAudioControlCallback, SetDualSenseFullAudioOutputCallback, SetDualSenseFullMicrophoneCallback and RemoveDualSense.
// @param serverHandle Handle to the USB server.
// @param outDeviceHandle Output parameter for the created device handle.
// @param busID ID of the bus to add the device to.
// @param autoAttachLocalhost If true, the device will be automatically attached to a USBIP-Client/Driver running on THIS machine.
// @param idVendor Optional USB vendor ID (0 = default).
// @param idProduct Optional USB product ID (0 = default).
// @param meta Optional pointer to initial device metadata. Pass NULL to use defaults.
//
//export CreateDualSenseFull
func CreateDualSenseFull(
	serverHandle C.USBServerHandle,
	outDeviceHandle *C.DSDeviceHandle,
	busID uint32,
	autoAttachLocalhost bool,
	idVendor uint16,
	idProduct uint16,
	meta *dsMetaState,
) bool {
	return createDualSenseFullDevice(serverHandle, outDeviceHandle, busID, autoAttachLocalhost, idVendor, idProduct, meta, false)
}

// CreateDualSenseEdgeFull creates a new DualSense Edge device with audio on the bus with the given ID on the server associated with the given handle.
// The returned handle is used with SetDualSenseDeviceState, SetDualSenseOutputCallback, SetDualSenseFullAudioControlCallback, SetDualSenseFullAudioOutputCallback, SetDualSenseFullMicrophoneCallback and RemoveDualSense.
// @param serverHandle Handle to the USB server.
// @param outDeviceHandle Output parameter for the created device handle.
// @param busID ID of the bus to add the device to.
// @param autoAttachLocalhost If true, the device will be automatically attached to a USBIP-Client/Driver running on THIS machine.
// @param idVendor Optional USB vendor ID (0 = default).
// @param idProduct Optional USB product ID (0 = default).
// @param meta Optional pointer to initial device metadata. Pass NULL to use defaults.
//
//export CreateDualSenseEdgeFull
func CreateDualSenseEdgeFull(
	serverHandle C.USBServerHandle,
	outDeviceHandle *C.DSDeviceHandle,
	busID uint32,
	autoAttachLocalhost bool,
	idVendor uint16,
	idProduct uint16,
	meta *dsMetaState,
) bool {
	return createDualSenseFullDevice(serverHandle, outDeviceHandle, busID, autoAttachLocalhost, idVendor, idProduct, meta, true)
}

func createDualSenseFullDevice(
	serverHandle C.USBServerHandle,
	outDeviceHandle *C.DSDeviceHandle,
	busID uint32,
	autoAttachLocalhost bool,
	idVendor uint16,
	idProduct uint16,
	meta *dsMetaState,
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
	createOptions := dualsensefull.CreateOptions{
		Meta: dualsense.MetaState{
			ShellColor: dualsense.DefaultShellColor,
		},
	}
	if meta != nil {
		createOptions.Meta = dualsense.MetaState{
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
	}
	b, err := json.Marshal(createOptions)
	if err != nil {
		return false
	}
	opts.DeviceSpecific = string(b)

	ctor := dualsensefull.New
	if edge {
		ctor = dualsensefull.NewEdge
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
		device:     d.DualSense,
		exportMeta: exportMeta,
		usbServer:  shw,
	}
	*outDeviceHandle = C.DSDeviceHandle(cgo.NewHandle(handleWrapper))

	shw.mtx.Lock()
	defer shw.mtx.Unlock()
	shw.deviceHandles[busID] = append(shw.deviceHandles[busID], deviceHandle(*outDeviceHandle))
	return true
}

// DualSenseFullOutputReport builds the USB output report (report ID 0x02) of a DualSense with audio from an output state and an audio control state.
// @param state Output state, as received by the output callback.
// @param audio Audio control state, as received by the audio control callback.
// @param report Output buffer of DS_OUTPUT_REPORT_SIZE bytes.
//
//export DualSenseFullOutputReport
func DualSenseFullOutputReport(state dsOutputState, audio C.DSAudioControlState, report *C.uint8_t) {
	goState := dualsensefull.OutputState{
		OutputState:      goDSOutputState(state),
		SpeakerVolume:    uint8(audio.SpeakerVolume),
		MicrophoneVolume: uint8(audio.MicrophoneVolume),
		AudioControl:     uint8(audio.AudioControl),
		MicMuteLed:       uint8(audio.MicMuteLed),
		PowerSaveControl: uint8(audio.PowerSaveControl),
		AudioControl2:    uint8(audio.AudioControl2),
	}
	out := goState.OutputReport()
	copy(unsafe.Slice((*byte)(unsafe.Pointer(report)), len(out)), out[:])
}

func dualSenseFullDevice(handle C.DSDeviceHandle) *dualsensefull.Device {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return nil
	}
	bus := dhw.usbServer.s.GetBus(dhw.exportMeta.BusID)
	if bus == nil {
		return nil
	}
	for _, dev := range bus.Devices() {
		full, ok := dev.(*dualsensefull.Device)
		if ok && full.DualSense == dhw.device {
			return full
		}
	}
	return nil
}

// SetDualSenseFullAudioOutputCallback sets a callback to be invoked with the audio output of the host.
// The callback receives 4 channel 48 kHz S16LE PCM (channels 0 and 1: speaker/headphones, channels 2 and 3: haptics).
// @param handle Handle to the DualSense device with audio.
// @param callback Callback receiving the PCM. Pass NULL to clear.
//
//export SetDualSenseFullAudioOutputCallback
func SetDualSenseFullAudioOutputCallback(handle C.DSDeviceHandle, cb C.DSAudioOutputCallback) bool {
	dsDevice := dualSenseFullDevice(handle)
	if dsDevice == nil {
		return false
	}
	if cb == nil {
		dsDevice.SetAudioOutputCallback(nil)
		return true
	}
	dsDevice.SetAudioOutputCallback(func(pcm []byte) {
		C.viiper_call_ds_audio_output(cb, handle, (*C.uint8_t)(unsafe.Pointer(&pcm[0])), C.uint32_t(len(pcm)))
	})
	return true
}

// SetDualSenseFullMicrophoneCallback sets a callback to be invoked when the host reads the microphone.
// The callback fills the buffer with 2 channel 48 kHz S16LE PCM. The buffer is zeroed (silence) before the call.
// @param handle Handle to the DualSense device with audio.
// @param callback Callback filling the PCM buffer. Pass NULL to clear.
//
//export SetDualSenseFullMicrophoneCallback
func SetDualSenseFullMicrophoneCallback(handle C.DSDeviceHandle, cb C.DSMicrophoneCallback) bool {
	dsDevice := dualSenseFullDevice(handle)
	if dsDevice == nil {
		return false
	}
	if cb == nil {
		dsDevice.SetMicrophoneCallback(nil)
		return true
	}
	dsDevice.SetMicrophoneCallback(func(pcm []byte) {
		C.viiper_call_ds_microphone(cb, handle, (*C.uint8_t)(unsafe.Pointer(&pcm[0])), C.uint32_t(len(pcm)))
	})
	return true
}

// SetDualSenseFullAudioControlCallback sets a callback to be invoked when the host sends audio control (volume/audio control/microphone mute) commands to the device.
// @param handle Handle to the DualSense device with audio.
// @param callback Callback receiving the audio control state. Pass NULL to clear.
//
//export SetDualSenseFullAudioControlCallback
func SetDualSenseFullAudioControlCallback(handle C.DSDeviceHandle, cb C.DSAudioControlCallback) bool {
	dsDevice := dualSenseFullDevice(handle)
	if dsDevice == nil {
		return false
	}
	if cb == nil {
		dsDevice.SetOutputCallback(nil)
		return true
	}
	dsDevice.SetOutputCallback(func(out dualsensefull.OutputState) {
		C.viiper_call_ds_audio_control(cb, handle, C.DSAudioControlState{
			UpdateFlags1:     C.uint8_t(out.UpdateFlags1),
			UpdateFlags2:     C.uint8_t(out.UpdateFlags2),
			SpeakerVolume:    C.uint8_t(out.SpeakerVolume),
			MicrophoneVolume: C.uint8_t(out.MicrophoneVolume),
			AudioControl:     C.uint8_t(out.AudioControl),
			MicMuteLed:       C.uint8_t(out.MicMuteLed),
			PowerSaveControl: C.uint8_t(out.PowerSaveControl),
			AudioControl2:    C.uint8_t(out.AudioControl2),
		})
	})
	return true
}
