package sdl

/*
#cgo CFLAGS: -I${SRCDIR}/../deps/SDL/include

#include <stdlib.h>

#include <SDL3/SDL_guid.h>
#include <SDL3/SDL_joystick.h>
*/
import "C"

import (
	"unsafe"
)

// GUID is a 128-bit identifier for an input device that identifies that device across runs of SDL programs on the same platform.
type GUID [16]byte

// String converts a GUID to an ASCII string representation.
func (g GUID) String() string {
	var buf [33]C.char
	cg := *(*C.SDL_GUID)(unsafe.Pointer(&g))
	C.SDL_GUIDToString(cg, &buf[0], C.int(len(buf)))
	return C.GoString(&buf[0])
}

// StringToGUID converts a GUID string into a GUID structure.
func StringToGUID(s string) GUID {
	cStr := C.CString(s)
	defer C.free(unsafe.Pointer(cStr))
	cg := C.SDL_StringToGUID(cStr)
	return *(*GUID)(unsafe.Pointer(&cg))
}

// JoystickID is a unique ID for a joystick for the time it is connected to the system, and is never reused for the lifetime of the application.
type JoystickID uint32

// JoystickType is an enum of some common joystick types.
type JoystickType uint32

// JoystickConnectionState is the possible connection states for a joystick device.
type JoystickConnectionState int32

// JoystickType values report common low-level joystick types.
const (
	JoystickTypeUnknown     JoystickType = C.SDL_JOYSTICK_TYPE_UNKNOWN
	JoystickTypeGamepad     JoystickType = C.SDL_JOYSTICK_TYPE_GAMEPAD
	JoystickTypeWheel       JoystickType = C.SDL_JOYSTICK_TYPE_WHEEL
	JoystickTypeArcadeStick JoystickType = C.SDL_JOYSTICK_TYPE_ARCADE_STICK
	JoystickTypeFlightStick JoystickType = C.SDL_JOYSTICK_TYPE_FLIGHT_STICK
	JoystickTypeDancePad    JoystickType = C.SDL_JOYSTICK_TYPE_DANCE_PAD
	JoystickTypeGuitar      JoystickType = C.SDL_JOYSTICK_TYPE_GUITAR
	JoystickTypeDrumKit     JoystickType = C.SDL_JOYSTICK_TYPE_DRUM_KIT
	JoystickTypeArcadePad   JoystickType = C.SDL_JOYSTICK_TYPE_ARCADE_PAD
	JoystickTypeThrottle    JoystickType = C.SDL_JOYSTICK_TYPE_THROTTLE
)

// JoystickConnectionState values report how a joystick is connected to the system.
const (
	JoystickConnectionInvalid  JoystickConnectionState = C.SDL_JOYSTICK_CONNECTION_INVALID
	JoystickConnectionUnknown  JoystickConnectionState = C.SDL_JOYSTICK_CONNECTION_UNKNOWN
	JoystickConnectionWired    JoystickConnectionState = C.SDL_JOYSTICK_CONNECTION_WIRED
	JoystickConnectionWireless JoystickConnectionState = C.SDL_JOYSTICK_CONNECTION_WIRELESS
)

// JoystickAxisMax and JoystickAxisMin are the largest and smallest values a joystick axis can report.
const (
	JoystickAxisMax = C.SDL_JOYSTICK_AXIS_MAX
	JoystickAxisMin = C.SDL_JOYSTICK_AXIS_MIN
)

// Joystick is the joystick structure used to identify an SDL joystick.
//
// This is opaque data.
type Joystick struct {
	cJoystick *C.SDL_Joystick
}

// InitJoystickSubSystem initializes the joystick subsystem.
//
// The joystick subsystem must be initialized before a joystick can be opened for use.
func InitJoystickSubSystem() error {
	return InitSubSystem(InitFlagJoystick)
}

// QuitJoystickSubSystem shuts down the joystick subsystem.
func QuitJoystickSubSystem() {
	QuitSubSystem(InitFlagJoystick)
}

// LockJoysticks locks the joysticks while processing.
func LockJoysticks() {
	C.SDL_LockJoysticks()
}

// TryLockJoysticks attempts to lock the joysticks while processing.
func TryLockJoysticks() bool {
	return bool(C.SDL_TryLockJoysticks())
}

// UnlockJoysticks unlocks the joysticks.
func UnlockJoysticks() {
	C.SDL_UnlockJoysticks()
}

// HasJoystick returns whether a joystick is currently connected.
func HasJoystick() bool {
	return bool(C.SDL_HasJoystick())
}

// GetJoysticks returns a list of currently connected joysticks.
func GetJoysticks() ([]JoystickID, error) {
	var count C.int
	cIDs := C.SDL_GetJoysticks(&count)
	if cIDs == nil {
		if count == 0 {
			return []JoystickID{}, nil
		}
		return nil, GetError()
	}
	defer C.SDL_free(unsafe.Pointer(cIDs))

	cSlice := unsafe.Slice(cIDs, int(count))
	ids := make([]JoystickID, 0, int(count))
	for _, id := range cSlice {
		ids = append(ids, JoystickID(id))
	}
	return ids, nil
}

// GetJoystickNameForID gets the implementation dependent name of a joystick.
//
// This can be called before any joysticks are opened.
func GetJoystickNameForID(instanceID JoystickID) string {
	cName := C.SDL_GetJoystickNameForID(C.SDL_JoystickID(instanceID))
	if cName == nil {
		return ""
	}
	return C.GoString(cName)
}

// GetJoystickPathForID gets the implementation dependent path of a joystick.
//
// This can be called before any joysticks are opened.
func GetJoystickPathForID(instanceID JoystickID) string {
	cPath := C.SDL_GetJoystickPathForID(C.SDL_JoystickID(instanceID))
	if cPath == nil {
		return ""
	}
	return C.GoString(cPath)
}

// GetJoystickPlayerIndexForID gets the player index of a joystick.
//
// This can be called before any joysticks are opened.
func GetJoystickPlayerIndexForID(instanceID JoystickID) int {
	return int(C.SDL_GetJoystickPlayerIndexForID(C.SDL_JoystickID(instanceID)))
}

// GetJoystickGUIDForID gets the implementation-dependent GUID of a joystick.
//
// This can be called before any joysticks are opened.
func GetJoystickGUIDForID(instanceID JoystickID) GUID {
	cg := C.SDL_GetJoystickGUIDForID(C.SDL_JoystickID(instanceID))
	return *(*GUID)(unsafe.Pointer(&cg))
}

// GetJoystickVendorForID gets the USB vendor ID of a joystick, if available.
//
// This can be called before any joysticks are opened.
func GetJoystickVendorForID(instanceID JoystickID) uint16 {
	return uint16(C.SDL_GetJoystickVendorForID(C.SDL_JoystickID(instanceID)))
}

// GetJoystickProductForID gets the USB product ID of a joystick, if available.
//
// This can be called before any joysticks are opened.
func GetJoystickProductForID(instanceID JoystickID) uint16 {
	return uint16(C.SDL_GetJoystickProductForID(C.SDL_JoystickID(instanceID)))
}

// GetJoystickProductVersionForID gets the product version of a joystick, if available.
//
// This can be called before any joysticks are opened.
func GetJoystickProductVersionForID(instanceID JoystickID) uint16 {
	return uint16(C.SDL_GetJoystickProductVersionForID(C.SDL_JoystickID(instanceID)))
}

// GetJoystickTypeForID gets the type of a joystick, if available.
//
// This can be called before any joysticks are opened.
func GetJoystickTypeForID(instanceID JoystickID) JoystickType {
	return JoystickType(C.SDL_GetJoystickTypeForID(C.SDL_JoystickID(instanceID)))
}

// OpenJoystick opens a joystick for use.
//
// The joystick subsystem must be initialized before a joystick can be opened for use.
func OpenJoystick(instanceID JoystickID) (*Joystick, error) {
	cj := C.SDL_OpenJoystick(C.SDL_JoystickID(instanceID))
	if cj == nil {
		return nil, GetError()
	}
	return &Joystick{cJoystick: cj}, nil
}

// GetJoystickFromID gets the SDL_Joystick associated with an instance ID, if it has been opened.
func GetJoystickFromID(instanceID JoystickID) (*Joystick, bool) {
	cj := C.SDL_GetJoystickFromID(C.SDL_JoystickID(instanceID))
	if cj == nil {
		return nil, false
	}
	return &Joystick{cJoystick: cj}, true
}

// GetJoystickFromPlayerIndex gets the SDL_Joystick associated with a player index.
func GetJoystickFromPlayerIndex(playerIndex int) (*Joystick, bool) {
	cj := C.SDL_GetJoystickFromPlayerIndex(C.int(playerIndex))
	if cj == nil {
		return nil, false
	}
	return &Joystick{cJoystick: cj}, true
}

// AttachVirtualJoystick attaches a new virtual joystick.
func AttachVirtualJoystick(desc unsafe.Pointer) JoystickID {
	return JoystickID(C.SDL_AttachVirtualJoystick((*C.SDL_VirtualJoystickDesc)(desc)))
}

// DetachVirtualJoystick detaches a virtual joystick.
func DetachVirtualJoystick(instanceID JoystickID) bool {
	return bool(C.SDL_DetachVirtualJoystick(C.SDL_JoystickID(instanceID)))
}

// IsJoystickVirtual queries whether or not a joystick is virtual.
func IsJoystickVirtual(instanceID JoystickID) bool {
	return bool(C.SDL_IsJoystickVirtual(C.SDL_JoystickID(instanceID)))
}

// SetJoystickEventsEnabled sets the state of joystick event processing.
func SetJoystickEventsEnabled(enabled bool) {
	C.SDL_SetJoystickEventsEnabled(C.bool(enabled))
}

// JoystickEventsEnabled queries the state of joystick event processing.
func JoystickEventsEnabled() bool {
	return bool(C.SDL_JoystickEventsEnabled())
}

// UpdateJoysticks updates the current state of the open joysticks.
func UpdateJoysticks() {
	C.SDL_UpdateJoysticks()
}

// Close closes a joystick previously opened with SDL_OpenJoystick().
func (j *Joystick) Close() {
	if j == nil || j.cJoystick == nil {
		return
	}
	C.SDL_CloseJoystick(j.cJoystick)
	j.cJoystick = nil
}

// Connected gets the status of a specified joystick.
func (j *Joystick) Connected() bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	return bool(C.SDL_JoystickConnected(j.cJoystick))
}

// ID gets the instance ID of an opened joystick.
func (j *Joystick) ID() JoystickID {
	if j == nil || j.cJoystick == nil {
		return 0
	}
	return JoystickID(C.SDL_GetJoystickID(j.cJoystick))
}

// Name gets the implementation dependent name of a joystick.
func (j *Joystick) Name() string {
	if j == nil || j.cJoystick == nil {
		return ""
	}
	cName := C.SDL_GetJoystickName(j.cJoystick)
	if cName == nil {
		return ""
	}
	return C.GoString(cName)
}

// Path gets the implementation dependent path of a joystick.
func (j *Joystick) Path() string {
	if j == nil || j.cJoystick == nil {
		return ""
	}
	cPath := C.SDL_GetJoystickPath(j.cJoystick)
	if cPath == nil {
		return ""
	}
	return C.GoString(cPath)
}

// PlayerIndex gets the player index of an opened joystick.
func (j *Joystick) PlayerIndex() int {
	if j == nil || j.cJoystick == nil {
		return -1
	}
	return int(C.SDL_GetJoystickPlayerIndex(j.cJoystick))
}

// SetPlayerIndex sets the player index of an opened joystick.
func (j *Joystick) SetPlayerIndex(playerIndex int) error {
	if j == nil || j.cJoystick == nil {
		return &SDLError{eStr: "invalid joystick handle"}
	}
	if !C.SDL_SetJoystickPlayerIndex(j.cJoystick, C.int(playerIndex)) {
		return GetError()
	}
	return nil
}

// GUID gets the implementation-dependent GUID for the joystick.
func (j *Joystick) GUID() GUID {
	if j == nil || j.cJoystick == nil {
		return GUID{}
	}
	cg := C.SDL_GetJoystickGUID(j.cJoystick)
	return *(*GUID)(unsafe.Pointer(&cg))
}

// Vendor gets the USB vendor ID of an opened joystick, if available.
func (j *Joystick) Vendor() uint16 {
	if j == nil || j.cJoystick == nil {
		return 0
	}
	return uint16(C.SDL_GetJoystickVendor(j.cJoystick))
}

// Product gets the USB product ID of an opened joystick, if available.
func (j *Joystick) Product() uint16 {
	if j == nil || j.cJoystick == nil {
		return 0
	}
	return uint16(C.SDL_GetJoystickProduct(j.cJoystick))
}

// ProductVersion gets the product version of an opened joystick, if available.
func (j *Joystick) ProductVersion() uint16 {
	if j == nil || j.cJoystick == nil {
		return 0
	}
	return uint16(C.SDL_GetJoystickProductVersion(j.cJoystick))
}

// FirmwareVersion gets the firmware version of an opened joystick, if available.
func (j *Joystick) FirmwareVersion() uint16 {
	if j == nil || j.cJoystick == nil {
		return 0
	}
	return uint16(C.SDL_GetJoystickFirmwareVersion(j.cJoystick))
}

// Serial gets the serial number of an opened joystick, if available.
func (j *Joystick) Serial() string {
	if j == nil || j.cJoystick == nil {
		return ""
	}
	cSerial := C.SDL_GetJoystickSerial(j.cJoystick)
	if cSerial == nil {
		return ""
	}
	return C.GoString(cSerial)
}

// Type gets the type of an opened joystick.
func (j *Joystick) Type() JoystickType {
	if j == nil || j.cJoystick == nil {
		return JoystickTypeUnknown
	}
	return JoystickType(C.SDL_GetJoystickType(j.cJoystick))
}

// NumAxes gets the number of general axis controls on a joystick.
func (j *Joystick) NumAxes() int {
	if j == nil || j.cJoystick == nil {
		return -1
	}
	return int(C.SDL_GetNumJoystickAxes(j.cJoystick))
}

// NumBalls gets the number of trackballs on a joystick.
func (j *Joystick) NumBalls() int {
	if j == nil || j.cJoystick == nil {
		return -1
	}
	return int(C.SDL_GetNumJoystickBalls(j.cJoystick))
}

// NumHats gets the number of POV hats on a joystick.
func (j *Joystick) NumHats() int {
	if j == nil || j.cJoystick == nil {
		return -1
	}
	return int(C.SDL_GetNumJoystickHats(j.cJoystick))
}

// NumButtons gets the number of buttons on a joystick.
func (j *Joystick) NumButtons() int {
	if j == nil || j.cJoystick == nil {
		return -1
	}
	return int(C.SDL_GetNumJoystickButtons(j.cJoystick))
}

// GetAxis gets the current state of an axis control on a joystick.
func (j *Joystick) GetAxis(axis int) int16 {
	if j == nil || j.cJoystick == nil {
		return 0
	}
	return int16(C.SDL_GetJoystickAxis(j.cJoystick, C.int(axis)))
}

// GetAxisInitialState gets the initial state of an axis control on a joystick.
func (j *Joystick) GetAxisInitialState(axis int) (bool, int16) {
	if j == nil || j.cJoystick == nil {
		return false, 0
	}
	var state C.Sint16
	hasState := C.SDL_GetJoystickAxisInitialState(j.cJoystick, C.int(axis), &state)
	return bool(hasState), int16(state)
}

// GetBall gets the ball axis change since the last poll.
func (j *Joystick) GetBall(ball int) (bool, int, int) {
	if j == nil || j.cJoystick == nil {
		return false, 0, 0
	}
	var dx, dy C.int
	ok := C.SDL_GetJoystickBall(j.cJoystick, C.int(ball), &dx, &dy)
	return bool(ok), int(dx), int(dy)
}

// GetHat gets the current state of a POV hat on a joystick.
func (j *Joystick) GetHat(hat int) uint8 {
	if j == nil || j.cJoystick == nil {
		return 0
	}
	return uint8(C.SDL_GetJoystickHat(j.cJoystick, C.int(hat)))
}

// GetButton gets the current state of a button on a joystick.
func (j *Joystick) GetButton(button int) bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	return bool(C.SDL_GetJoystickButton(j.cJoystick, C.int(button)))
}

// Rumble starts a rumble effect.
func (j *Joystick) Rumble(lowFrequencyRumble, highFrequencyRumble uint16, durationMs uint32) bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	return bool(C.SDL_RumbleJoystick(j.cJoystick, C.Uint16(lowFrequencyRumble), C.Uint16(highFrequencyRumble), C.Uint32(durationMs)))
}

// RumbleTriggers starts a rumble effect in the joystick's triggers.
func (j *Joystick) RumbleTriggers(leftRumble, rightRumble uint16, durationMs uint32) bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	return bool(C.SDL_RumbleJoystickTriggers(j.cJoystick, C.Uint16(leftRumble), C.Uint16(rightRumble), C.Uint32(durationMs)))
}

// SetLED updates a joystick's LED color.
func (j *Joystick) SetLED(red, green, blue uint8) bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	return bool(C.SDL_SetJoystickLED(j.cJoystick, C.Uint8(red), C.Uint8(green), C.Uint8(blue)))
}

// SendEffect sends a joystick specific effect packet.
func (j *Joystick) SendEffect(data unsafe.Pointer, size int) bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	return bool(C.SDL_SendJoystickEffect(j.cJoystick, data, C.int(size)))
}

// ConnectionState gets the connection state of a joystick.
func (j *Joystick) ConnectionState() JoystickConnectionState {
	if j == nil || j.cJoystick == nil {
		return JoystickConnectionInvalid
	}
	return JoystickConnectionState(C.SDL_GetJoystickConnectionState(j.cJoystick))
}

// PowerInfo gets the battery state of a joystick.
func (j *Joystick) PowerInfo() (int, int) {
	if j == nil || j.cJoystick == nil {
		return -1, -1
	}
	var percent C.int
	state := C.SDL_GetJoystickPowerInfo(j.cJoystick, &percent)
	return int(state), int(percent)
}

// GetProperties gets the properties associated with a joystick.
func (j *Joystick) GetProperties() uintptr {
	if j == nil || j.cJoystick == nil {
		return 0
	}
	return uintptr(C.SDL_GetJoystickProperties(j.cJoystick))
}

// GetJoystickGUIDInfo gets the device information encoded in a GUID structure.
func GetJoystickGUIDInfo(guid GUID) (vendor, product, version, crc16 uint16) {
	var cVendor, cProduct, cVersion, cCRC16 C.Uint16
	cg := *(*C.SDL_GUID)(unsafe.Pointer(&guid))
	C.SDL_GetJoystickGUIDInfo(cg, &cVendor, &cProduct, &cVersion, &cCRC16)
	return uint16(cVendor), uint16(cProduct), uint16(cVersion), uint16(cCRC16)
}

// SetVirtualAxis sets the state of an axis on an opened virtual joystick.
func (j *Joystick) SetVirtualAxis(axis int, value int16) bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	return bool(C.SDL_SetJoystickVirtualAxis(j.cJoystick, C.int(axis), C.Sint16(value)))
}

// SetVirtualBall generates ball motion on an opened virtual joystick.
func (j *Joystick) SetVirtualBall(ball int, xRel, yRel int16) bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	return bool(C.SDL_SetJoystickVirtualBall(j.cJoystick, C.int(ball), C.Sint16(xRel), C.Sint16(yRel)))
}

// SetVirtualButton sets the state of a button on an opened virtual joystick.
func (j *Joystick) SetVirtualButton(button int, down bool) bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	return bool(C.SDL_SetJoystickVirtualButton(j.cJoystick, C.int(button), C.bool(down)))
}

// SetVirtualHat sets the state of a hat on an opened virtual joystick.
func (j *Joystick) SetVirtualHat(hat int, value uint8) bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	return bool(C.SDL_SetJoystickVirtualHat(j.cJoystick, C.int(hat), C.Uint8(value)))
}

// SetVirtualTouchpad sets touchpad finger state on an opened virtual joystick.
func (j *Joystick) SetVirtualTouchpad(touchpad, finger int, down bool, x, y, pressure float32) bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	return bool(C.SDL_SetJoystickVirtualTouchpad(
		j.cJoystick,
		C.int(touchpad),
		C.int(finger),
		C.bool(down),
		C.float(x),
		C.float(y),
		C.float(pressure),
	))
}

// SendVirtualSensorData sends a sensor update for an opened virtual joystick.
func (j *Joystick) SendVirtualSensorData(sensorType int32, sensorTimestamp uint64, values []float32) bool {
	if j == nil || j.cJoystick == nil {
		return false
	}
	if len(values) == 0 {
		return bool(C.SDL_SendJoystickVirtualSensorData(j.cJoystick, C.SDL_SensorType(sensorType), C.Uint64(sensorTimestamp), nil, 0))
	}
	return bool(C.SDL_SendJoystickVirtualSensorData(
		j.cJoystick,
		C.SDL_SensorType(sensorType),
		C.Uint64(sensorTimestamp),
		(*C.float)(unsafe.Pointer(&values[0])),
		C.int(len(values)),
	))
}
