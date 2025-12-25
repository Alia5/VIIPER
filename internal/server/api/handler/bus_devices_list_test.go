package handler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Alia5/VIIPER/apiclient"
	"github.com/Alia5/VIIPER/device/xbox360"
	handlerTest "github.com/Alia5/VIIPER/internal/_testing"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/internal/server/usb"
	"github.com/Alia5/VIIPER/virtualbus"
)

func TestBusDevicesList(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T, s *usb.Server)
		pathParams       map[string]string
		expectedResponse string
	}{
		{
			name: "list devices on existing bus",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(60008)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "60008"},
			expectedResponse: `{"devices":[]}`,
		},
		{
			name: "list devices after adding one",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(60009)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
				if _, err := b.Add(xbox360.New(nil)); err != nil {
					t.Fatalf("add device failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "60009"},
			expectedResponse: `{"devices":[{"busId":60009,"devId":"1","vid":"0x045e","pid":"0x028e","type":"xbox360"}]}`,
		},
		{
			name: "list devices with multiple additions",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(60010)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
				if _, err := b.Add(xbox360.New(nil)); err != nil {
					t.Fatalf("add device 1 failed: %v", err)
				}
				if _, err := b.Add(xbox360.New(nil)); err != nil {
					t.Fatalf("add device 2 failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "60010"},
			expectedResponse: `{"devices":[{"busId":60010,"devId":"1","vid":"0x045e","pid":"0x028e","type":"xbox360"},{"busId":60010,"devId":"2","vid":"0x045e","pid":"0x028e","type":"xbox360"}]}`,
		},
		{
			name:             "list devices on non-existing bus",
			setup:            nil,
			pathParams:       map[string]string{"id": "99999"},
			expectedResponse: `{"status":404,"title":"Not Found","detail":"bus 99999 not found"}`,
		},
		{
			name:             "invalid bus number",
			setup:            nil,
			pathParams:       map[string]string{"id": "abc"},
			expectedResponse: `{"status":400,"title":"Bad Request","detail":"invalid busId: strconv.ParseUint: parsing \"abc\": invalid syntax"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, srv, done := handlerTest.StartAPIServer(t, func(r *api.Router, s *usb.Server, apiSrv *api.Server) {
				r.Register("bus/{id}/list", handler.BusDevicesList(s))
			})
			defer done()

			c := apiclient.NewTransport(addr)
			if tt.setup != nil {
				tt.setup(t, srv)
			}
			line, err := c.Do("bus/{id}/list", nil, tt.pathParams)
			assert.NoError(t, err)
			if tt.expectedResponse[0] == '{' {
				assert.JSONEq(t, tt.expectedResponse, line)
			} else {
				assert.Equal(t, tt.expectedResponse, line)
			}
		})
	}
}
