package viiperclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/viipertypes"
)

// Client provides a high-level interface to the VIIPER API, handling request
// formatting, response parsing, and error handling.
type Client struct{ transport *Transport }

// New constructs a high-level API client using the internal low-level Transport.
// The addr parameter specifies the TCP address (host:port) of the VIIPER API server.
func New(addr string) *Client { return &Client{transport: NewTransport(addr)} }

// NewWithPassword constructs a client that authenticates with the given password.
func NewWithPassword(addr, password string) *Client {
	return &Client{transport: NewTransportWithPassword(addr, password)}
}

// NewWithConfig constructs a client with custom transport timeouts.
func NewWithConfig(addr string, cfg *Config) *Client {
	return &Client{transport: NewTransportWithConfig(addr, cfg)}
}

// WithTransport constructs a Client using a custom Transport implementation.
// This is primarily useful for testing or when advanced transport configuration is needed.
func WithTransport(t *Transport) *Client { return &Client{transport: t} }

// Ping returns the version and identity of the VIIPER server.
func (c *Client) Ping() (*viipertypes.PingResponse, error) {
	return c.PingCtx(context.Background())
}

// PingCtx is the context-aware version of Ping.
func (c *Client) PingCtx(ctx context.Context) (*viipertypes.PingResponse, error) {
	const path = "ping"
	raw, err := c.transport.DoCtx(ctx, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return parse[viipertypes.PingResponse](raw)
}

// BusCreate creates a new virtual USB bus with the specified bus number.
// Returns the created bus ID or an error if the bus number is already allocated.
func (c *Client) BusCreate(busID uint32) (*viipertypes.BusCreateResponse, error) {
	return c.BusCreateCtx(context.Background(), busID)
}

func (c *Client) BusCreateCtx(ctx context.Context, busID uint32) (*viipertypes.BusCreateResponse, error) {
	const path = "bus/create"
	raw, err := c.transport.DoCtx(ctx, path, fmt.Sprintf("%d", busID), nil)
	if err != nil {
		return nil, err
	}
	return parse[viipertypes.BusCreateResponse](raw)
}

// BusRemove removes an existing virtual USB bus and all devices attached to it.
// Returns the removed bus ID or an error if the bus does not exist.
func (c *Client) BusRemove(busID uint32) (*viipertypes.BusRemoveResponse, error) {
	return c.BusRemoveCtx(context.Background(), busID)
}

func (c *Client) BusRemoveCtx(ctx context.Context, busID uint32) (*viipertypes.BusRemoveResponse, error) {
	const path = "bus/remove"
	raw, err := c.transport.DoCtx(ctx, path, fmt.Sprintf("%d", busID), nil)
	if err != nil {
		return nil, err
	}
	return parse[viipertypes.BusRemoveResponse](raw)
}

// BusList retrieves a list of all active virtual USB bus numbers.
func (c *Client) BusList() (*viipertypes.BusListResponse, error) {
	return c.BusListCtx(context.Background())
}

func (c *Client) BusListCtx(ctx context.Context) (*viipertypes.BusListResponse, error) {
	const path = "bus/list"
	raw, err := c.transport.DoCtx(ctx, path, nil, nil)
	if err != nil {
		return nil, err
	}
	return parse[viipertypes.BusListResponse](raw)
}

// DeviceAdd adds a new device of the specified type to the given bus.
// The devType parameter specifies the device type (e.g., "xbox360").
// Returns the assigned bus ID (e.g., "1-1") or an error if the bus does not exist
// or the device type is unknown.
func (c *Client) DeviceAdd(busID uint32, devType string, o *device.CreateOptions) (*viipertypes.Device, error) {
	return c.DeviceAddCtx(context.Background(), busID, devType, o)
}

func (c *Client) DeviceAddCtx(ctx context.Context, busID uint32, devType string, o *device.CreateOptions) (*viipertypes.Device, error) {
	pathParams := map[string]string{"id": fmt.Sprintf("%d", busID)}
	const path = "bus/{id}/add"

	if o == nil {
		o = &device.CreateOptions{}
	}
	var deviceSpecific map[string]any
	if o.DeviceSpecific != "" {
		err := json.Unmarshal([]byte(o.DeviceSpecific), &deviceSpecific)
		if err != nil {
			return nil, fmt.Errorf("invalid CreateOptions.DeviceSpecific JSON: %w", err)
		}
	}
	req := viipertypes.DeviceCreateRequest{
		Type:           &devType,
		IDVendor:       o.IDVendor,
		IDProduct:      o.IDProduct,
		DeviceSpecific: deviceSpecific,
	}
	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal device create request: %w", err)
	}
	raw, err := c.transport.DoCtx(ctx, path, string(payloadBytes), pathParams)
	if err != nil {
		return nil, err
	}
	return parse[viipertypes.Device](raw)
}

// DeviceRemove removes a device from the specified bus by its device ID.
// The devID parameter is the device number (e.g., "1") on the given bus.
// Active USB-IP connections to the device will be closed.
// Returns the removed device's bus and device ID or an error if not found.
func (c *Client) DeviceRemove(busID uint32, devID string) (*viipertypes.DeviceRemoveResponse, error) {
	return c.DeviceRemoveCtx(context.Background(), busID, devID)
}

func (c *Client) DeviceRemoveCtx(ctx context.Context, busID uint32, devID string) (*viipertypes.DeviceRemoveResponse, error) {
	pathParams := map[string]string{"id": fmt.Sprintf("%d", busID)}
	const path = "bus/{id}/remove"
	raw, err := c.transport.DoCtx(ctx, path, devID, pathParams)
	if err != nil {
		return nil, err
	}
	return parse[viipertypes.DeviceRemoveResponse](raw)
}

// DevicesList retrieves a list of all devices attached to the specified bus.
// Each device entry includes bus ID, device ID, VID, PID, and device type.
func (c *Client) DevicesList(busID uint32) (*viipertypes.DevicesListResponse, error) {
	return c.DevicesListCtx(context.Background(), busID)
}

func (c *Client) DevicesListCtx(ctx context.Context, busID uint32) (*viipertypes.DevicesListResponse, error) {
	pathParams := map[string]string{"id": fmt.Sprintf("%d", busID)}
	const path = "bus/{id}/list"
	raw, err := c.transport.DoCtx(ctx, path, nil, pathParams)
	if err != nil {
		return nil, err
	}
	return parse[viipertypes.DevicesListResponse](raw)
}

func parse[T any](data string) (*T, error) {
	if data == "" {
		return nil, errors.New("empty response")
	}
	var problem viipertypes.APIError
	if err := json.Unmarshal([]byte(data), &problem); err == nil && (problem.Status != 0 || problem.Title != "") {
		return nil, &problem
	}
	var out T
	dec := json.NewDecoder(bytes.NewReader([]byte(data)))
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &out, nil
}
