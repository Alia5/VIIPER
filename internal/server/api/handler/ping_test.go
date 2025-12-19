package handler_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Alia5/VIIPER/apiclient"
	"github.com/Alia5/VIIPER/apitypes"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/internal/server/usb"
	handlerTest "github.com/Alia5/VIIPER/internal/testing"
)

func TestPing(t *testing.T) {
	addr, _, done := handlerTest.StartAPIServer(t, func(r *api.Router, s *usb.Server, apiSrv *api.Server) {
		r.Register("ping", handler.Ping())
	})
	defer done()

	c := apiclient.NewTransport(addr)
	line, err := c.Do("ping", nil, nil)
	assert.NoError(t, err)

	var out apitypes.PingResponse
	err = json.Unmarshal([]byte(line), &out)
	assert.NoError(t, err)
	assert.Equal(t, "VIIPER", out.Server)
	assert.NotEmpty(t, out.Version)
}
