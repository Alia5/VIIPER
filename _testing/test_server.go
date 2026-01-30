package testing

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Alia5/VIIPER/internal/cmd"
	"github.com/Alia5/VIIPER/internal/config"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/usb"
)

type MockServer struct {
	ApiServer *api.Server
	UsbServer *usb.Server
}

func NewTestServerWithConfig(t *testing.T, cfg *config.CLI) *MockServer {
	t.Helper()

	logger := slog.Default()

	usbServer := usb.New(cfg.Server.UsbServerConfig, logger, nil)

	usbErrCh := make(chan error, 1)
	go func() {
		usbErrCh <- usbServer.ListenAndServe()
	}()
	select {
	case <-usbServer.Ready():
		// ok
	case err := <-usbErrCh:
		if err == nil {
			err = io.ErrUnexpectedEOF
		}
		t.Fatalf("USB server failed to start: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatalf("USB server did not become ready")
	}

	return &MockServer{
		UsbServer: usbServer,
		ApiServer: api.New(
			usbServer,
			cfg.Server.ApiServerConfig.Addr,
			cfg.Server.ApiServerConfig,
			logger,
		),
	}
}

func NewTestServer(t *testing.T) *MockServer {
	t.Helper()

	cfg := TestServerConfig(t)
	return NewTestServerWithConfig(t, cfg)
}

func TestServerConfig(t *testing.T) *config.CLI {
	t.Helper()

	return &config.CLI{
		Server: cmd.Server{
			UsbServerConfig: usb.ServerConfig{
				Addr:              "localhost:0",
				ConnectionTimeout: 1 * time.Second,
				BusCleanupTimeout: 1 * time.Second,
			},
			ApiServerConfig: api.ServerConfig{
				Addr:                        "localhost:0",
				DeviceHandlerConnectTimeout: 1 * time.Second,
				ConnectionTimeout:           1 * time.Second,
				AutoAttachLocalClient:       false,
			},
		},
	}
}
