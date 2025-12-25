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
	"regexp"
	"strconv"
	"strings"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/internal/server/usb"
	pusb "github.com/Alia5/VIIPER/usb"
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

// Addr returns the actual address the server is listening on.
// If Start hasn't been called yet, it returns the configured address.
func (a *Server) Addr() string {
	if a.ln != nil {
		return a.ln.Addr().String()
	}
	return a.addr
}

// Start listens on the configured address and serves incoming API commands.
func (a *Server) Start() error {
	ln, err := net.Listen("tcp", a.addr)
	if err != nil {
		return err
	}
	a.ln = ln

	a.addr = ln.Addr().String()
	a.config.Addr = a.addr
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
		if tcpConn, ok := c.(*net.TCPConn); ok {
			if err := tcpConn.SetNoDelay(true); err != nil {
				a.logger.Warn("failed to set TCP_NODELAY", "error", err)
			}
		}
		go a.handleConn(c)
	}
}

func (a *Server) writeError(w io.Writer, err error) {
	apiErr := WrapError(err)
	problemJSON, _ := json.Marshal(apiErr)
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

	// Read until null terminator
	reqData, err := r.ReadString('\x00')
	if err != nil {
		if err == io.EOF {
			connLogger.Error("api incomplete request (no null terminator)")
		} else {
			connLogger.Error("read api data", "error", err)
		}
		return
	}
	// Remove null terminator
	reqData = strings.TrimSuffix(reqData, "\x00")

	if reqData == "" {
		connLogger.Error("api empty command")
		a.writeError(w, ErrBadRequest("empty request"))
		return
	}

	// Split on first whitespace character using regex \s
	wsRegex := regexp.MustCompile(`\s`)
	loc := wsRegex.FindStringIndex(reqData)

	var path, payload string
	if loc != nil {
		path = reqData[:loc[0]]
		payload = reqData[loc[1]:]
	} else {
		path = reqData
		payload = ""
	}

	if path == "" {
		connLogger.Error("api empty path")
		a.writeError(w, ErrBadRequest("empty path"))
		return
	}

	path = strings.ToLower(path)
	connLogger.Info("api cmd", "path", path)

	if h, params := a.router.Match(path); h != nil {
		req := &Request{Ctx: connCtx, Params: params, Payload: payload}
		res := &Response{}
		if err := h(req, res, connLogger); err != nil {
			connLogger.Error("api handler error", "path", path, "error", err)
			a.writeError(w, err)
			return
		}
		connLogger.Debug("api handler success", "path", path)
		a.writeOK(w, res.JSON)
		return
	} else if sh, params := a.router.MatchStream(path); sh != nil {
		connLogger.Info("api stream begin", "path", path)
		busIDStr, ok := params["busId"]
		if !ok {
			a.writeError(w, ErrBadRequest("missing busId parameter"))
			return
		}
		devIDStr, ok := params["deviceid"]
		if !ok {
			a.writeError(w, ErrBadRequest("missing deviceid parameter"))
			return
		}

		busID, err := strconv.ParseUint(busIDStr, 10, 32)
		if err != nil {
			a.writeError(w, ErrBadRequest(fmt.Sprintf("invalid busId: %v", err)))
			return
		}
		bus := a.usbs.GetBus(uint32(busID))
		if bus == nil {
			a.writeError(w, ErrNotFound(fmt.Sprintf("bus %d not found", busID)))
			return
		}
		var dev pusb.Device
		var devCtx context.Context
		metas := bus.GetAllDeviceMetas()
		for _, meta := range metas {
			if fmt.Sprintf("%d", meta.Meta.DevId) == devIDStr {
				dev = meta.Dev
				devCtx = bus.GetDeviceContext(dev)
				break
			}
		}
		if dev == nil || devCtx == nil {
			a.writeError(w, ErrNotFound(fmt.Sprintf("device %s not found on bus %d", devIDStr, busID)))
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
						return
					}
					connLogger.Warn("disconnect timeout: device context closed but metadata missing")
				}
			}()
		}

		return
	}
	connLogger.Error("api unknown path", "path", path)
	a.writeError(w, ErrNotFound(fmt.Sprintf("unknown path: %s", path)))
}
