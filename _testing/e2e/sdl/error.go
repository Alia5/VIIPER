package sdl

/*
#cgo CFLAGS: -I${SRCDIR}/../deps/SDL/include

#include <stdlib.h>

#include <SDL3/SDL.h>
*/
import "C"

// SDLError wraps the SDL error message for the current thread.
type SDLError struct {
	eStr string
}

// Error returns the message with information about the specific error that occurred.
func (e *SDLError) Error() string {
	return e.eStr
}

// GetError retrieves a message about the last error that occurred on the current thread.
//
// It is possible for multiple errors to occur before calling SDL_GetError().
// Only the last error is returned.
func GetError() *SDLError {
	return &SDLError{
		eStr: C.GoString(C.SDL_GetError()),
	}
}
