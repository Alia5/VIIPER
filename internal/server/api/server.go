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

	"github.com/Alia5/VIIPER/apitypes"
	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/internal/server/api/auth"
	apierror "github.com/Alia5/VIIPER/internal/server/api/error"
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
	config *ServerConfig
}

// New creates a new ApiServer bound to a server.Server instance.
func New(s *usb.Server, addr string, config ServerConfig, logger *slog.Logger) *Server {
	cfg := config
	a := &Server{
		usbs:   s,
		addr:   addr,
		logger: logger,
		config: &cfg,
	}
	a.router = NewRouter()
	return a
}

// Router returns the router used by the API server so callers can register handlers.
func (s *Server) Router() *Router { return s.router }

// USB returns the underlying USB server.
func (s *Server) USB() *usb.Server { return s.usbs }

// Config returns the server configuration.
func (s *Server) Config() *ServerConfig { return s.config }

// Addr returns the actual address the server is listening on.
// If Start hasn't been called yet, it returns the configured address.
func (s *Server) Addr() string {
	if s.ln != nil {
		return s.ln.Addr().String()
	}
	return s.addr
}

// Start listens on the configured address and serves incoming API commands.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln

	s.addr = ln.Addr().String()
	s.config.Addr = s.addr
	s.logger.Info("API listening", "addr", s.addr)
	go s.serve()
	return nil
}

// Close stops the API server.
func (s *Server) Close() {
	if s.ln != nil {
		_ = s.ln.Close()
	}
}

func (s *Server) serve() {
	for {
		c, err := s.ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || strings.Contains(strings.ToLower(err.Error()), "use of closed network connection") {
				s.logger.Info("API server stopped")
				return
			}
			s.logger.Info("API accept error", "error", err)
			return
		}
		if tcpConn, ok := c.(*net.TCPConn); ok {
			if err := tcpConn.SetNoDelay(true); err != nil {
				s.logger.Warn("failed to set TCP_NODELAY", "error", err)
			}
		}
		go s.handleConn(c)
	}
}

func (s *Server) writeError(w io.Writer, err error) {
	apiErr := apierror.WrapError(err)
	problemJSON, _ := json.Marshal(apiErr)
	fmt.Fprintf(w, "%s\n", string(problemJSON))
}

func (s *Server) writeOK(w io.Writer, rest string) {
	if rest == "" {
		fmt.Fprintln(w)
	} else {
		fmt.Fprintf(w, "%s\n", rest)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	connCtx, connCancel := context.WithCancel(context.Background())
	defer connCancel()

	connLogger := s.logger.With("remote", conn.RemoteAddr().String())
	r := bufio.NewReader(conn)
	w := conn

	isAuth, err := auth.IsAuthHandshake(r)
	if err != nil {
		connLogger.Error("api handshake check", "error", err)
		// continue as unauthenticated
	}

	if !isAuth && s.requiresAuth(conn.RemoteAddr()) {
		connLogger.Error("authentication required")
		s.writeError(w, apierror.ErrUnauthorized("authentication required"))
		return
	}

	if isAuth {
		connLogger.Debug("Detected auth attempt")
		key, err := auth.DeriveKey(s.config.Password)
		if err != nil {
			connLogger.Error("derive key failed", "error", err)
			return
		}

		clientNonce, serverNonce, err := auth.HandleAuthHandshake(r, w, key, false)
		if err != nil {
			var apiErr apitypes.ApiError
			if errors.As(err, &apiErr) {
				connLogger.Error("auth handshake failed", "error", err)
				s.writeError(w, err)
				return
			}
			connLogger.Error("auth handshake failed", "error", err)
			return
		}

		sessionKey := auth.DeriveSessionKey(key, serverNonce, clientNonce)
		secConn, err := auth.WrapConn(conn, sessionKey)
		if err != nil {
			connLogger.Error("wrap secure conn failed", "error", err)
			return
		}
		conn = secConn
		r = bufio.NewReader(conn)
		w = conn

		connLogger.Debug("authenticated connection established")
	} else {
		connLogger.Debug("continuing unauthenticated connection")
	}

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
		s.writeError(w, apierror.ErrBadRequest("empty request"))
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
		s.writeError(w, apierror.ErrBadRequest("empty path"))
		return
	}

	path = strings.ToLower(path)
	connLogger.Info("api cmd", "path", path)

	if h, params := s.router.Match(path); h != nil {
		req := &Request{Ctx: connCtx, Params: params, Payload: payload}
		res := &Response{}
		if err := h(req, res, connLogger); err != nil {
			connLogger.Error("api handler error", "path", path, "error", err)
			s.writeError(w, err)
			return
		}
		connLogger.Debug("api handler success", "path", path)
		s.writeOK(w, res.JSON)
		return
	} else if sh, params := s.router.MatchStream(path); sh != nil {
		connLogger.Info("api stream begin", "path", path)
		busIDStr, ok := params["busId"]
		if !ok {
			s.writeError(w, apierror.ErrBadRequest("missing busId parameter"))
			return
		}
		devIDStr, ok := params["deviceid"]
		if !ok {
			s.writeError(w, apierror.ErrBadRequest("missing deviceid parameter"))
			return
		}

		busID, err := strconv.ParseUint(busIDStr, 10, 32)
		if err != nil {
			s.writeError(w, apierror.ErrBadRequest(fmt.Sprintf("invalid busId: %v", err)))
			return
		}
		bus := s.usbs.GetBus(uint32(busID))
		if bus == nil {
			s.writeError(w, apierror.ErrNotFound(fmt.Sprintf("bus %d not found", busID)))
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
			s.writeError(w, apierror.ErrNotFound(fmt.Sprintf("device %s not found on bus %d", devIDStr, busID)))
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
			connTimer.Reset(s.config.DeviceHandlerConnectTimeout)
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
	s.writeError(w, apierror.ErrNotFound(fmt.Sprintf("unknown path: %s", path)))
}

func (s *Server) isLocalHostClient(addr net.Addr) bool {
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return false
	}
	switch host {
	case "localhost", "127.0.0.1", "[::1]", "::1":
		return true
	}

	return false
}

func (s *Server) requiresAuth(addr net.Addr) bool {
	if s.isLocalHostClient(addr) {
		return s.config.RequireLocalHostAuth
	}
	return true
}
