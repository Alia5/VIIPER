package handler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Alia5/VIIPER/apiclient"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/internal/server/usb"
	handlerTest "github.com/Alia5/VIIPER/internal/testing"
	"github.com/Alia5/VIIPER/virtualbus"
)

func TestBusList(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T, s *usb.Server)
		expectedResponse string
	}{
		{
			name:             "empty list",
			setup:            nil,
			expectedResponse: `{"buses":[]}`,
		},
		{
			name: "list with one bus",
			setup: func(t *testing.T, s *usb.Server) {
				b, err := virtualbus.NewWithBusId(60005)
				if err != nil {
					t.Fatalf("create bus failed: %v", err)
				}
				if err := s.AddBus(b); err != nil {
					t.Fatalf("add bus failed: %v", err)
				}
			},
			expectedResponse: `{"buses":[60005]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, srv, done := handlerTest.StartAPIServer(t, func(r *api.Router, s *usb.Server, apiSrv *api.Server) {
				r.Register("bus/list", handler.BusList(s))
			})
			defer done()

			c := apiclient.NewTransport(addr)
			if tt.setup != nil {
				tt.setup(t, srv)
			}
			line, err := c.Do("bus/list", nil, nil)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResponse, line)
		})
	}
}
