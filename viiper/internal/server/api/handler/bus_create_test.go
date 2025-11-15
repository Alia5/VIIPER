package handler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"viiper/internal/server/api"
	"viiper/internal/server/api/handler"
	"viiper/internal/server/usb"
	handlerTest "viiper/internal/testing"
	"viiper/pkg/apiclient"
	"viiper/pkg/virtualbus"
)

func TestBusCreate(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T, s *usb.Server)
		payload          any
		expectedResponse string
	}{
		{
			name:             "valid create",
			setup:            nil,
			payload:          "60001",
			expectedResponse: `{"busId":60001}`,
		},
		{
			name: "duplicate bus",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(60002)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			payload:          "60002",
			expectedResponse: `{"error":"bus number 60002 already allocated"}`,
		},
		{
			name: "create after remove allows reuse",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(60003)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
				if err := s.RemoveBus(60003); err != nil {
					t.Fatalf("remove bus failed: %v", err)
				}
			},
			payload:          "60003",
			expectedResponse: `{"busId":60003}`,
		},
		{
			name:             "invalid bus number",
			setup:            nil,
			payload:          "foo",
			expectedResponse: `{"error":"strconv.ParseUint: parsing \"foo\": invalid syntax"}`,
		},
		{
			name:             "negative bus number",
			setup:            nil,
			payload:          "-1",
			expectedResponse: `{"error":"strconv.ParseUint: parsing \"-1\": invalid syntax"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, srv, done := handlerTest.StartAPIServer(t, func(r *api.Router, s *usb.Server, apiSrv *api.Server) {
				r.Register("bus/create", handler.BusCreate(s))
			})
			defer done()
			c := apiclient.NewTransport(addr)
			if tt.setup != nil {
				tt.setup(t, srv)
			}
			line, err := c.Do("bus/create", tt.payload, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResponse, line)
		})
	}
}
