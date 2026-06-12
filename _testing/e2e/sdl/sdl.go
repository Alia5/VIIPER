package sdl

/*
#cgo CFLAGS: -I${SRCDIR}/../deps/SDL/include
#cgo LDFLAGS: -L${SRCDIR}/../deps/SDL/build/Debug -lSDL3

#include <stdlib.h>

#include <SDL3/SDL.h>
#include <SDL3/SDL_init.h>
*/
import "C"
import "runtime"

func init() {
	runtime.LockOSThread()
}

// InitFlags for SDL_Init and/or SDL_InitSubSystem.
//
// These are the flags which may be passed to SDL_Init(). You should specify
// the subsystems which you will be using in your application.
type InitFlags uint32

// These flags may be passed to SDL_Init().
const (
	InitFlagAudio    InitFlags = C.SDL_INIT_AUDIO
	InitFlagVideo    InitFlags = C.SDL_INIT_VIDEO
	InitFlagJoystick InitFlags = C.SDL_INIT_JOYSTICK
	InitFlagHaptic   InitFlags = C.SDL_INIT_HAPTIC
	InitFlagGamepad  InitFlags = C.SDL_INIT_GAMEPAD
	InitFlagEvents   InitFlags = C.SDL_INIT_EVENTS
	InitFlagSensor   InitFlags = C.SDL_INIT_SENSOR
	InitFlagCamera   InitFlags = C.SDL_INIT_CAMERA
)

// Init initializes the SDL library.
//
// Init simply forwards to calling SDL_InitSubSystem(). Therefore, the
// two may be used interchangeably.
//
// Subsystem initialization is ref-counted; call QuitSubSystem() for each
// InitSubSystem(), or call Quit() to force shutdown.
//
// This function should only be called on the main thread.
func Init(flags InitFlags) error {
	res := C.SDL_Init(C.Uint32(flags))
	if !res {
		return GetError()
	}
	return nil
}

// InitSubSystem compatibility function to initialize the SDL library.
//
// This function and SDL_Init() are interchangeable.
//
// This function should only be called on the main thread.
func InitSubSystem(flags InitFlags) error {
	res := C.SDL_InitSubSystem(C.Uint32(flags))
	if !res {
		return GetError()
	}
	return nil
}

// QuitSubSystem shuts down specific SDL subsystems.
//
// You still need to call SDL_Quit() even if you close all open subsystems
// with SDL_QuitSubSystem().
func QuitSubSystem(flags InitFlags) {
	C.SDL_QuitSubSystem(C.Uint32(flags))
}

// Quit cleans up all initialized subsystems.
//
// This function should only be called on the main thread.
func Quit() {
	C.SDL_Quit()
}
