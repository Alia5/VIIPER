package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alia5/VIIPER/internal/log"
	"github.com/Alia5/VIIPER/internal/server/proxy"
)

type Proxy struct {
	ListenAddr        string        `help:"Proxy listen address" default:":3241" env:"VIIPER_PROXY_ADDR"`
	UpstreamAddr      string        `help:"Upstream USB-IP server address" required:"" env:"VIIPER_PROXY_UPSTREAM"`
	ConnectionTimeout time.Duration `help:"Connection timeout" default:"30s" env:"VIIPER_PROXY_TIMEOUT"`
}

// Run is called by Kong when the proxy command is executed.
func (p *Proxy) Run(logger *slog.Logger, rawLogger log.RawLogger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if p.UpstreamAddr == "" {
		return errors.New("Upstream address is empty")
	}

	logger.Info("Starting VIIPER USB-IP proxy", "listen", p.ListenAddr, "upstream", p.UpstreamAddr)
	proxySrv := proxy.New(p.ListenAddr, p.UpstreamAddr, p.ConnectionTimeout, logger, rawLogger)

	proxyErrCh := make(chan error, 1)
	go func() {
		proxyErrCh <- proxySrv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutting down proxy server")
		_ = proxySrv.Close()
		_ = <-proxyErrCh
		return nil
	case err := <-proxyErrCh:
		return err
	}
}
