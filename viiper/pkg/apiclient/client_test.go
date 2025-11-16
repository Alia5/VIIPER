package apiclient

import (
	"context"
	"errors"
	"testing"

	apitypes "viiper/pkg/apitypes"

	"github.com/stretchr/testify/assert"
)

// mockTransport captures requests and returns predefined responses.
type mockState struct {
	responses   map[string]string
	err         error
	lastPath    string
	lastPayload string
}

func newMockTransport(ms *mockState) *Transport {
	return NewMockTransport(func(path string, payload any, pathParams map[string]string) (string, error) {
		if ms.err != nil {
			return "", ms.err
		}
		var ps string
		switch v := payload.(type) {
		case string:
			ps = v
		case nil:
			ps = ""
		default:
			ps = "<json>"
		}
		ms.lastPath = path
		ms.lastPayload = ps
		if out, ok := ms.responses[path]; ok {
			return out, nil
		}
		return "", nil
	})
}

func TestHighLevelClient(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(ms *mockState)
		call       func(c *Client) (any, error)
		wantErr    string
		assertFunc func(t *testing.T, got any)
	}{
		{
			name:  "bus create success",
			setup: func(ms *mockState) { ms.responses["bus/create"] = `{"busId":42}` },
			call:  func(c *Client) (any, error) { return c.BusCreate(42) },
			assertFunc: func(t *testing.T, got any) {
				_, ok := got.(*apitypes.BusCreateResponse)
				assert.True(t, ok, "expected *apitypes.BusCreateResponse type")
			},
		},
		{
			name:    "bus create error",
			setup:   func(ms *mockState) { ms.responses["bus/create"] = `{"error":"boom"}` },
			call:    func(c *Client) (any, error) { return c.BusCreate(0) },
			wantErr: "boom",
		},
		{
			name: "devices list",
			setup: func(ms *mockState) {
				ms.responses["bus/{id}/list"] = `{"devices":[{"busId":1,"devId":"1","vid":"0x1234","pid":"0xabcd","type":"x"}]}`
			},
			call: func(c *Client) (any, error) { return c.DevicesList(1) },
			assertFunc: func(t *testing.T, got any) {
				assert.NotNil(t, got)
			},
		},
		{
			name:    "transport failure",
			setup:   func(ms *mockState) { ms.err = errors.New("dial fail") },
			call:    func(c *Client) (any, error) { return c.BusList() },
			wantErr: "dial fail",
		},
		{
			name:    "blank response error",
			setup:   func(ms *mockState) { /* no response set so blank */ },
			call:    func(c *Client) (any, error) { return c.BusList() },
			wantErr: "empty response",
		},
		{
			name: "devices list empty",
			setup: func(ms *mockState) {
				ms.responses["bus/{id}/list"] = `{"devices":[]}`
			},
			call: func(c *Client) (any, error) { return c.DevicesList(1) },
			assertFunc: func(t *testing.T, got any) {
				resp := got.(*apitypes.DevicesListResponse)
				assert.Len(t, resp.Devices, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &mockState{responses: map[string]string{}}
			if tt.setup != nil {
				tt.setup(ms)
			}
			c := WithTransport(newMockTransport(ms))
			got, err := tt.call(c)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			assert.NoError(t, err)
			if tt.assertFunc != nil {
				tt.assertFunc(t, got)
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	// Use a real transport but cancel the context before dialing.
	c := WithTransport(NewTransport("127.0.0.1:9")) // address irrelevant due to early cancel
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.BusListCtx(ctx)
	assert.Error(t, err)
}

func TestStrictJSONDecode(t *testing.T) {
	ms := &mockState{responses: map[string]string{}}
	// extra field should cause decode error due to DisallowUnknownFields
	ms.responses["bus/list"] = `{"buses":[1,2,3],"extra":true}`
	c := WithTransport(newMockTransport(ms))
	_, err := c.BusList()
	assert.Error(t, err)
}
