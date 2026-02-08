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

func TestBusRemove(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T, s *usb.Server)
		payload          any
		expectedResponse string
	}{
		{
			name: "remove existing bus",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(70001)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			payload:          "70001",
			expectedResponse: `{"busId":70001}`,
		},
		{
			name: "remove bus and reuse bus number",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(70002)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			payload:          "70002",
			expectedResponse: `{"busId":70002}`,
		},
		{
			name:             "remove non-existing bus",
			setup:            nil,
			payload:          "99999",
			expectedResponse: `{"status":404,"title":"Not Found","detail":"bus 99999 not found"}`,
		},
		{
			name: "remove bus with devices attached",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(70004)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
				dev, err := xbox360.New(nil)
				if err != nil {
					t.Fatalf("create device 1 failed: %v", err)
				}
				if _, err := b.Add(dev); err != nil {
					t.Fatalf("add device 1 failed: %v", err)
				}
				dev, err = xbox360.New(nil)
				if err != nil {
					t.Fatalf("create device 2 failed: %v", err)
				}
				if _, err := b.Add(dev); err != nil {
					t.Fatalf("add device 2 failed: %v", err)
				}
			},
			payload:          "70004",
			expectedResponse: `{"busId":70004}`,
		},
		{
			name:             "invalid bus number",
			setup:            nil,
			payload:          "bar",
			expectedResponse: `{"status":400,"title":"Bad Request","detail":"invalid busId: strconv.ParseUint: parsing \"bar\": invalid syntax"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, srv, done := handlerTest.StartAPIServer(t, func(r *api.Router, s *usb.Server, apiSrv *api.Server) {
				r.Register("bus/create", handler.BusCreate(s))
				r.Register("bus/remove", handler.BusRemove(s))
			})
			defer done()

			c := apiclient.NewTransport(addr)
			if tt.setup != nil {
				tt.setup(t, srv)
			}
			line, err := c.Do("bus/remove", tt.payload, nil)
			assert.NoError(t, err)
			if tt.expectedResponse[0] == '{' {
				assert.JSONEq(t, tt.expectedResponse, line)
			} else {
				assert.Equal(t, tt.expectedResponse, line)
			}

			if tt.name == "remove bus and reuse bus number" {
				b, err := virtualbus.NewWithBusId(70002)
				assert.NoError(t, err, "should be able to reuse bus number after removal")
				err = srv.AddBus(b)
				assert.NoError(t, err, "should be able to add bus with reused number")
			}
		})
	}
}
