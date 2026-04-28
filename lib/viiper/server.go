package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef struct {
	char* addr; // default "0.0.0.0:3241"
	uint64_t connection_timeout_ms; // default 30000 (30s)
	uint64_t device_handler_connect_timeout_ms; // default 5000 (5s)
	uint32_t write_batch_flush_interval_ms; // default 1 (1ms)
} USBServerConfig;

typedef uintptr_t USBServerHandle;

typedef enum {
    VIIPER_LOG_DEBUG = -4,
    VIIPER_LOG_INFO  = 0,
    VIIPER_LOG_WARN  = 4,
    VIIPER_LOG_ERROR = 8,
} VIIPERLogLevel;

typedef void (*VIIPERLogCallback)(VIIPERLogLevel level, const char* message);

static void viiper_call_log(VIIPERLogCallback fn, VIIPERLogLevel level, const char* msg) {
	fn(level, msg);
}
*/
import "C"

import (
	"log/slog"
	"runtime/cgo"
	"time"
	"unsafe"

	"github.com/Alia5/VIIPER/internal/server/usb"
)

// NewUSBServer creates a new USB server with the given configuration and returns a handle to it.
// The server will run in the background and can be stopped by calling CloseUSBServer with the returned handle.
// @param config Server configuration
// @param outHandle Output parameter for the created server handle
// @param logCallback Optional callback function for log messages from the USB server
//
//export NewUSBServer
func NewUSBServer(config *C.USBServerConfig, outHandle *C.USBServerHandle, logCallback C.VIIPERLogCallback) bool {
	addr := C.GoString(config.addr)
	connectionTimeout := time.Duration(config.connection_timeout_ms) * time.Millisecond
	busCleanupTimeout := time.Duration(config.device_handler_connect_timeout_ms) * time.Millisecond
	writeBatchFlushInterval := time.Duration(config.write_batch_flush_interval_ms) * time.Millisecond

	if addr == "" {
		addr = ":3241"
	}
	if connectionTimeout == 0 {
		connectionTimeout = 30 * time.Second
	}
	if busCleanupTimeout == 0 {
		busCleanupTimeout = 5 * time.Second
	}
	if writeBatchFlushInterval == 0 {
		writeBatchFlushInterval = 1 * time.Millisecond
	}

	var logger *slog.Logger
	if logCallback != nil {
		logger = slog.New(&funcLogHandler{
			func(level slog.Level, msg string) {
				if logCallback == nil {
					return
				}
				cMsg := C.CString(msg)
				defer C.free(unsafe.Pointer(cMsg))
				C.viiper_call_log(logCallback, C.VIIPERLogLevel(level), cMsg)
			},
		})
	} else {
		logger = slog.New(slog.DiscardHandler)
	}
	slog.SetDefault(logger)

	s := usb.New(usb.ServerConfig{
		Addr:                    addr,
		ConnectionTimeout:       connectionTimeout,
		BusCleanupTimeout:       busCleanupTimeout,
		WriteBatchFlushInterval: writeBatchFlushInterval,
	}, logger, nil)

	readyChan := s.Ready()
	errChan := make(chan error, 1)

	go func() {
		errChan <- s.ListenAndServe()
	}()

	select {
	case <-readyChan:
		*outHandle = C.USBServerHandle(cgo.NewHandle(&usbServerHandleWrapper{
			s:             s,
			deviceHandles: make(map[uint32][]deviceHandle),
		}))
		return true
	case err := <-errChan:
		logger.Error("NewUSBServer: ListenAndServe failed", "error", err)
		return false
	}
}

// CloseUSBServer closes the USB server associated with the given handle.
// Automatically removes busses and devices associated with the server.
// @param handle Handle to the USB server to close.
//
//export CloseUSBServer
func CloseUSBServer(handle C.USBServerHandle) bool {
	h := cgo.Handle(handle)
	hw, ok := h.Value().(*usbServerHandleWrapper)
	if !ok {
		return false
	}
	hw.mtx.Lock()
	defer hw.mtx.Unlock()

	for busID, dhs := range hw.deviceHandles {
		for _, dh := range dhs {
			cgo.Handle(dh).Delete()
		}
		delete(hw.deviceHandles, busID)
	}
	hw.deviceHandles = nil

	if err := hw.s.Close(); err != nil {
		return false
	}
	h.Delete()
	return true
}
