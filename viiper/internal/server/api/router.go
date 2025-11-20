package api

import (
	"context"
	"log/slog"
	"net"
	"strings"
	"viiper/pkg/usb"
)

// Request contains route parameters and additional args from the command.
type Request struct {
	Ctx    context.Context
	Params map[string]string
	Args   []string
}

// Response holds the JSON string to return to the client.
type Response struct {
	JSON string
}

// HandlerFunc processes a request and populates the response.
// Returns an error on failure. The logger provided is a connection-scoped logger
// enriched with remote address metadata by the API server.
type HandlerFunc func(req *Request, res *Response, logger *slog.Logger) error

// StreamHandlerFunc handles long-lived TCP connections for bidirectional streaming.
// The handler takes ownership of the connection and should close it when done.
// The logger provided is connection-scoped. Returning a non-nil error indicates
// the handler encountered a terminal failure; the dispatcher/server will log it.
type StreamHandlerFunc func(conn net.Conn, dev *usb.Device, logger *slog.Logger) error

// Router implements simple path pattern matching with placeholders in {name}.
type Router struct {
	routes       []routeEntry
	streamRoutes []streamRouteEntry
}

type routeEntry struct {
	pattern         string
	originalPattern string
	parts           []string
	handler         HandlerFunc
}

type streamRouteEntry struct {
	pattern         string
	originalPattern string
	parts           []string
	handler         StreamHandlerFunc
}

// NewRouter returns a new Router instance.
func NewRouter() *Router { return &Router{} }

// Register registers a handler for a path pattern like "bus/{id}/list".
func (r *Router) Register(pattern string, handler HandlerFunc) {
	p := strings.ToLower(pattern)
	parts := strings.Split(p, "/")
	r.routes = append(r.routes, routeEntry{pattern: p, originalPattern: pattern, parts: parts, handler: handler})
}

// RegisterStream registers a StreamHandler for long-lived TCP connections.
func (r *Router) RegisterStream(pattern string, handler StreamHandlerFunc) {
	p := strings.ToLower(pattern)
	parts := strings.Split(p, "/")
	r.streamRoutes = append(r.streamRoutes, streamRouteEntry{pattern: p, originalPattern: pattern, parts: parts, handler: handler})
}

// Match returns the HandlerFunc and params if the given path matches any
// registered pattern. Returns nil if none match.
func (r *Router) Match(path string) (HandlerFunc, map[string]string) {
	p := strings.ToLower(path)
	parts := strings.Split(p, "/")
	for _, rt := range r.routes {
		if len(rt.parts) != len(parts) {
			continue
		}
		params := map[string]string{}
		ok := true
		originalParts := strings.Split(rt.originalPattern, "/")
		for i := range parts {
			if strings.HasPrefix(rt.parts[i], "{") && strings.HasSuffix(rt.parts[i], "}") {

				name := originalParts[i][1 : len(originalParts[i])-1]
				params[name] = parts[i]
				continue
			}
			if rt.parts[i] != parts[i] {
				ok = false
				break
			}
		}
		if ok {
			return rt.handler, params
		}
	}
	return nil, nil
}

// MatchStream returns the StreamHandler and params if the given path matches
// any registered stream pattern. Returns nil if none match.
func (r *Router) MatchStream(path string) (StreamHandlerFunc, map[string]string) {
	p := strings.ToLower(path)
	parts := strings.Split(p, "/")
	for _, rt := range r.streamRoutes {
		if len(rt.parts) != len(parts) {
			continue
		}
		params := map[string]string{}
		ok := true
		originalParts := strings.Split(rt.originalPattern, "/")
		for i := range parts {
			if strings.HasPrefix(rt.parts[i], "{") && strings.HasSuffix(rt.parts[i], "}") {

				name := originalParts[i][1 : len(originalParts[i])-1]
				params[name] = parts[i]
				continue
			}
			if rt.parts[i] != parts[i] {
				ok = false
				break
			}
		}
		if ok {
			return rt.handler, params
		}
	}
	return nil, nil
}
