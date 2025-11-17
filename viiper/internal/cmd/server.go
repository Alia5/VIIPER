package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"viiper/internal/log"
	"viiper/internal/server/api"
	"viiper/internal/server/api/handler"
	"viiper/internal/server/usb"
)

type Server struct {
	UsbServerConfig   usb.ServerConfig `embed:"" prefix:"usb."`
	ApiServerConfig   api.ServerConfig `embed:"" prefix:"api."`
	ConnectionTimeout time.Duration    `help:"ConnectionTimeout operation timeout" default:"30s" env:"VIIPER_CONNECTION_TIMEOUT"`
}

// Run is called by Kong when the server command is executed.
func (s *Server) Run(logger *slog.Logger, rawLogger log.RawLogger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s.UsbServerConfig.ConnectionTimeout = s.ConnectionTimeout
	s.ApiServerConfig.ConnectionTimeout = s.ConnectionTimeout

	logger.Info("Starting VIIPER USB-IP server", "addr", s.UsbServerConfig.Addr)
	usbSrv := usb.New(s.UsbServerConfig, logger, rawLogger)

	usbErrCh := make(chan error, 1)
	go func() {
		usbErrCh <- usbSrv.ListenAndServe()
	}()

	select {
	case err := <-usbErrCh:
		return err
	case <-usbSrv.Ready():
	}

	if s.ApiServerConfig.Addr == "" {
		logger.Error("API server address must be set (default :3242).")
		return fmt.Errorf("API server address must be set (default :3242).")
	}

	apiSrv := api.New(usbSrv, s.ApiServerConfig.Addr, s.ApiServerConfig, logger)
	r := apiSrv.Router()
	r.Register("bus/list", handler.BusList(usbSrv))
	r.Register("bus/create", handler.BusCreate(usbSrv))
	r.Register("bus/remove", handler.BusRemove(usbSrv))
	r.Register("bus/{id}/list", handler.BusDevicesList(usbSrv))
	r.Register("bus/{id}/add", handler.BusDeviceAdd(usbSrv, apiSrv))
	r.Register("bus/{id}/remove", handler.BusDeviceRemove(usbSrv))
	r.RegisterStream("bus/{busId}/{deviceid}", api.DeviceStreamHandler(usbSrv))
	if err := apiSrv.Start(); err != nil {
		logger.Error("failed to start API server", "error", err)
		return err
	}

	select {
	case <-ctx.Done():
		if apiSrv != nil {
			apiSrv.Close()
		}
		_ = usbSrv.Close()
		_ = <-usbErrCh
		return nil
	case err := <-usbErrCh:
		if apiSrv != nil {
			apiSrv.Close()
		}
		return err
	}
}
