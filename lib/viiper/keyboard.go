package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef uintptr_t USBServerHandle;

typedef uintptr_t KeyboardDeviceHandle;

#define KB_MOD_LEFT_CTRL   0x01u
#define KB_MOD_LEFT_SHIFT  0x02u
#define KB_MOD_LEFT_ALT    0x04u
#define KB_MOD_LEFT_GUI    0x08u
#define KB_MOD_RIGHT_CTRL  0x10u
#define KB_MOD_RIGHT_SHIFT 0x20u
#define KB_MOD_RIGHT_ALT   0x40u
#define KB_MOD_RIGHT_GUI   0x80u

#define KB_LED_NUM_LOCK    0x01u
#define KB_LED_CAPS_LOCK   0x02u
#define KB_LED_SCROLL_LOCK 0x04u
#define KB_LED_COMPOSE     0x08u
#define KB_LED_KANA        0x10u

#define KB_KEY_A 0x04u
#define KB_KEY_B 0x05u
#define KB_KEY_C 0x06u
#define KB_KEY_D 0x07u
#define KB_KEY_E 0x08u
#define KB_KEY_F 0x09u
#define KB_KEY_G 0x0Au
#define KB_KEY_H 0x0Bu
#define KB_KEY_I 0x0Cu
#define KB_KEY_J 0x0Du
#define KB_KEY_K 0x0Eu
#define KB_KEY_L 0x0Fu
#define KB_KEY_M 0x10u
#define KB_KEY_N 0x11u
#define KB_KEY_O 0x12u
#define KB_KEY_P 0x13u
#define KB_KEY_Q 0x14u
#define KB_KEY_R 0x15u
#define KB_KEY_S 0x16u
#define KB_KEY_T 0x17u
#define KB_KEY_U 0x18u
#define KB_KEY_V 0x19u
#define KB_KEY_W 0x1Au
#define KB_KEY_X 0x1Bu
#define KB_KEY_Y 0x1Cu
#define KB_KEY_Z 0x1Du
#define KB_KEY_1 0x1Eu
#define KB_KEY_2 0x1Fu
#define KB_KEY_3 0x20u
#define KB_KEY_4 0x21u
#define KB_KEY_5 0x22u
#define KB_KEY_6 0x23u
#define KB_KEY_7 0x24u
#define KB_KEY_8 0x25u
#define KB_KEY_9 0x26u
#define KB_KEY_0 0x27u
#define KB_KEY_ENTER      0x28u
#define KB_KEY_ESCAPE     0x29u
#define KB_KEY_BACKSPACE  0x2Au
#define KB_KEY_TAB        0x2Bu
#define KB_KEY_SPACE      0x2Cu
#define KB_KEY_MINUS      0x2Du
#define KB_KEY_EQUAL      0x2Eu
#define KB_KEY_LEFT_BRACE  0x2Fu
#define KB_KEY_RIGHT_BRACE 0x30u
#define KB_KEY_BACKSLASH  0x31u
#define KB_KEY_SEMICOLON  0x33u
#define KB_KEY_APOSTROPHE 0x34u
#define KB_KEY_GRAVE      0x35u
#define KB_KEY_COMMA      0x36u
#define KB_KEY_PERIOD     0x37u
#define KB_KEY_SLASH      0x38u
#define KB_KEY_CAPS_LOCK  0x39u
#define KB_KEY_F1  0x3Au
#define KB_KEY_F2  0x3Bu
#define KB_KEY_F3  0x3Cu
#define KB_KEY_F4  0x3Du
#define KB_KEY_F5  0x3Eu
#define KB_KEY_F6  0x3Fu
#define KB_KEY_F7  0x40u
#define KB_KEY_F8  0x41u
#define KB_KEY_F9  0x42u
#define KB_KEY_F10 0x43u
#define KB_KEY_F11 0x44u
#define KB_KEY_F12 0x45u
#define KB_KEY_PRINT_SCREEN 0x46u
#define KB_KEY_SCROLL_LOCK  0x47u
#define KB_KEY_PAUSE        0x48u
#define KB_KEY_INSERT       0x49u
#define KB_KEY_HOME         0x4Au
#define KB_KEY_PAGE_UP      0x4Bu
#define KB_KEY_DELETE       0x4Cu
#define KB_KEY_END          0x4Du
#define KB_KEY_PAGE_DOWN    0x4Eu
#define KB_KEY_RIGHT        0x4Fu
#define KB_KEY_LEFT         0x50u
#define KB_KEY_DOWN         0x51u
#define KB_KEY_UP           0x52u
#define KB_KEY_NUM_LOCK     0x53u
#define KB_KEY_KP_SLASH     0x54u
#define KB_KEY_KP_ASTERISK  0x55u
#define KB_KEY_KP_MINUS     0x56u
#define KB_KEY_KP_PLUS      0x57u
#define KB_KEY_KP_ENTER     0x58u
#define KB_KEY_KP_1 0x59u
#define KB_KEY_KP_2 0x5Au
#define KB_KEY_KP_3 0x5Bu
#define KB_KEY_KP_4 0x5Cu
#define KB_KEY_KP_5 0x5Du
#define KB_KEY_KP_6 0x5Eu
#define KB_KEY_KP_7 0x5Fu
#define KB_KEY_KP_8 0x60u
#define KB_KEY_KP_9 0x61u
#define KB_KEY_KP_0 0x62u
#define KB_KEY_KP_DOT 0x63u
#define KB_KEY_MUTE        0x7Fu
#define KB_KEY_VOLUME_UP   0x80u
#define KB_KEY_VOLUME_DOWN 0x81u

typedef struct {
	uint8_t Modifiers;
	uint8_t KeyBitmap[32];
} KeyboardDeviceState;

typedef void (*KeyboardLEDCallback)(KeyboardDeviceHandle handle, uint8_t leds);

static void viiper_call_kb_led(KeyboardLEDCallback fn, KeyboardDeviceHandle handle, uint8_t leds) {
	fn(handle, leds);
}

*/
import "C"
import (
	"context"
	"fmt"
	"log/slog"
	"runtime/cgo"
	"slices"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/keyboard"
	"github.com/Alia5/VIIPER/internal/server/api"
)

// CreateKeyboardDevice creates a new HID keyboard device on the bus with the given ID on the server associated with the given handle.
// @param serverHandle Handle to the USB server.
// @param outDeviceHandle Output parameter for the created device handle.
// @param busID ID of the bus to add the device to.
// @param autoAttachLocalhost If true, the device will be automatically attached to a USBIP-Client/Driver running on THIS machine.
// @param idVendor Optional USB vendor ID (0 = default).
// @param idProduct Optional USB product ID (0 = default).
//
//export CreateKeyboardDevice
func CreateKeyboardDevice(
	serverHandle C.USBServerHandle,
	outDeviceHandle *C.KeyboardDeviceHandle,
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

	d, err := keyboard.New(opts)
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
	*outDeviceHandle = C.KeyboardDeviceHandle(cgo.NewHandle(handleWrapper))

	shw.mtx.Lock()
	defer shw.mtx.Unlock()
	shw.deviceHandles[busID] = append(shw.deviceHandles[busID], deviceHandle(*outDeviceHandle))
	return true
}

// SetKeyboardDeviceState updates the input state of the keyboard device associated with the given handle.
// @param handle Handle to the keyboard device.
// @param state New input state (Modifiers bitmask + 256-bit key bitmap).
//
//export SetKeyboardDeviceState
func SetKeyboardDeviceState(handle C.KeyboardDeviceHandle, state C.KeyboardDeviceState) bool {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return false
	}
	kbDevice, ok := dhw.device.(*keyboard.Keyboard)
	if !ok {
		return false
	}
	s := keyboard.InputState{
		Modifiers: uint8(state.Modifiers),
	}
	for i, v := range state.KeyBitmap {
		s.KeyBitmap[i] = byte(v)
	}
	kbDevice.UpdateInputState(s)
	return true
}

// SetKeyboardLEDCallback sets a callback to be invoked when the host changes keyboard LED state.
// @param handle Handle to the keyboard device.
// @param callback Callback receiving the raw LED bitmask byte (KB_LED_* flags). Pass NULL to clear.
//
//export SetKeyboardLEDCallback
func SetKeyboardLEDCallback(handle C.KeyboardDeviceHandle, cb C.KeyboardLEDCallback) bool {
	dh := cgo.Handle(handle)
	dhw, ok := dh.Value().(*deviceHandleWrapper)
	if !ok {
		return false
	}
	kbDevice, ok := dhw.device.(*keyboard.Keyboard)
	if !ok {
		return false
	}
	if cb == nil {
		kbDevice.SetLEDCallback(nil)
		return true
	}
	kbDevice.SetLEDCallback(func(led keyboard.LEDState) {
		var raw C.uint8_t
		if led.NumLock {
			raw |= 0x01
		}
		if led.CapsLock {
			raw |= 0x02
		}
		if led.ScrollLock {
			raw |= 0x04
		}
		if led.Compose {
			raw |= 0x08
		}
		if led.Kana {
			raw |= 0x10
		}
		C.viiper_call_kb_led(cb, handle, raw)
	})
	return true
}

// RemoveKeyboardDevice removes the keyboard device associated with the given handle from the server.
// @param handle Handle to the keyboard device to remove.
//
//export RemoveKeyboardDevice
func RemoveKeyboardDevice(handle C.KeyboardDeviceHandle) bool {
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
