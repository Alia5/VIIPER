package sdl

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../_testing/e2e/deps/SDL/include

#include <stdlib.h>

#include <SDL3/SDL_events.h>
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

// BaseEvent contains fields shared by every event.
type BaseEvent struct {
	Type      EventType
	Timestamp uint64
}

// Event is the common interface for all wrapped SDL events.
type Event interface {
	Base() BaseEvent
}

// ---

// PollEvent polls for currently pending events.
//
// It returns false if there are no events available.
func PollEvent() (Event, bool) {
	var cEvent C.SDL_Event
	if !C.SDL_PollEvent(&cEvent) {
		return nil, false
	}

	return wrapEvent(&cEvent), true
}

// WaitEvent waits indefinitely for the next available event.
func WaitEvent() (Event, error) {
	var cEvent C.SDL_Event
	if !C.SDL_WaitEvent(&cEvent) {
		return nil, GetError()
	}

	return wrapEvent(&cEvent), nil
}

// WaitEventTimeout waits until the specified timeout for the next available event.
func WaitEventTimeout(timeout time.Duration) (Event, bool) {
	var cEvent C.SDL_Event
	if !C.SDL_WaitEventTimeout(&cEvent, C.int(timeout.Milliseconds())) {
		err := GetError()
		if err != nil {
			return nil, false
		}
		return nil, false
	}

	return wrapEvent(&cEvent), true
}

// ---

// EventType specifies the types of events that can be delivered.
type EventType uint32

func (t EventType) String() string {
	if name, ok := eventTypeNames[t]; ok {
		return name
	}

	return fmt.Sprintf("EventType(%d)", uint32(t))
}

// Event type constants.
const (
	EventTypeFirst                      EventType = C.SDL_EVENT_FIRST
	EventTypeQuit                       EventType = C.SDL_EVENT_QUIT
	EventTypeTerminating                EventType = C.SDL_EVENT_TERMINATING
	EventTypeLowMemory                  EventType = C.SDL_EVENT_LOW_MEMORY
	EventTypeWillEnterBackground        EventType = C.SDL_EVENT_WILL_ENTER_BACKGROUND
	EventTypeDidEnterBackground         EventType = C.SDL_EVENT_DID_ENTER_BACKGROUND
	EventTypeWillEnterForeground        EventType = C.SDL_EVENT_WILL_ENTER_FOREGROUND
	EventTypeDidEnterForeground         EventType = C.SDL_EVENT_DID_ENTER_FOREGROUND
	EventTypeLocaleChanged              EventType = C.SDL_EVENT_LOCALE_CHANGED
	EventTypeSystemThemeChanged         EventType = C.SDL_EVENT_SYSTEM_THEME_CHANGED
	EventTypeDisplayOrientation         EventType = C.SDL_EVENT_DISPLAY_ORIENTATION
	EventTypeDisplayAdded               EventType = C.SDL_EVENT_DISPLAY_ADDED
	EventTypeDisplayRemoved             EventType = C.SDL_EVENT_DISPLAY_REMOVED
	EventTypeDisplayMoved               EventType = C.SDL_EVENT_DISPLAY_MOVED
	EventTypeDisplayDesktopModeChanged  EventType = C.SDL_EVENT_DISPLAY_DESKTOP_MODE_CHANGED
	EventTypeDisplayCurrentModeChanged  EventType = C.SDL_EVENT_DISPLAY_CURRENT_MODE_CHANGED
	EventTypeDisplayContentScaleChanged EventType = C.SDL_EVENT_DISPLAY_CONTENT_SCALE_CHANGED
	EventTypeDisplayUsableBoundsChanged EventType = C.SDL_EVENT_DISPLAY_USABLE_BOUNDS_CHANGED
	EventTypeWindowShown                EventType = C.SDL_EVENT_WINDOW_SHOWN
	EventTypeWindowHidden               EventType = C.SDL_EVENT_WINDOW_HIDDEN
	EventTypeWindowExposed              EventType = C.SDL_EVENT_WINDOW_EXPOSED
	EventTypeWindowMoved                EventType = C.SDL_EVENT_WINDOW_MOVED
	EventTypeWindowResized              EventType = C.SDL_EVENT_WINDOW_RESIZED
	EventTypeWindowPixelSizeChanged     EventType = C.SDL_EVENT_WINDOW_PIXEL_SIZE_CHANGED
	EventTypeWindowMetalViewResized     EventType = C.SDL_EVENT_WINDOW_METAL_VIEW_RESIZED
	EventTypeWindowMinimized            EventType = C.SDL_EVENT_WINDOW_MINIMIZED
	EventTypeWindowMaximized            EventType = C.SDL_EVENT_WINDOW_MAXIMIZED
	EventTypeWindowRestored             EventType = C.SDL_EVENT_WINDOW_RESTORED
	EventTypeWindowMouseEnter           EventType = C.SDL_EVENT_WINDOW_MOUSE_ENTER
	EventTypeWindowMouseLeave           EventType = C.SDL_EVENT_WINDOW_MOUSE_LEAVE
	EventTypeWindowFocusGained          EventType = C.SDL_EVENT_WINDOW_FOCUS_GAINED
	EventTypeWindowFocusLost            EventType = C.SDL_EVENT_WINDOW_FOCUS_LOST
	EventTypeWindowCloseRequested       EventType = C.SDL_EVENT_WINDOW_CLOSE_REQUESTED
	EventTypeWindowHitTest              EventType = C.SDL_EVENT_WINDOW_HIT_TEST
	EventTypeWindowICCProfChanged       EventType = C.SDL_EVENT_WINDOW_ICCPROF_CHANGED
	EventTypeWindowDisplayChanged       EventType = C.SDL_EVENT_WINDOW_DISPLAY_CHANGED
	EventTypeWindowDisplayScaleChanged  EventType = C.SDL_EVENT_WINDOW_DISPLAY_SCALE_CHANGED
	EventTypeWindowSafeAreaChanged      EventType = C.SDL_EVENT_WINDOW_SAFE_AREA_CHANGED
	EventTypeWindowOccluded             EventType = C.SDL_EVENT_WINDOW_OCCLUDED
	EventTypeWindowEnterFullscreen      EventType = C.SDL_EVENT_WINDOW_ENTER_FULLSCREEN
	EventTypeWindowLeaveFullscreen      EventType = C.SDL_EVENT_WINDOW_LEAVE_FULLSCREEN
	EventTypeWindowDestroyed            EventType = C.SDL_EVENT_WINDOW_DESTROYED
	EventTypeWindowHDRStateChanged      EventType = C.SDL_EVENT_WINDOW_HDR_STATE_CHANGED
	EventTypeKeyDown                    EventType = C.SDL_EVENT_KEY_DOWN
	EventTypeKeyUp                      EventType = C.SDL_EVENT_KEY_UP
	EventTypeTextEditing                EventType = C.SDL_EVENT_TEXT_EDITING
	EventTypeTextInput                  EventType = C.SDL_EVENT_TEXT_INPUT
	EventTypeKeymapChanged              EventType = C.SDL_EVENT_KEYMAP_CHANGED
	EventTypeKeyboardAdded              EventType = C.SDL_EVENT_KEYBOARD_ADDED
	EventTypeKeyboardRemoved            EventType = C.SDL_EVENT_KEYBOARD_REMOVED
	EventTypeTextEditingCandidates      EventType = C.SDL_EVENT_TEXT_EDITING_CANDIDATES
	EventTypeScreenKeyboardShown        EventType = C.SDL_EVENT_SCREEN_KEYBOARD_SHOWN
	EventTypeScreenKeyboardHidden       EventType = C.SDL_EVENT_SCREEN_KEYBOARD_HIDDEN
	EventTypeMouseMotion                EventType = C.SDL_EVENT_MOUSE_MOTION
	EventTypeMouseButtonDown            EventType = C.SDL_EVENT_MOUSE_BUTTON_DOWN
	EventTypeMouseButtonUp              EventType = C.SDL_EVENT_MOUSE_BUTTON_UP
	EventTypeMouseWheel                 EventType = C.SDL_EVENT_MOUSE_WHEEL
	EventTypeMouseAdded                 EventType = C.SDL_EVENT_MOUSE_ADDED
	EventTypeMouseRemoved               EventType = C.SDL_EVENT_MOUSE_REMOVED
	EventTypeJoystickAxisMotion         EventType = C.SDL_EVENT_JOYSTICK_AXIS_MOTION
	EventTypeJoystickBallMotion         EventType = C.SDL_EVENT_JOYSTICK_BALL_MOTION
	EventTypeJoystickHatMotion          EventType = C.SDL_EVENT_JOYSTICK_HAT_MOTION
	EventTypeJoystickButtonDown         EventType = C.SDL_EVENT_JOYSTICK_BUTTON_DOWN
	EventTypeJoystickButtonUp           EventType = C.SDL_EVENT_JOYSTICK_BUTTON_UP
	EventTypeJoystickAdded              EventType = C.SDL_EVENT_JOYSTICK_ADDED
	EventTypeJoystickRemoved            EventType = C.SDL_EVENT_JOYSTICK_REMOVED
	EventTypeJoystickBatteryUpdated     EventType = C.SDL_EVENT_JOYSTICK_BATTERY_UPDATED
	EventTypeJoystickUpdateComplete     EventType = C.SDL_EVENT_JOYSTICK_UPDATE_COMPLETE
	EventTypeGamepadAxisMotion          EventType = C.SDL_EVENT_GAMEPAD_AXIS_MOTION
	EventTypeGamepadButtonDown          EventType = C.SDL_EVENT_GAMEPAD_BUTTON_DOWN
	EventTypeGamepadButtonUp            EventType = C.SDL_EVENT_GAMEPAD_BUTTON_UP
	EventTypeGamepadAdded               EventType = C.SDL_EVENT_GAMEPAD_ADDED
	EventTypeGamepadRemoved             EventType = C.SDL_EVENT_GAMEPAD_REMOVED
	EventTypeGamepadRemapped            EventType = C.SDL_EVENT_GAMEPAD_REMAPPED
	EventTypeGamepadTouchpadDown        EventType = C.SDL_EVENT_GAMEPAD_TOUCHPAD_DOWN
	EventTypeGamepadTouchpadMotion      EventType = C.SDL_EVENT_GAMEPAD_TOUCHPAD_MOTION
	EventTypeGamepadTouchpadUp          EventType = C.SDL_EVENT_GAMEPAD_TOUCHPAD_UP
	EventTypeGamepadSensorUpdate        EventType = C.SDL_EVENT_GAMEPAD_SENSOR_UPDATE
	EventTypeGamepadUpdateComplete      EventType = C.SDL_EVENT_GAMEPAD_UPDATE_COMPLETE
	EventTypeGamepadSteamHandleUpdated  EventType = C.SDL_EVENT_GAMEPAD_STEAM_HANDLE_UPDATED
	EventTypeFingerDown                 EventType = C.SDL_EVENT_FINGER_DOWN
	EventTypeFingerUp                   EventType = C.SDL_EVENT_FINGER_UP
	EventTypeFingerMotion               EventType = C.SDL_EVENT_FINGER_MOTION
	EventTypeFingerCanceled             EventType = C.SDL_EVENT_FINGER_CANCELED
	EventTypePinchBegin                 EventType = C.SDL_EVENT_PINCH_BEGIN
	EventTypePinchUpdate                EventType = C.SDL_EVENT_PINCH_UPDATE
	EventTypePinchEnd                   EventType = C.SDL_EVENT_PINCH_END
	EventTypeClipboardUpdate            EventType = C.SDL_EVENT_CLIPBOARD_UPDATE
	EventTypeDropFile                   EventType = C.SDL_EVENT_DROP_FILE
	EventTypeDropText                   EventType = C.SDL_EVENT_DROP_TEXT
	EventTypeDropBegin                  EventType = C.SDL_EVENT_DROP_BEGIN
	EventTypeDropComplete               EventType = C.SDL_EVENT_DROP_COMPLETE
	EventTypeDropPosition               EventType = C.SDL_EVENT_DROP_POSITION
	EventTypeAudioDeviceAdded           EventType = C.SDL_EVENT_AUDIO_DEVICE_ADDED
	EventTypeAudioDeviceRemoved         EventType = C.SDL_EVENT_AUDIO_DEVICE_REMOVED
	EventTypeAudioDeviceFormatChanged   EventType = C.SDL_EVENT_AUDIO_DEVICE_FORMAT_CHANGED
	EventTypeSensorUpdate               EventType = C.SDL_EVENT_SENSOR_UPDATE
	EventTypePenProximityIn             EventType = C.SDL_EVENT_PEN_PROXIMITY_IN
	EventTypePenProximityOut            EventType = C.SDL_EVENT_PEN_PROXIMITY_OUT
	EventTypePenDown                    EventType = C.SDL_EVENT_PEN_DOWN
	EventTypePenUp                      EventType = C.SDL_EVENT_PEN_UP
	EventTypePenButtonDown              EventType = C.SDL_EVENT_PEN_BUTTON_DOWN
	EventTypePenButtonUp                EventType = C.SDL_EVENT_PEN_BUTTON_UP
	EventTypePenMotion                  EventType = C.SDL_EVENT_PEN_MOTION
	EventTypePenAxis                    EventType = C.SDL_EVENT_PEN_AXIS
	EventTypeCameraDeviceAdded          EventType = C.SDL_EVENT_CAMERA_DEVICE_ADDED
	EventTypeCameraDeviceRemoved        EventType = C.SDL_EVENT_CAMERA_DEVICE_REMOVED
	EventTypeCameraDeviceApproved       EventType = C.SDL_EVENT_CAMERA_DEVICE_APPROVED
	EventTypeCameraDeviceDenied         EventType = C.SDL_EVENT_CAMERA_DEVICE_DENIED
	EventTypeRenderTargetsReset         EventType = C.SDL_EVENT_RENDER_TARGETS_RESET
	EventTypeRenderDeviceReset          EventType = C.SDL_EVENT_RENDER_DEVICE_RESET
	EventTypeRenderDeviceLost           EventType = C.SDL_EVENT_RENDER_DEVICE_LOST
	EventTypePrivate0                   EventType = C.SDL_EVENT_PRIVATE0
	EventTypePrivate1                   EventType = C.SDL_EVENT_PRIVATE1
	EventTypePrivate2                   EventType = C.SDL_EVENT_PRIVATE2
	EventTypePrivate3                   EventType = C.SDL_EVENT_PRIVATE3
	EventTypePollSentinel               EventType = C.SDL_EVENT_POLL_SENTINEL
	EventTypeUser                       EventType = C.SDL_EVENT_USER
	EventTypeLast                       EventType = C.SDL_EVENT_LAST
	EventTypeEnumPadding                EventType = C.SDL_EVENT_ENUM_PADDING
)

var eventTypeNames = map[EventType]string{
	EventTypeFirst:                      "First",
	EventTypeQuit:                       "Quit",
	EventTypeTerminating:                "Terminating",
	EventTypeLowMemory:                  "LowMemory",
	EventTypeWillEnterBackground:        "WillEnterBackground",
	EventTypeDidEnterBackground:         "DidEnterBackground",
	EventTypeWillEnterForeground:        "WillEnterForeground",
	EventTypeDidEnterForeground:         "DidEnterForeground",
	EventTypeLocaleChanged:              "LocaleChanged",
	EventTypeSystemThemeChanged:         "SystemThemeChanged",
	EventTypeDisplayOrientation:         "DisplayOrientation",
	EventTypeDisplayAdded:               "DisplayAdded",
	EventTypeDisplayRemoved:             "DisplayRemoved",
	EventTypeDisplayMoved:               "DisplayMoved",
	EventTypeDisplayDesktopModeChanged:  "DisplayDesktopModeChanged",
	EventTypeDisplayCurrentModeChanged:  "DisplayCurrentModeChanged",
	EventTypeDisplayContentScaleChanged: "DisplayContentScaleChanged",
	EventTypeDisplayUsableBoundsChanged: "DisplayUsableBoundsChanged",
	EventTypeWindowShown:                "WindowShown",
	EventTypeWindowHidden:               "WindowHidden",
	EventTypeWindowExposed:              "WindowExposed",
	EventTypeWindowMoved:                "WindowMoved",
	EventTypeWindowResized:              "WindowResized",
	EventTypeWindowPixelSizeChanged:     "WindowPixelSizeChanged",
	EventTypeWindowMetalViewResized:     "WindowMetalViewResized",
	EventTypeWindowMinimized:            "WindowMinimized",
	EventTypeWindowMaximized:            "WindowMaximized",
	EventTypeWindowRestored:             "WindowRestored",
	EventTypeWindowMouseEnter:           "WindowMouseEnter",
	EventTypeWindowMouseLeave:           "WindowMouseLeave",
	EventTypeWindowFocusGained:          "WindowFocusGained",
	EventTypeWindowFocusLost:            "WindowFocusLost",
	EventTypeWindowCloseRequested:       "WindowCloseRequested",
	EventTypeWindowHitTest:              "WindowHitTest",
	EventTypeWindowICCProfChanged:       "WindowICCProfChanged",
	EventTypeWindowDisplayChanged:       "WindowDisplayChanged",
	EventTypeWindowDisplayScaleChanged:  "WindowDisplayScaleChanged",
	EventTypeWindowSafeAreaChanged:      "WindowSafeAreaChanged",
	EventTypeWindowOccluded:             "WindowOccluded",
	EventTypeWindowEnterFullscreen:      "WindowEnterFullscreen",
	EventTypeWindowLeaveFullscreen:      "WindowLeaveFullscreen",
	EventTypeWindowDestroyed:            "WindowDestroyed",
	EventTypeWindowHDRStateChanged:      "WindowHDRStateChanged",
	EventTypeKeyDown:                    "KeyDown",
	EventTypeKeyUp:                      "KeyUp",
	EventTypeTextEditing:                "TextEditing",
	EventTypeTextInput:                  "TextInput",
	EventTypeKeymapChanged:              "KeymapChanged",
	EventTypeKeyboardAdded:              "KeyboardAdded",
	EventTypeKeyboardRemoved:            "KeyboardRemoved",
	EventTypeTextEditingCandidates:      "TextEditingCandidates",
	EventTypeScreenKeyboardShown:        "ScreenKeyboardShown",
	EventTypeScreenKeyboardHidden:       "ScreenKeyboardHidden",
	EventTypeMouseMotion:                "MouseMotion",
	EventTypeMouseButtonDown:            "MouseButtonDown",
	EventTypeMouseButtonUp:              "MouseButtonUp",
	EventTypeMouseWheel:                 "MouseWheel",
	EventTypeMouseAdded:                 "MouseAdded",
	EventTypeMouseRemoved:               "MouseRemoved",
	EventTypeJoystickAxisMotion:         "JoystickAxisMotion",
	EventTypeJoystickBallMotion:         "JoystickBallMotion",
	EventTypeJoystickHatMotion:          "JoystickHatMotion",
	EventTypeJoystickButtonDown:         "JoystickButtonDown",
	EventTypeJoystickButtonUp:           "JoystickButtonUp",
	EventTypeJoystickAdded:              "JoystickAdded",
	EventTypeJoystickRemoved:            "JoystickRemoved",
	EventTypeJoystickBatteryUpdated:     "JoystickBatteryUpdated",
	EventTypeJoystickUpdateComplete:     "JoystickUpdateComplete",
	EventTypeGamepadAxisMotion:          "GamepadAxisMotion",
	EventTypeGamepadButtonDown:          "GamepadButtonDown",
	EventTypeGamepadButtonUp:            "GamepadButtonUp",
	EventTypeGamepadAdded:               "GamepadAdded",
	EventTypeGamepadRemoved:             "GamepadRemoved",
	EventTypeGamepadRemapped:            "GamepadRemapped",
	EventTypeGamepadTouchpadDown:        "GamepadTouchpadDown",
	EventTypeGamepadTouchpadMotion:      "GamepadTouchpadMotion",
	EventTypeGamepadTouchpadUp:          "GamepadTouchpadUp",
	EventTypeGamepadSensorUpdate:        "GamepadSensorUpdate",
	EventTypeGamepadUpdateComplete:      "GamepadUpdateComplete",
	EventTypeGamepadSteamHandleUpdated:  "GamepadSteamHandleUpdated",
	EventTypeFingerDown:                 "FingerDown",
	EventTypeFingerUp:                   "FingerUp",
	EventTypeFingerMotion:               "FingerMotion",
	EventTypeFingerCanceled:             "FingerCanceled",
	EventTypePinchBegin:                 "PinchBegin",
	EventTypePinchUpdate:                "PinchUpdate",
	EventTypePinchEnd:                   "PinchEnd",
	EventTypeClipboardUpdate:            "ClipboardUpdate",
	EventTypeDropFile:                   "DropFile",
	EventTypeDropText:                   "DropText",
	EventTypeDropBegin:                  "DropBegin",
	EventTypeDropComplete:               "DropComplete",
	EventTypeDropPosition:               "DropPosition",
	EventTypeAudioDeviceAdded:           "AudioDeviceAdded",
	EventTypeAudioDeviceRemoved:         "AudioDeviceRemoved",
	EventTypeAudioDeviceFormatChanged:   "AudioDeviceFormatChanged",
	EventTypeSensorUpdate:               "SensorUpdate",
	EventTypePenProximityIn:             "PenProximityIn",
	EventTypePenProximityOut:            "PenProximityOut",
	EventTypePenDown:                    "PenDown",
	EventTypePenUp:                      "PenUp",
	EventTypePenButtonDown:              "PenButtonDown",
	EventTypePenButtonUp:                "PenButtonUp",
	EventTypePenMotion:                  "PenMotion",
	EventTypePenAxis:                    "PenAxis",
	EventTypeCameraDeviceAdded:          "CameraDeviceAdded",
	EventTypeCameraDeviceRemoved:        "CameraDeviceRemoved",
	EventTypeCameraDeviceApproved:       "CameraDeviceApproved",
	EventTypeCameraDeviceDenied:         "CameraDeviceDenied",
	EventTypeRenderTargetsReset:         "RenderTargetsReset",
	EventTypeRenderDeviceReset:          "RenderDeviceReset",
	EventTypeRenderDeviceLost:           "RenderDeviceLost",
	EventTypePrivate0:                   "Private0",
	EventTypePrivate1:                   "Private1",
	EventTypePrivate2:                   "Private2",
	EventTypePrivate3:                   "Private3",
	EventTypePollSentinel:               "PollSentinel",
	EventTypeUser:                       "User",
	EventTypeLast:                       "Last",
	EventTypeEnumPadding:                "EnumPadding",
}

type wrapperFn func(*C.SDL_Event) Event

var eventWrappers = map[EventType]wrapperFn{
	EventTypeFirst: wrapBasic,

	EventTypeQuit:                       wrapQuit,
	EventTypeTerminating:                wrapBasic,
	EventTypeLowMemory:                  wrapBasic,
	EventTypeWillEnterBackground:        wrapBasic,
	EventTypeDidEnterBackground:         wrapBasic,
	EventTypeWillEnterForeground:        wrapBasic,
	EventTypeDidEnterForeground:         wrapBasic,
	EventTypeLocaleChanged:              wrapBasic,
	EventTypeSystemThemeChanged:         wrapBasic,
	EventTypeDisplayOrientation:         wrapDisplay,
	EventTypeDisplayAdded:               wrapDisplay,
	EventTypeDisplayRemoved:             wrapDisplay,
	EventTypeDisplayMoved:               wrapDisplay,
	EventTypeDisplayDesktopModeChanged:  wrapDisplay,
	EventTypeDisplayCurrentModeChanged:  wrapDisplay,
	EventTypeDisplayContentScaleChanged: wrapDisplay,
	EventTypeDisplayUsableBoundsChanged: wrapDisplay,
	EventTypeWindowShown:                wrapWindow,
	EventTypeWindowHidden:               wrapWindow,
	EventTypeWindowExposed:              wrapWindow,
	EventTypeWindowMoved:                wrapWindow,
	EventTypeWindowResized:              wrapWindow,
	EventTypeWindowPixelSizeChanged:     wrapWindow,
	EventTypeWindowMetalViewResized:     wrapWindow,
	EventTypeWindowMinimized:            wrapWindow,
	EventTypeWindowMaximized:            wrapWindow,
	EventTypeWindowRestored:             wrapWindow,
	EventTypeWindowMouseEnter:           wrapWindow,
	EventTypeWindowMouseLeave:           wrapWindow,
	EventTypeKeyDown:                    wrapKeyboard,
	EventTypeKeyUp:                      wrapKeyboard,
	EventTypeTextEditing:                wrapTextEditing,
	EventTypeTextInput:                  wrapTextInput,
	EventTypeKeymapChanged:              wrapBasic,
	EventTypeKeyboardAdded:              wrapKeyboardDevice,
	EventTypeKeyboardRemoved:            wrapKeyboardDevice,
	EventTypeTextEditingCandidates:      wrapTextEditingCandidates,
	EventTypeScreenKeyboardShown:        wrapBasic,
	EventTypeScreenKeyboardHidden:       wrapBasic,
	EventTypeMouseMotion:                wrapMouseMotion,
	EventTypeMouseButtonDown:            wrapMouseButton,
	EventTypeMouseButtonUp:              wrapMouseButton,
	EventTypeMouseWheel:                 wrapMouseWheel,
	EventTypeMouseAdded:                 wrapMouseDevice,
	EventTypeMouseRemoved:               wrapMouseDevice,
	EventTypeJoystickAxisMotion:         wrapJoyAxis,
	EventTypeJoystickBallMotion:         wrapJoyBall,
	EventTypeJoystickHatMotion:          wrapJoyHat,
	EventTypeJoystickButtonDown:         wrapJoyButton,
	EventTypeJoystickButtonUp:           wrapJoyButton,
	EventTypeJoystickAdded:              wrapJoyDevice,
	EventTypeJoystickRemoved:            wrapJoyDevice,
	EventTypeJoystickBatteryUpdated:     wrapJoyBattery,
	EventTypeJoystickUpdateComplete:     wrapJoyDevice,
	EventTypeGamepadAxisMotion:          wrapGamepadAxis,
	EventTypeGamepadButtonDown:          wrapGamepadButton,
	EventTypeGamepadButtonUp:            wrapGamepadButton,
	EventTypeWindowFocusGained:          wrapWindow,
	EventTypeWindowFocusLost:            wrapWindow,
	EventTypeWindowCloseRequested:       wrapWindow,
	EventTypeWindowEnterFullscreen:      wrapWindow,
	EventTypeWindowLeaveFullscreen:      wrapWindow,
	EventTypeWindowDestroyed:            wrapWindow,
	EventTypeWindowHDRStateChanged:      wrapWindow,
	EventTypeGamepadSensorUpdate:        wrapGamepadSensor,
	EventTypeWindowHitTest:              wrapWindow,
	EventTypeWindowICCProfChanged:       wrapWindow,
	EventTypeWindowDisplayChanged:       wrapWindow,
	EventTypeWindowDisplayScaleChanged:  wrapWindow,
	EventTypeWindowSafeAreaChanged:      wrapWindow,
	EventTypeWindowOccluded:             wrapWindow,
	EventTypeGamepadAdded:               wrapGamepadDevice,
	EventTypeGamepadRemoved:             wrapGamepadDevice,
	EventTypeGamepadRemapped:            wrapGamepadDevice,
	EventTypeGamepadUpdateComplete:      wrapGamepadDevice,
	EventTypeGamepadSteamHandleUpdated:  wrapGamepadDevice,
	EventTypeGamepadTouchpadDown:        wrapGamepadTouchpad,
	EventTypeGamepadTouchpadMotion:      wrapGamepadTouchpad,
	EventTypeGamepadTouchpadUp:          wrapGamepadTouchpad,
	EventTypeFingerDown:                 wrapTouchFinger,
	EventTypeFingerUp:                   wrapTouchFinger,
	EventTypeFingerMotion:               wrapTouchFinger,
	EventTypeFingerCanceled:             wrapTouchFinger,
	EventTypePinchBegin:                 wrapPinch,
	EventTypePinchUpdate:                wrapPinch,
	EventTypePinchEnd:                   wrapPinch,
	EventTypeClipboardUpdate:            wrapClipboard,
	EventTypeDropFile:                   wrapDrop,
	EventTypeDropText:                   wrapDrop,
	EventTypeDropBegin:                  wrapDrop,
	EventTypeDropComplete:               wrapDrop,
	EventTypeDropPosition:               wrapDrop,
	EventTypeAudioDeviceAdded:           wrapAudioDevice,
	EventTypeAudioDeviceRemoved:         wrapAudioDevice,
	EventTypeAudioDeviceFormatChanged:   wrapAudioDevice,
	EventTypeSensorUpdate:               wrapSensor,
	EventTypePenProximityIn:             wrapPenProximity,
	EventTypePenProximityOut:            wrapPenProximity,
	EventTypePenDown:                    wrapPenTouch,
	EventTypePenUp:                      wrapPenTouch,
	EventTypePenButtonDown:              wrapPenButton,
	EventTypePenButtonUp:                wrapPenButton,
	EventTypePenMotion:                  wrapPenMotion,
	EventTypePenAxis:                    wrapPenAxis,
	EventTypeCameraDeviceAdded:          wrapCameraDevice,
	EventTypeCameraDeviceRemoved:        wrapCameraDevice,
	EventTypeCameraDeviceApproved:       wrapCameraDevice,
	EventTypeCameraDeviceDenied:         wrapCameraDevice,
	EventTypeRenderTargetsReset:         wrapRender,
	EventTypeRenderDeviceReset:          wrapRender,
	EventTypeRenderDeviceLost:           wrapRender,
	EventTypePrivate0:                   wrapBasic,
	EventTypePrivate1:                   wrapBasic,
	EventTypePrivate2:                   wrapBasic,
	EventTypePrivate3:                   wrapBasic,
	EventTypePollSentinel:               wrapBasic,
	EventTypeUser:                       wrapUser,
	EventTypeLast:                       wrapUser,
}

func wrapCommon(ev *C.SDL_Event) BaseEvent {
	c := (*C.SDL_CommonEvent)(unsafe.Pointer(ev))
	return BaseEvent{
		Type:      EventType(c._type),
		Timestamp: uint64(c.timestamp),
	}
}

func wrapQuit(ev *C.SDL_Event) Event {
	return &QuitEvent{BaseEvent: wrapCommon(ev)}
}

func wrapBasic(ev *C.SDL_Event) Event {
	return &BasicEvent{BaseEvent: wrapCommon(ev)}
}

func wrapDisplay(ev *C.SDL_Event) Event {
	d := (*C.SDL_DisplayEvent)(unsafe.Pointer(ev))
	return &DisplayEvent{
		BaseEvent: wrapCommon(ev),
		DisplayID: uint32(d.displayID),
		Data1:     int32(d.data1),
		Data2:     int32(d.data2),
	}
}

func wrapWindow(ev *C.SDL_Event) Event {
	w := (*C.SDL_WindowEvent)(unsafe.Pointer(ev))
	return &WindowEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(w.windowID),
		Data1:     int32(w.data1),
		Data2:     int32(w.data2),
	}
}

func wrapKeyboardDevice(ev *C.SDL_Event) Event {
	kd := (*C.SDL_KeyboardDeviceEvent)(unsafe.Pointer(ev))
	return &KeyboardDeviceEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(kd.which),
	}
}

func wrapKeyboard(ev *C.SDL_Event) Event {
	k := (*C.SDL_KeyboardEvent)(unsafe.Pointer(ev))
	return &KeyboardEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(k.windowID),
		Which:     uint32(k.which),
		Scancode:  uint32(k.scancode),
		Key:       KeyCode(k.key),
		Mod:       uint16(k.mod),
		Raw:       uint16(k.raw),
		Down:      bool(k.down),
		Repeat:    bool(k.repeat),
	}
}

func wrapTextEditing(ev *C.SDL_Event) Event {
	te := (*C.SDL_TextEditingEvent)(unsafe.Pointer(ev))
	return &TextEditingEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(te.windowID),
		Text:      unsafe.Pointer(te.text),
		Start:     int32(te.start),
		Length:    int32(te.length),
	}
}

func wrapTextEditingCandidates(ev *C.SDL_Event) Event {
	tec := (*C.SDL_TextEditingCandidatesEvent)(unsafe.Pointer(ev))
	return &TextEditingCandidatesEvent{
		BaseEvent:         wrapCommon(ev),
		WindowID:          uint32(tec.windowID),
		Candidates:        unsafe.Pointer(tec.candidates),
		NumCandidates:     int32(tec.num_candidates),
		SelectedCandidate: int32(tec.selected_candidate),
		Horizontal:        bool(tec.horizontal),
	}
}

func wrapTextInput(ev *C.SDL_Event) Event {
	ti := (*C.SDL_TextInputEvent)(unsafe.Pointer(ev))
	return &TextInputEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(ti.windowID),
		Text:      unsafe.Pointer(ti.text),
	}
}

func wrapMouseDevice(ev *C.SDL_Event) Event {
	md := (*C.SDL_MouseDeviceEvent)(unsafe.Pointer(ev))
	return &MouseDeviceEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(md.which),
	}
}

func wrapMouseMotion(ev *C.SDL_Event) Event {
	mm := (*C.SDL_MouseMotionEvent)(unsafe.Pointer(ev))
	return &MouseMotionEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(mm.windowID),
		Which:     uint32(mm.which),
		State:     uint32(mm.state),
		X:         float32(mm.x),
		Y:         float32(mm.y),
		XRel:      float32(mm.xrel),
		YRel:      float32(mm.yrel),
	}
}

func wrapMouseButton(ev *C.SDL_Event) Event {
	mb := (*C.SDL_MouseButtonEvent)(unsafe.Pointer(ev))
	return &MouseButtonEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(mb.windowID),
		Which:     uint32(mb.which),
		Button:    uint8(mb.button),
		Down:      bool(mb.down),
		Clicks:    uint8(mb.clicks),
		X:         float32(mb.x),
		Y:         float32(mb.y),
	}
}

func wrapMouseWheel(ev *C.SDL_Event) Event {
	mw := (*C.SDL_MouseWheelEvent)(unsafe.Pointer(ev))
	return &MouseWheelEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(mw.windowID),
		Which:     uint32(mw.which),
		X:         float32(mw.x),
		Y:         float32(mw.y),
		Direction: uint32(mw.direction),
		MouseX:    float32(mw.mouse_x),
		MouseY:    float32(mw.mouse_y),
		IntegerX:  int32(mw.integer_x),
		IntegerY:  int32(mw.integer_y),
	}
}

func wrapJoyAxis(ev *C.SDL_Event) Event {
	ja := (*C.SDL_JoyAxisEvent)(unsafe.Pointer(ev))
	return &JoyAxisEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(ja.which),
		Axis:      uint8(ja.axis),
		Value:     int16(ja.value),
	}
}

func wrapJoyBall(ev *C.SDL_Event) Event {
	jb := (*C.SDL_JoyBallEvent)(unsafe.Pointer(ev))
	return &JoyBallEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(jb.which),
		Ball:      uint8(jb.ball),
		XRel:      int16(jb.xrel),
		YRel:      int16(jb.yrel),
	}
}

func wrapJoyHat(ev *C.SDL_Event) Event {
	jh := (*C.SDL_JoyHatEvent)(unsafe.Pointer(ev))
	return &JoyHatEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(jh.which),
		Hat:       uint8(jh.hat),
		Value:     uint8(jh.value),
	}
}

func wrapJoyButton(ev *C.SDL_Event) Event {
	jb := (*C.SDL_JoyButtonEvent)(unsafe.Pointer(ev))
	return &JoyButtonEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(jb.which),
		Button:    uint8(jb.button),
		Down:      bool(jb.down),
	}
}

func wrapJoyDevice(ev *C.SDL_Event) Event {
	jd := (*C.SDL_JoyDeviceEvent)(unsafe.Pointer(ev))
	return &JoyDeviceEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(jd.which),
	}
}

func wrapJoyBattery(ev *C.SDL_Event) Event {
	jb := (*C.SDL_JoyBatteryEvent)(unsafe.Pointer(ev))
	return &JoyBatteryEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(jb.which),
		State:     int32(jb.state),
		Percent:   int32(jb.percent),
	}
}

func wrapGamepadAxis(ev *C.SDL_Event) Event {
	ga := (*C.SDL_GamepadAxisEvent)(unsafe.Pointer(ev))
	return &GamepadAxisEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(ga.which),
		Axis:      uint8(ga.axis),
		Value:     int16(ga.value),
	}
}

func wrapGamepadButton(ev *C.SDL_Event) Event {
	gb := (*C.SDL_GamepadButtonEvent)(unsafe.Pointer(ev))
	return &GamepadButtonEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(gb.which),
		Button:    uint8(gb.button),
		Down:      bool(gb.down),
	}
}

func wrapGamepadDevice(ev *C.SDL_Event) Event {
	gd := (*C.SDL_GamepadDeviceEvent)(unsafe.Pointer(ev))
	return &GamepadDeviceEvent{
		BaseEvent: wrapCommon(ev),
		Which:     int32(gd.which),
	}
}

func wrapGamepadTouchpad(ev *C.SDL_Event) Event {
	gt := (*C.SDL_GamepadTouchpadEvent)(unsafe.Pointer(ev))
	return &GamepadTouchpadEvent{
		BaseEvent: wrapCommon(ev),
		Which:     int32(gt.which),
		Touchpad:  int32(gt.touchpad),
		Finger:    int32(gt.finger),
		X:         float32(gt.x),
		Y:         float32(gt.y),
		Pressure:  float32(gt.pressure),
	}
}

func wrapGamepadSensor(ev *C.SDL_Event) Event {
	gs := (*C.SDL_GamepadSensorEvent)(unsafe.Pointer(ev))
	return &GamepadSensorEvent{
		BaseEvent:       wrapCommon(ev),
		Which:           uint32(gs.which),
		Sensor:          int32(gs.sensor),
		Data0:           float32(gs.data[0]),
		Data1:           float32(gs.data[1]),
		Data2:           float32(gs.data[2]),
		SensorTimestamp: uint64(gs.sensor_timestamp),
	}
}

func wrapAudioDevice(ev *C.SDL_Event) Event {
	ad := (*C.SDL_AudioDeviceEvent)(unsafe.Pointer(ev))
	return &AudioDeviceEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(ad.which),
		Recording: bool(ad.recording),
	}
}

func wrapCameraDevice(ev *C.SDL_Event) Event {
	cd := (*C.SDL_CameraDeviceEvent)(unsafe.Pointer(ev))
	return &CameraDeviceEvent{
		BaseEvent: wrapCommon(ev),
		Which:     uint32(cd.which),
	}
}

func wrapSensor(ev *C.SDL_Event) Event {
	s := (*C.SDL_SensorEvent)(unsafe.Pointer(ev))
	return &SensorEvent{
		BaseEvent:       wrapCommon(ev),
		Which:           uint32(s.which),
		Data0:           float32(s.data[0]),
		Data1:           float32(s.data[1]),
		Data2:           float32(s.data[2]),
		Data3:           float32(s.data[3]),
		Data4:           float32(s.data[4]),
		Data5:           float32(s.data[5]),
		SensorTimestamp: uint64(s.sensor_timestamp),
	}
}

func wrapTouchFinger(ev *C.SDL_Event) Event {
	t := (*C.SDL_TouchFingerEvent)(unsafe.Pointer(ev))
	return &TouchFingerEvent{
		BaseEvent: wrapCommon(ev),
		TouchID:   uint64(t.touchID),
		FingerID:  uint64(t.fingerID),
		X:         float32(t.x),
		Y:         float32(t.y),
		DX:        float32(t.dx),
		DY:        float32(t.dy),
		Pressure:  float32(t.pressure),
		WindowID:  uint32(t.windowID),
	}
}

func wrapPinch(ev *C.SDL_Event) Event {
	p := (*C.SDL_PinchFingerEvent)(unsafe.Pointer(ev))
	return &PinchFingerEvent{
		BaseEvent: wrapCommon(ev),
		Scale:     float32(p.scale),
		WindowID:  uint32(p.windowID),
	}
}

func wrapPenProximity(ev *C.SDL_Event) Event {
	pp := (*C.SDL_PenProximityEvent)(unsafe.Pointer(ev))
	return &PenProximityEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(pp.windowID),
		Which:     uint32(pp.which),
	}
}

func wrapPenTouch(ev *C.SDL_Event) Event {
	pt := (*C.SDL_PenTouchEvent)(unsafe.Pointer(ev))
	return &PenTouchEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(pt.windowID),
		Which:     uint32(pt.which),
		PenState:  uint32(pt.pen_state),
		X:         float32(pt.x),
		Y:         float32(pt.y),
		Eraser:    bool(pt.eraser),
		Down:      bool(pt.down),
	}
}

func wrapPenMotion(ev *C.SDL_Event) Event {
	pm := (*C.SDL_PenMotionEvent)(unsafe.Pointer(ev))
	return &PenMotionEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(pm.windowID),
		Which:     uint32(pm.which),
		PenState:  uint32(pm.pen_state),
		X:         float32(pm.x),
		Y:         float32(pm.y),
	}
}

func wrapPenButton(ev *C.SDL_Event) Event {
	pb := (*C.SDL_PenButtonEvent)(unsafe.Pointer(ev))
	return &PenButtonEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(pb.windowID),
		Which:     uint32(pb.which),
		PenState:  uint32(pb.pen_state),
		X:         float32(pb.x),
		Y:         float32(pb.y),
		Button:    uint8(pb.button),
		Down:      bool(pb.down),
	}
}

func wrapPenAxis(ev *C.SDL_Event) Event {
	pa := (*C.SDL_PenAxisEvent)(unsafe.Pointer(ev))
	return &PenAxisEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(pa.windowID),
		Which:     uint32(pa.which),
		PenState:  uint32(pa.pen_state),
		X:         float32(pa.x),
		Y:         float32(pa.y),
		Axis:      uint32(pa.axis),
		Value:     float32(pa.value),
	}
}

func wrapRender(ev *C.SDL_Event) Event {
	r := (*C.SDL_RenderEvent)(unsafe.Pointer(ev))
	return &RenderEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(r.windowID),
	}
}

func wrapDrop(ev *C.SDL_Event) Event {
	d := (*C.SDL_DropEvent)(unsafe.Pointer(ev))
	return &DropEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(d.windowID),
		X:         float32(d.x),
		Y:         float32(d.y),
		Source:    unsafe.Pointer(d.source),
		Data:      unsafe.Pointer(d.data),
	}
}

func wrapClipboard(ev *C.SDL_Event) Event {
	c := (*C.SDL_ClipboardEvent)(unsafe.Pointer(ev))
	return &ClipboardEvent{
		BaseEvent:    wrapCommon(ev),
		Owner:        bool(c.owner),
		NumMimeTypes: int32(c.num_mime_types),
		MimeTypes:    unsafe.Pointer(c.mime_types),
	}
}

func wrapUser(ev *C.SDL_Event) Event {
	u := (*C.SDL_UserEvent)(unsafe.Pointer(ev))
	return &UserEvent{
		BaseEvent: wrapCommon(ev),
		WindowID:  uint32(u.windowID),
		Code:      int32(u.code),
		Data1:     u.data1,
		Data2:     u.data2,
	}
}

func wrapEvent(ev *C.SDL_Event) Event {
	t := EventType(((*C.SDL_CommonEvent)(unsafe.Pointer(ev)))._type)
	if dec, ok := eventWrappers[t]; ok {
		return dec(ev)
	}

	if t >= EventTypeUser && t <= EventTypeLast {
		return wrapUser(ev)
	}

	return &UnknownEvent{BaseEvent: wrapCommon(ev)}
}

//---

// UnknownEvent is emitted for event types without a specific wrapper.
type UnknownEvent struct {
	BaseEvent
}

func (e *UnknownEvent) Base() BaseEvent {
	return e.BaseEvent
}

// QuitEvent is the quit requested event.
type QuitEvent struct {
	BaseEvent
}

func (e *QuitEvent) Base() BaseEvent {
	return e.BaseEvent
}

// WindowEvent is window state change event data.
type WindowEvent struct {
	BaseEvent
	WindowID uint32
	Data1    int32
	Data2    int32
}

func (e *WindowEvent) Base() BaseEvent {
	return e.BaseEvent
}

// GamepadDeviceEvent is gamepad device event data.
type GamepadDeviceEvent struct {
	BaseEvent
	Which int32
}

func (e *GamepadDeviceEvent) Base() BaseEvent {
	return e.BaseEvent
}

// GamepadTouchpadEvent is gamepad touchpad event data.
type GamepadTouchpadEvent struct {
	BaseEvent
	Which    int32
	Touchpad int32
	Finger   int32
	X        float32
	Y        float32
	Pressure float32
}

func (e *GamepadTouchpadEvent) Base() BaseEvent {
	return e.BaseEvent
}

// BasicEvent wraps events that only provide common fields.
type BasicEvent struct {
	BaseEvent
}

func (e *BasicEvent) Base() BaseEvent {
	return e.BaseEvent
}

// DisplayEvent is display state change event data.
type DisplayEvent struct {
	BaseEvent
	DisplayID uint32
	Data1     int32
	Data2     int32
}

func (e *DisplayEvent) Base() BaseEvent {
	return e.BaseEvent
}

// KeyboardDeviceEvent is keyboard device event structure.
type KeyboardDeviceEvent struct {
	BaseEvent
	Which uint32
}

func (e *KeyboardDeviceEvent) Base() BaseEvent {
	return e.BaseEvent
}

// KeyboardEvent is keyboard button event structure.
type KeyboardEvent struct {
	BaseEvent
	WindowID uint32
	Which    uint32
	Scancode uint32
	Key      KeyCode
	Mod      uint16
	Raw      uint16
	Down     bool
	Repeat   bool
}

func (e *KeyboardEvent) Base() BaseEvent {
	return e.BaseEvent
}

// TextEditingEvent is keyboard text editing event structure.
type TextEditingEvent struct {
	BaseEvent
	WindowID uint32
	Text     unsafe.Pointer
	Start    int32
	Length   int32
}

func (e *TextEditingEvent) Base() BaseEvent {
	return e.BaseEvent
}

// TextEditingCandidatesEvent is keyboard IME candidates event structure.
type TextEditingCandidatesEvent struct {
	BaseEvent
	WindowID          uint32
	Candidates        unsafe.Pointer
	NumCandidates     int32
	SelectedCandidate int32
	Horizontal        bool
}

func (e *TextEditingCandidatesEvent) Base() BaseEvent {
	return e.BaseEvent
}

// TextInputEvent is keyboard text input event structure.
type TextInputEvent struct {
	BaseEvent
	WindowID uint32
	Text     unsafe.Pointer
}

func (e *TextInputEvent) Base() BaseEvent {
	return e.BaseEvent
}

// MouseDeviceEvent is mouse device event structure.
type MouseDeviceEvent struct {
	BaseEvent
	Which uint32
}

func (e *MouseDeviceEvent) Base() BaseEvent {
	return e.BaseEvent
}

// MouseMotionEvent is mouse motion event structure.
type MouseMotionEvent struct {
	BaseEvent
	WindowID uint32
	Which    uint32
	State    uint32
	X        float32
	Y        float32
	XRel     float32
	YRel     float32
}

func (e *MouseMotionEvent) Base() BaseEvent {
	return e.BaseEvent
}

// MouseButtonEvent is mouse button event structure.
type MouseButtonEvent struct {
	BaseEvent
	WindowID uint32
	Which    uint32
	Button   uint8
	Down     bool
	Clicks   uint8
	X        float32
	Y        float32
}

func (e *MouseButtonEvent) Base() BaseEvent {
	return e.BaseEvent
}

// MouseWheelEvent is mouse wheel event structure.
type MouseWheelEvent struct {
	BaseEvent
	WindowID  uint32
	Which     uint32
	X         float32
	Y         float32
	Direction uint32
	MouseX    float32
	MouseY    float32
	IntegerX  int32
	IntegerY  int32
}

func (e *MouseWheelEvent) Base() BaseEvent {
	return e.BaseEvent
}

// JoyAxisEvent is joystick axis motion event structure.
type JoyAxisEvent struct {
	BaseEvent
	Which uint32
	Axis  uint8
	Value int16
}

func (e *JoyAxisEvent) Base() BaseEvent {
	return e.BaseEvent
}

// JoyBallEvent is joystick trackball motion event structure.
type JoyBallEvent struct {
	BaseEvent
	Which uint32
	Ball  uint8
	XRel  int16
	YRel  int16
}

func (e *JoyBallEvent) Base() BaseEvent {
	return e.BaseEvent
}

// JoyHatEvent is joystick hat position change event structure.
type JoyHatEvent struct {
	BaseEvent
	Which uint32
	Hat   uint8
	Value uint8
}

func (e *JoyHatEvent) Base() BaseEvent {
	return e.BaseEvent
}

// JoyButtonEvent is joystick button event structure.
type JoyButtonEvent struct {
	BaseEvent
	Which  uint32
	Button uint8
	Down   bool
}

func (e *JoyButtonEvent) Base() BaseEvent {
	return e.BaseEvent
}

// JoyDeviceEvent is joystick device event structure.
type JoyDeviceEvent struct {
	BaseEvent
	Which uint32
}

func (e *JoyDeviceEvent) Base() BaseEvent {
	return e.BaseEvent
}

// JoyBatteryEvent is joystick battery level change event structure.
type JoyBatteryEvent struct {
	BaseEvent
	Which   uint32
	State   int32
	Percent int32
}

func (e *JoyBatteryEvent) Base() BaseEvent {
	return e.BaseEvent
}

// GamepadAxisEvent is gamepad axis motion event structure.
type GamepadAxisEvent struct {
	BaseEvent
	Which uint32
	Axis  uint8
	Value int16
}

func (e *GamepadAxisEvent) Base() BaseEvent {
	return e.BaseEvent
}

// GamepadButtonEvent is gamepad button event structure.
type GamepadButtonEvent struct {
	BaseEvent
	Which  uint32
	Button uint8
	Down   bool
}

func (e *GamepadButtonEvent) Base() BaseEvent {
	return e.BaseEvent
}

// GamepadSensorEvent is gamepad sensor event structure.
type GamepadSensorEvent struct {
	BaseEvent
	Which           uint32
	Sensor          int32
	Data0           float32
	Data1           float32
	Data2           float32
	SensorTimestamp uint64
}

func (e *GamepadSensorEvent) Base() BaseEvent {
	return e.BaseEvent
}

// AudioDeviceEvent is audio device event structure.
type AudioDeviceEvent struct {
	BaseEvent
	Which     uint32
	Recording bool
}

func (e *AudioDeviceEvent) Base() BaseEvent {
	return e.BaseEvent
}

// CameraDeviceEvent is camera device event structure.
type CameraDeviceEvent struct {
	BaseEvent
	Which uint32
}

func (e *CameraDeviceEvent) Base() BaseEvent {
	return e.BaseEvent
}

// SensorEvent is sensor event structure.
type SensorEvent struct {
	BaseEvent
	Which           uint32
	Data0           float32
	Data1           float32
	Data2           float32
	Data3           float32
	Data4           float32
	Data5           float32
	SensorTimestamp uint64
}

func (e *SensorEvent) Base() BaseEvent {
	return e.BaseEvent
}

// TouchFingerEvent is touch finger event structure.
type TouchFingerEvent struct {
	BaseEvent
	TouchID  uint64
	FingerID uint64
	X        float32
	Y        float32
	DX       float32
	DY       float32
	Pressure float32
	WindowID uint32
}

func (e *TouchFingerEvent) Base() BaseEvent {
	return e.BaseEvent
}

// PinchFingerEvent is pinch event structure.
type PinchFingerEvent struct {
	BaseEvent
	Scale    float32
	WindowID uint32
}

func (e *PinchFingerEvent) Base() BaseEvent {
	return e.BaseEvent
}

// PenProximityEvent is pressure-sensitive pen proximity event structure.
type PenProximityEvent struct {
	BaseEvent
	WindowID uint32
	Which    uint32
}

func (e *PenProximityEvent) Base() BaseEvent {
	return e.BaseEvent
}

// PenTouchEvent is pressure-sensitive pen touched event structure.
type PenTouchEvent struct {
	BaseEvent
	WindowID uint32
	Which    uint32
	PenState uint32
	X        float32
	Y        float32
	Eraser   bool
	Down     bool
}

func (e *PenTouchEvent) Base() BaseEvent {
	return e.BaseEvent
}

// PenMotionEvent is pressure-sensitive pen motion event structure.
type PenMotionEvent struct {
	BaseEvent
	WindowID uint32
	Which    uint32
	PenState uint32
	X        float32
	Y        float32
}

func (e *PenMotionEvent) Base() BaseEvent {
	return e.BaseEvent
}

// PenButtonEvent is pressure-sensitive pen button event structure.
type PenButtonEvent struct {
	BaseEvent
	WindowID uint32
	Which    uint32
	PenState uint32
	X        float32
	Y        float32
	Button   uint8
	Down     bool
}

func (e *PenButtonEvent) Base() BaseEvent {
	return e.BaseEvent
}

// PenAxisEvent is pressure-sensitive pen pressure/angle event structure.
type PenAxisEvent struct {
	BaseEvent
	WindowID uint32
	Which    uint32
	PenState uint32
	X        float32
	Y        float32
	Axis     uint32
	Value    float32
}

func (e *PenAxisEvent) Base() BaseEvent {
	return e.BaseEvent
}

// RenderEvent is renderer event structure.
type RenderEvent struct {
	BaseEvent
	WindowID uint32
}

func (e *RenderEvent) Base() BaseEvent {
	return e.BaseEvent
}

// DropEvent is an event used to drop text or request a file open by the system.
type DropEvent struct {
	BaseEvent
	WindowID uint32
	X        float32
	Y        float32
	Source   unsafe.Pointer
	Data     unsafe.Pointer
}

func (e *DropEvent) Base() BaseEvent {
	return e.BaseEvent
}

// ClipboardEvent is triggered when the clipboard contents have changed.
type ClipboardEvent struct {
	BaseEvent
	Owner        bool
	NumMimeTypes int32
	MimeTypes    unsafe.Pointer
}

func (e *ClipboardEvent) Base() BaseEvent {
	return e.BaseEvent
}

// UserEvent is a user-defined event type.
type UserEvent struct {
	BaseEvent
	WindowID uint32
	Code     int32
	Data1    unsafe.Pointer
	Data2    unsafe.Pointer
}

func (e *UserEvent) Base() BaseEvent {
	return e.BaseEvent
}
