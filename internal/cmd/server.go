package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/Alia5/VIIPER/internal/configpaths"
	"github.com/Alia5/VIIPER/internal/log"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/auth"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/internal/server/usb"
	"github.com/Alia5/VIIPER/internal/util"
)

const keyFileName = "viiper.key.txt"

type Server struct {
	UsbServerConfig   usb.ServerConfig `embed:"" prefix:"usb."`
	ApiServerConfig   api.ServerConfig `embed:"" prefix:"api."`
	ConnectionTimeout time.Duration    `help:"ConnectionTimeout operation timeout" default:"30s" env:"VIIPER_CONNECTION_TIMEOUT"`
}

// Run is called by Kong when the server command is executed.
func (s *Server) Run(logger *slog.Logger, rawLogger log.RawLogger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return s.StartServer(ctx, logger, rawLogger)
}

func (s *Server) StartServer(ctx context.Context, logger *slog.Logger, rawLogger log.RawLogger) error {
	s.UsbServerConfig.ConnectionTimeout = s.ConnectionTimeout
	s.ApiServerConfig.ConnectionTimeout = s.ConnectionTimeout
	s.UsbServerConfig.BusCleanupTimeout = s.ApiServerConfig.DeviceHandlerConnectTimeout

	logger.Info("Starting VIIPER USB-IP server", "addr", s.UsbServerConfig.Addr)

	keyFileDir, err := configpaths.DefaultConfigDir()
	if err != nil {
		return fmt.Errorf("failed to resolve key file path: %w", err)
	}
	keyFilePath := path.Join(keyFileDir, keyFileName)
	if pwd, err := os.ReadFile(keyFilePath); err == nil {
		s.ApiServerConfig.Password = strings.TrimSpace(string(pwd))
	} else {
		newPwd, err := auth.GenerateKey()
		if err != nil {
			return fmt.Errorf("failed to generate new API password: %w", err)
		}
		if err := os.MkdirAll(keyFileDir, 0o700); err != nil {
			return fmt.Errorf("failed to create config dir for key file: %w", err)
		}
		if err := os.WriteFile(keyFilePath, []byte(newPwd), 0o600); err != nil {
			return fmt.Errorf("failed to write new API password to file: %w", err)
		}
		s.ApiServerConfig.Password = newPwd
		logger.Info("Generated API server password", "path", keyFilePath)
		logger.Info("-------------------------------------")
		logger.Info("Your VIIPER API server password is:")
		logger.Info("-------------------------------------")
		logger.Info(newPwd)
		logger.Info("-------------------------------------")
		logger.Info("You can change this password at any time by editing the file")
	}

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
	r.Register("ping", handler.Ping())
	r.Register("bus/list", handler.BusList(usbSrv))
	r.Register("bus/create", handler.BusCreate(usbSrv))
	r.Register("bus/remove", handler.BusRemove(usbSrv))
	r.Register("bus/{id}/list", handler.BusDevicesList(usbSrv))
	r.Register("bus/{id}/add", handler.BusDeviceAdd(usbSrv, apiSrv))
	r.Register("bus/{id}/remove", handler.BusDeviceRemove(usbSrv))
	r.RegisterStream("bus/{busId}/{deviceid}", api.DeviceStreamHandler(usbSrv))

	if s.ApiServerConfig.AutoAttachLocalClient {
		logger.Info("Auto-attach is enabled, checking prerequisites...")
		if !api.CheckAutoAttachPrerequisites(s.ApiServerConfig.AutoAttachWindowsNative, logger) {
			logger.Warn("Auto-attach prerequisites not met")
			logger.Warn("Device auto-attachment will fail until requirements are satisfied")
			logger.Info("You can disable auto-attach with --api.auto-attach-local-client=false")
		} else {
			logger.Info("Auto-attach prerequisites satisfied")
		}
	}

	if err := apiSrv.Start(); err != nil {
		logger.Error("failed to start API server", "error", err)
		if util.IsRunFromGUI() {
			fmt.Println("Press any key to exit...")
			var b []byte = make([]byte, 1)
			_, _ = os.Stdin.Read(b)
		}
		return err
	}

	if util.IsRunFromGUI() {
		go (func() {
			time.Sleep(250 * time.Millisecond)
			util.HideConsoleWindow()
		})()
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
