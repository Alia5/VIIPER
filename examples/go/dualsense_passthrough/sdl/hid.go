package sdl

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../_testing/e2e/deps/SDL/include

#include <stdlib.h>

#include <SDL3/SDL.h>
#include <SDL3/SDL_hidapi.h>
*/
import "C"

import "unsafe"

type HIDDevice struct {
	cDevice *C.SDL_hid_device
}

func HIDOpenPath(path string) (*HIDDevice, error) {
	if C.SDL_hid_init() != 0 {
		return nil, GetError()
	}
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	cDevice := C.SDL_hid_open_path(cPath)
	if cDevice == nil {
		return nil, GetError()
	}
	return &HIDDevice{cDevice: cDevice}, nil
}

func (d *HIDDevice) ReadTimeout(buf []byte, timeoutMs int) int {
	return int(C.SDL_hid_read_timeout(d.cDevice, (*C.uchar)(unsafe.Pointer(&buf[0])), C.size_t(len(buf)), C.int(timeoutMs)))
}

func (d *HIDDevice) Close() {
	C.SDL_hid_close(d.cDevice)
}
