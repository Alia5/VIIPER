package testing

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Alia5/VIIPER/internal/log"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/usb"

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

// ExecCmd dials the API server, sends cmd and reads the full response.
// The command should not include a trailing newline. Returns the response
// without the trailing newline.
func ExecCmd(t *testing.T, addr string, cmd string) string {
	t.Helper()
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer c.Close()

	// Send command with null terminator (\x00) — this matches API server framing
	_, _ = fmt.Fprintf(c, "%s\x00", cmd)

	// Read response
	r := bufio.NewReader(c)
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		t.Fatalf("read failed: %v", err)
	}

	result := strings.TrimSuffix(line, "\n")
	result = strings.TrimSuffix(result, "\r")
	return result
}

// ExecuteLine routes a command string through the provided router,
// emulating ApiServer.handleConn logic but without network IO.
// The data parameter is the full request data (path + optional payload).
// Returns the full response as produced by the API contract.
func ExecuteLine(t *testing.T, r *api.Router, data string) string {
	t.Helper()
	if data == "" {
		return jsonError("empty")
	}

	// Split on first whitespace character using regex \s
	wsRegex := regexp.MustCompile(`\s`)
	loc := wsRegex.FindStringIndex(data)

	var path, payload string
	if loc != nil {
		path = data[:loc[0]]
		payload = data[loc[1]:]
	} else {
		path = data
		payload = ""
	}

	if path == "" {
		return jsonError("empty path")
	}

	path = strings.ToLower(path)

	if h, params := r.Match(path); h != nil {
		req := &api.Request{Params: params, Payload: payload}
		res := &api.Response{}
		if err := h(req, res, slog.Default()); err != nil {
			return jsonError(err.Error())
		}
		if res.JSON == "" {
			return ""
		}
		return res.JSON
	}
	return jsonError("unknown path")
}

func jsonError(msg string) string {
	problem := map[string]string{"error": msg}
	b, _ := json.Marshal(problem)
	return string(b)
}
