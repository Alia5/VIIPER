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

func TestBusDeviceRemove(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T, s *usb.Server)
		pathParams       map[string]string
		payload          any
		expectedResponse string
	}{
		{
			name: "remove existing device",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(90001)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
				dev, err := xbox360.New(nil)
				if err != nil {
					t.Fatalf("create device failed: %v", err)
				}
				if _, err := b.Add(dev); err != nil {
					t.Fatalf("add device failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "90001"},
			payload:          "1",
			expectedResponse: `{"busId":90001,"devId":"1"}`,
		},
		{
			name:             "remove from non-existing bus",
			setup:            nil,
			pathParams:       map[string]string{"id": "90001"},
			payload:          "1",
			expectedResponse: `{"status":404,"title":"Not Found","detail":"bus 90001 not found"}`,
		},
		{
			name: "remove non-existing device",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(90002)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			pathParams:       map[string]string{"id": "90002"},
			payload:          "1",
			expectedResponse: `{"status":404,"title":"Not Found","detail":"device 1 not found on bus 90002"}`,
		},
		{
			name:             "invalid bus number",
			setup:            nil,
			pathParams:       map[string]string{"id": "abc"},
			payload:          "1",
			expectedResponse: `{"status":400,"title":"Bad Request","detail":"invalid busId: strconv.ParseUint: parsing \"abc\": invalid syntax"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, srv, done := handlerTest.StartAPIServer(t, func(r *api.Router, s *usb.Server, apiSrv *api.Server) {
				r.Register("bus/{id}/remove", handler.BusDeviceRemove(s))
			})
			defer done()

			c := apiclient.NewTransport(addr)
			if tt.setup != nil {
				tt.setup(t, srv)
			}
			line, err := c.Do("bus/{id}/remove", tt.payload, tt.pathParams)
			assert.NoError(t, err)
			if tt.expectedResponse[0] == '{' {
				assert.JSONEq(t, tt.expectedResponse, line)
			} else {
				assert.Equal(t, tt.expectedResponse, line)
			}
		})
	}
}
