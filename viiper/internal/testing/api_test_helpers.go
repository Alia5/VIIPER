package testing

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"viiper/internal/log"
	"viiper/internal/server/api"
	"viiper/internal/server/usb"

	"log/slog"
)

// StartAPIServer starts an API server on a free port and calls register to allow
// the caller to register the handlers needed for the test. Returns the address
// and a function to call when done.
func StartAPIServer(t *testing.T, register func(r *api.Router, s *usb.Server, apiSrv *api.Server)) (addr string, srv *usb.Server, done func()) {
	t.Helper()
	cfg := usb.ServerConfig{
		Addr: "127.0.0.1:0",
	}
	srv = usb.New(cfg, slog.Default(), log.NewRaw(nil))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	addr = ln.Addr().String()
	_ = ln.Close()

	apiSrv := api.New(srv, addr, api.ServerConfig{}, slog.Default())
	if register != nil {
		register(apiSrv.Router(), srv, apiSrv)
	}
	if err := apiSrv.Start(); err != nil {
		t.Fatalf("api start failed: %v", err)
	}

	done = func() {
		apiSrv.Close()
		time.Sleep(10 * time.Millisecond)
	}
	return addr, srv, done
}

// ExecCmd dials the API server, sends cmd (newline not required) and returns
// the full preliminary response line (OK/ERR plus payload). The caller should
// inspect the response. Client errors call t.Fatalf.
// ExecCmd executes a raw command against a running API server and returns the full
// response line (including JSON payload if present) without the trailing newline.
func ExecCmd(t *testing.T, addr string, cmd string) string {
	t.Helper()
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer c.Close()
	r := bufio.NewReader(c)
	_, _ = fmt.Fprintf(c, "%s\n", cmd)
	line, err := r.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			t.Fatalf("read failed: %v", err)
		}
	}
	if len(line) == 0 {
		return ""
	}
	return line[:len(line)-1] // strip newline; may be empty string
}

// RunAPICmd executes a command through the ApiServer's connection handler using in-memory pipes.
// It exercises routing and handler invocation without a network listener.
// ExecuteLine routes a single command string (one full line without trailing newline)
// through the provided router, emulating ApiServer.handleConn logic but without network IO.
// Returns the full response line (without trailing newline) as produced by the API contract.
func ExecuteLine(t *testing.T, r *api.Router, line string) string {
	t.Helper()
	line = strings.TrimSpace(line)
	if line == "" {
		return jsonError("empty")
	}
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return jsonError("empty")
	}
	path := strings.ToLower(fields[0])
	args := fields[1:]
	if h, params := r.Match(path); h != nil {
		req := &api.Request{Params: params, Args: args}
		res := &api.Response{}
		if err := h(req, res, slog.Default()); err != nil {
			return jsonError(err.Error())
		}
		if res.JSON == "" {
			return "OK"
		}
		return "OK " + res.JSON
	}
	return jsonError("unknown path")
}

func jsonError(msg string) string {
	problem := map[string]string{"error": msg}
	b, _ := json.Marshal(problem)
	return string(b)
}
