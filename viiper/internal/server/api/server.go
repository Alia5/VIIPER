package api

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"

	"viiper/internal/server/usb"
	"viiper/pkg/device"
	pusb "viiper/pkg/usb"
)

// Server implements a small TCP API for managing virtual bus topology.
type Server struct {
	usbs   *usb.Server
	addr   string
	ln     net.Listener
	logger *slog.Logger
	router *Router
	config ServerConfig
}

// New creates a new ApiServer bound to a server.Server instance.
func New(s *usb.Server, addr string, config ServerConfig, logger *slog.Logger) *Server {
	a := &Server{
		usbs:   s,
		addr:   addr,
		logger: logger,
		config: config,
	}
	a.router = NewRouter()
	return a
}

// Router returns the router used by the API server so callers can register handlers.
func (a *Server) Router() *Router { return a.router }

// USB returns the underlying USB server.
func (a *Server) USB() *usb.Server { return a.usbs }

// Config returns the server configuration.
func (a *Server) Config() ServerConfig { return a.config }

// Start listens on the configured address and serves incoming API commands.
func (a *Server) Start() error {
	ln, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}
	a.ln = ln
	a.logger.Info("API listening", "addr", a.addr)
	go a.serve()
	return nil
}

// Close stops the API server.
func (a *Server) Close() {
	if a.ln != nil {
		_ = a.ln.Close()
	}
}

func (a *Server) serve() {
	for {
		c, err := a.ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || strings.Contains(strings.ToLower(err.Error()), "use of closed network connection") {
				a.logger.Info("API server stopped")
				return
			}
			a.logger.Info("API accept error", "error", err)
			return
		}
		go a.handleConn(c)
	}
}

func (a *Server) writeError(w io.Writer, msg string) {
	problem := map[string]string{"error": msg}
	problemJSON, _ := json.Marshal(problem)
	fmt.Fprintf(w, "%s\n", string(problemJSON))
}

func (a *Server) writeOK(w io.Writer, rest string) {
	if rest == "" {
		fmt.Fprintln(w)
	} else {
		fmt.Fprintf(w, "%s\n", rest)
	}
}

func (a *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	connCtx, connCancel := context.WithCancel(context.Background())
	defer connCancel()

	connLogger := a.logger.With("remote", conn.RemoteAddr().String())
	r := bufio.NewReader(conn)
	w := conn
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				connLogger.Error("read api line", "error", err)
			}
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		connLogger.Info("api cmd", "cmd", line)
		fields := strings.Fields(line)
		if len(fields) == 0 {
			connLogger.Error("api empty command")
			a.writeError(w, "empty")
			continue
		}
		path := strings.ToLower(fields[0])
		args := fields[1:]

		if h, params := a.router.Match(path); h != nil {
			req := &Request{Ctx: connCtx, Params: params, Args: args}
			res := &Response{}
			if err := h(req, res, connLogger); err != nil {
				connLogger.Error("api handler error", "path", path, "error", err)
				a.writeError(w, err.Error())
				continue
			}
			connLogger.Debug("api handler success", "path", path)
			a.writeOK(w, res.JSON)
			continue
		} else if sh, params := a.router.MatchStream(path); sh != nil {
			connLogger.Info("api stream begin", "path", path)
			busIdStr, ok := params["busId"]
			if !ok {
				a.writeError(w, "missing busId parameter")
				return
			}
			devIdStr, ok := params["deviceid"]
			if !ok {
				a.writeError(w, "missing deviceid parameter")
				return
			}

			busID, err := strconv.ParseUint(busIdStr, 10, 32)
			if err != nil {
				a.writeError(w, fmt.Sprintf("invalid busId: %v", err))
				return
			}
			bus := a.usbs.GetBus(uint32(busID))
			if bus == nil {
				a.writeError(w, "bus not found")
				return
			}
			var dev pusb.Device
			var devCtx context.Context
			metas := bus.GetAllDeviceMetas()
			for _, meta := range metas {
				if fmt.Sprintf("%d", meta.Meta.DevId) == devIdStr {
					dev = meta.Dev
					devCtx = bus.GetDeviceContext(dev)
					break
				}
			}
			if dev == nil || devCtx == nil {
				a.writeError(w, "device not found")
				return
			}

			connTimer := device.GetConnTimer(devCtx)
			if connTimer != nil {
				connTimer.Stop()
			}

			// Stream handler takes ownership of connection
			if err := sh(conn, &dev, connLogger); err != nil {
				connLogger.Error("api stream handler error", "path", path, "error", err)
			}
			connLogger.Info("api stream end", "path", path)

			connTimer = device.GetConnTimer(devCtx)
			if connTimer != nil {
				connTimer.Reset(a.config.DeviceHandlerConnectTimeout)
				go func() {
					select {
					case <-devCtx.Done():
						connTimer.Stop()
						return
					case <-connTimer.C:
						exportMeta := device.GetDeviceMeta(devCtx)
						if exportMeta != nil {
							deviceIDStr := fmt.Sprintf("%d", exportMeta.DevId)
							if err := bus.RemoveDeviceByID(deviceIDStr); err != nil {
								connLogger.Error("disconnect timeout: failed to remove device", "busID", busID, "deviceID", deviceIDStr, "error", err)
							} else {
								connLogger.Info("disconnect timeout: removed device (no reconnection)", "busID", busID, "deviceID", deviceIDStr)
							}
						}
					}
				}()
			}

			return
		} else {
			connLogger.Error("api unknown path", "path", path)
			a.writeError(w, "unknown path")
		}

	}
}
