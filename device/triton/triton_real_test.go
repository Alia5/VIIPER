package triton_test

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/device/triton"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/usb"
	"github.com/Alia5/VIIPER/virtualbus"
)

func TestTritonReal(t *testing.T) {

	logger := slog.Default()

	usbSrv := usb.New(usb.ServerConfig{
		Addr:                    "localhost:3245",
		ConnectionTimeout:       5 * time.Minute,
		BusCleanupTimeout:       1 * time.Minute,
		WriteBatchFlushInterval: 1 * time.Millisecond,
	}, logger, nil)

	_, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := usbSrv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	busId := usbSrv.NextFreeBusID()
	bus := virtualbus.New(busId)

	if err := usbSrv.AddBus(bus); err != nil {
		t.Fatalf("Failed to add bus: %v", err)
	}

	tritonDevice := triton.New(&device.CreateOptions{})
	devCtx, err := bus.Add(tritonDevice)
	if err != nil {
		t.Fatalf("Failed to add device to bus: %v", err)
	}

	exportMeta := device.GetDeviceMeta(devCtx)
	if exportMeta == nil {
		t.Fatalf("Failed to get device metadata from context")
	}

	err = api.AttachLocalhostClient(
		context.Background(),
		exportMeta,
		usbSrv.GetListenPort(),
		true,
		logger,
	)
	if err != nil {
		t.Fatalf("Failed to attach localhost client: %v", err)
	}

	for {
		select {
		case <-devCtx.Done():
			return
		}
	}

}
