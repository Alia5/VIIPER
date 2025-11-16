package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	apitypes "viiper/pkg/apitypes"
)

// Client provides a high-level interface to the VIIPER API, handling request
// formatting, response parsing, and error handling.
type Client struct{ transport *Transport }

// New constructs a high-level API client using the internal low-level Transport.
// The addr parameter specifies the TCP address (host:port) of the VIIPER API server.
func New(addr string) *Client { return &Client{transport: NewTransport(addr)} }

// NewWithConfig constructs a client with custom transport timeouts.
func NewWithConfig(addr string, cfg *Config) *Client {
	return &Client{transport: NewTransportWithConfig(addr, cfg)}
}

// WithTransport constructs a Client using a custom Transport implementation.
// This is primarily useful for testing or when advanced transport configuration is needed.
func WithTransport(t *Transport) *Client { return &Client{transport: t} }

// BusCreate creates a new virtual USB bus with the specified bus number.
// Returns the created bus ID or an error if the bus number is already allocated.
func (c *Client) BusCreate(busID uint32) (*apitypes.BusCreateResponse, error) {
	return c.BusCreateCtx(context.Background(), busID)
}

func (c *Client) BusCreateCtx(ctx context.Context, busID uint32) (*apitypes.BusCreateResponse, error) {
	const path = "bus/create"
	line, err := c.transport.DoCtx(ctx, path, fmt.Sprintf("%d", busID), nil)
	if err != nil {
		return nil, err
	}
	return parse[apitypes.BusCreateResponse](line)
}

// BusRemove removes an existing virtual USB bus and all devices attached to it.
// Returns the removed bus ID or an error if the bus does not exist.
func (c *Client) BusRemove(busID uint32) (*apitypes.BusRemoveResponse, error) {
	return c.BusRemoveCtx(context.Background(), busID)
}

func (c *Client) BusRemoveCtx(ctx context.Context, busID uint32) (*apitypes.BusRemoveResponse, error) {
	const path = "bus/remove"
	line, err := c.transport.DoCtx(ctx, path, fmt.Sprintf("%d", busID), nil)
	if err != nil {
		return nil, err
	}
	return parse[apitypes.BusRemoveResponse](line)
}

// BusList retrieves a list of all active virtual USB bus numbers.
func (c *Client) BusList() (*apitypes.BusListResponse, error) {
	return c.BusListCtx(context.Background())
}

func (c *Client) BusListCtx(ctx context.Context) (*apitypes.BusListResponse, error) {
	const path = "bus/list"
	line, err := c.transport.DoCtx(ctx, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return parse[apitypes.BusListResponse](line)
}

// DeviceAdd adds a new device of the specified type to the given bus.
// The devType parameter specifies the device type (e.g., "xbox360").
// Returns the assigned bus ID (e.g., "1-1") or an error if the bus does not exist
// or the device type is unknown.
func (c *Client) DeviceAdd(busID uint32, devType string) (*apitypes.DeviceAddResponse, error) {
	return c.DeviceAddCtx(context.Background(), busID, devType)
}

func (c *Client) DeviceAddCtx(ctx context.Context, busID uint32, devType string) (*apitypes.DeviceAddResponse, error) {
	pathParams := map[string]string{"id": fmt.Sprintf("%d", busID)}
	const path = "bus/{id}/add"
	line, err := c.transport.DoCtx(ctx, path, devType, pathParams)
	if err != nil {
		return nil, err
	}
	return parse[apitypes.DeviceAddResponse](line)
}

// DeviceRemove removes a device from the specified bus by its device ID.
// The busid parameter is the device number (e.g., "1") on the given bus.
// Active USB-IP connections to the device will be closed.
// Returns the removed device's bus and device ID or an error if not found.
func (c *Client) DeviceRemove(busID uint32, busid string) (*apitypes.DeviceRemoveResponse, error) {
	return c.DeviceRemoveCtx(context.Background(), busID, busid)
}

func (c *Client) DeviceRemoveCtx(ctx context.Context, busID uint32, busid string) (*apitypes.DeviceRemoveResponse, error) {
	pathParams := map[string]string{"id": fmt.Sprintf("%d", busID)}
	const path = "bus/{id}/remove"
	line, err := c.transport.DoCtx(ctx, path, busid, pathParams)
	if err != nil {
		return nil, err
	}
	return parse[apitypes.DeviceRemoveResponse](line)
}

// DevicesList retrieves a list of all devices attached to the specified bus.
// Each device entry includes bus ID, device ID, VID, PID, and device type.
func (c *Client) DevicesList(busID uint32) (*apitypes.DevicesListResponse, error) {
	return c.DevicesListCtx(context.Background(), busID)
}

func (c *Client) DevicesListCtx(ctx context.Context, busID uint32) (*apitypes.DevicesListResponse, error) {
	pathParams := map[string]string{"id": fmt.Sprintf("%d", busID)}
	// Endpoint path corrected to match server registration ("bus/{id}/list").
	// Previously used "bus/{id}/devices" which is not registered by the API server.
	const path = "bus/{id}/list"
	line, err := c.transport.DoCtx(ctx, path, nil, pathParams)
	if err != nil {
		return nil, err
	}
	return parse[apitypes.DevicesListResponse](line)
}

func parse[T any](line string) (*T, error) {
	if line == "" {
		return nil, errors.New("empty response")
	}
	var ae apitypes.ApiError
	if err := json.Unmarshal([]byte(line), &ae); err == nil && ae.Error != "" {
		return nil, errors.New(ae.Error)
	}
	var out T
	dec := json.NewDecoder(bytes.NewReader([]byte(line)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &out, nil
}
