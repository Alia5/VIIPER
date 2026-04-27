package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef uintptr_t USBServerHandle;
*/
import "C"
import (
	"runtime/cgo"

	"github.com/Alia5/VIIPER/virtualbus"
)

// CreateUSBBus creates a new USB bus on the server associated with the given handle.
// @param handle Handle to the USB server.
// @param busID ID of the bus to create. If 0 or NULL, the server will assign the next free bus ID.
//
//export CreateUSBBus
func CreateUSBBus(handle C.USBServerHandle, busID *uint32) bool {
	h := cgo.Handle(handle)
	hw, ok := h.Value().(*usbServerHandleWrapper)
	if !ok {
		return false
	}

	if busID == nil {
		id := hw.s.NextFreeBusID()
		busID = &id
	} else if *busID == 0 {
		*busID = hw.s.NextFreeBusID()
	}

	b, err := virtualbus.NewWithBusId(*busID)
	if err != nil {
		return false
	}
	if err := hw.s.AddBus(b); err != nil {
		return false
	}
	hw.mtx.Lock()
	defer hw.mtx.Unlock()
	hw.deviceHandles[*busID] = make([]deviceHandle, 0)

	return true
}

// RemoveUSBBus removes the USB bus with the given ID from the server associated with the given handle.
// Automatically removes devices associated with the bus.
// @param handle Handle to the USB server.
// @param busID ID of the bus to remove.
//
//export RemoveUSBBus
func RemoveUSBBus(handle C.USBServerHandle, busID uint32) bool {
	h := cgo.Handle(handle)
	hw, ok := h.Value().(*usbServerHandleWrapper)
	if !ok {
		return false
	}

	if err := hw.s.RemoveBus(busID); err != nil {
		return false
	}
	hw.mtx.Lock()
	defer hw.mtx.Unlock()
	for _, dh := range hw.deviceHandles[busID] {
		cgo.Handle(dh).Delete()
	}
	delete(hw.deviceHandles, busID)

	return true
}
