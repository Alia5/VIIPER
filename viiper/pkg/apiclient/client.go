package apiclient

import (
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

// WithTransport constructs a Client using a custom Transport implementation.
// This is primarily useful for testing or when advanced transport configuration is needed.
func WithTransport(t *Transport) *Client { return &Client{transport: t} }

type errorEnvelope struct {
	Error string `json:"error"`
}

// BusCreate creates a new virtual USB bus with the specified bus number.
// Returns the created bus ID or an error if the bus number is already allocated.
func (c *Client) BusCreate(busID uint32) (*apitypes.BusCreateResponse, error) {
	line, err := c.transport.Do("bus/create", fmt.Sprintf("%d", busID), nil)
	if err != nil {
		return nil, err
	}
	return parse[apitypes.BusCreateResponse](line)
}

// BusRemove removes an existing virtual USB bus and all devices attached to it.
// Returns the removed bus ID or an error if the bus does not exist.
func (c *Client) BusRemove(busID uint32) (*apitypes.BusRemoveResponse, error) {
	line, err := c.transport.Do("bus/remove", fmt.Sprintf("%d", busID), nil)
	if err != nil {
		return nil, err
	}
	return parse[apitypes.BusRemoveResponse](line)
}

// BusList retrieves a list of all active virtual USB bus numbers.
func (c *Client) BusList() (*apitypes.BusListResponse, error) {
	line, err := c.transport.Do("bus/list", nil, nil)
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
	pathParams := map[string]string{"id": fmt.Sprintf("%d", busID)}
	line, err := c.transport.Do("bus/{id}/add", devType, pathParams)
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
	pathParams := map[string]string{"id": fmt.Sprintf("%d", busID)}
	line, err := c.transport.Do("bus/{id}/remove", busid, pathParams)
	if err != nil {
		return nil, err
	}
	return parse[apitypes.DeviceRemoveResponse](line)
}

// DevicesList retrieves a list of all devices attached to the specified bus.
// Each device entry includes bus ID, device ID, VID, PID, and device type.
func (c *Client) DevicesList(busID uint32) (*apitypes.DevicesListResponse, error) {
	pathParams := map[string]string{"id": fmt.Sprintf("%d", busID)}
	line, err := c.transport.Do("bus/{id}/devices", nil, pathParams)
	if err != nil {
		return nil, err
	}
	return parse[apitypes.DevicesListResponse](line)
}

func parse[T any](line string) (*T, error) {
	if line == "" {
		return nil, errors.New("empty response")
	}
	var ee errorEnvelope
	if err := json.Unmarshal([]byte(line), &ee); err == nil && ee.Error != "" {
		return nil, errors.New(ee.Error)
	}
	var out T
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return &out, nil
}
