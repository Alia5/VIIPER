package apiclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceStream_MockTransportError(t *testing.T) {
	ms := &mockState{responses: map[string]string{}}
	c := WithTransport(newMockTransport(ms))

	_, err := c.OpenStream(context.Background(), 1, "1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported with mock transport")
}

func TestAddDeviceAndConnect_ParsesDeviceID(t *testing.T) {
	ms := &mockState{responses: map[string]string{}}
	ms.responses["bus/{id}/add"] = `{"id":"42-7"}`
	c := WithTransport(newMockTransport(ms))

	// Should parse but fail on stream connection (mock transport)
	_, resp, err := c.AddDeviceAndConnect(context.Background(), 42, "test")
	require.NotNil(t, resp)
	assert.Equal(t, "42-7", resp.ID)
	assert.Error(t, err) // Stream connection will fail with mock
	assert.Contains(t, err.Error(), "not supported with mock transport")
}

func TestDeviceStream_ClosedStreamErrors(t *testing.T) {
	// Create a minimal stream and close it
	s := &DeviceStream{closed: true}

	buf := make([]byte, 10)
	_, err := s.Read(buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream closed")

	_, err = s.Write([]byte("test"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream closed")

	// Close should be idempotent
	assert.NoError(t, s.Close())
}

func TestDeviceStream_Deadlines(t *testing.T) {
	// DeviceStream deadline methods delegate to the underlying connection.
	// Without a real connection we can't test them, so just verify compilation.
	t.Skip("deadline methods require real connection")
}

type mockBinaryMarshaler struct {
	data []byte
}

func (m *mockBinaryMarshaler) MarshalBinary() ([]byte, error) {
	return m.data, nil
}

func (m *mockBinaryMarshaler) UnmarshalBinary(data []byte) error {
	m.data = make([]byte, len(data))
	copy(m.data, data)
	return nil
}

func TestDeviceStream_WriteBinary(t *testing.T) {
	s := &DeviceStream{closed: true}

	msg := &mockBinaryMarshaler{data: []byte("test")}
	err := s.WriteBinary(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream closed")
}

func TestDeviceStream_StartReading_RequiresConnection(t *testing.T) {
	// StartReading requires a real connection to function properly
	t.Skip("StartReading requires real connection for testing")
}
