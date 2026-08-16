package viiperclient

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestDeviceStream_CloseRacesWithReadAndWrite(t *testing.T) {
	client, peer := net.Pipe()
	defer peer.Close()

	stream := &DeviceStream{conn: client}
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			if _, err := stream.Write([]byte{0x01}); err != nil {
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		buf := make([]byte, 1)
		for {
			if _, err := stream.Read(buf); err != nil {
				return
			}
		}
	}()

	// Give both operations time to enter their loops before closing the
	// connection. The test is intentionally concurrent; -race verifies the
	// closed flag is not accessed as an ordinary bool.
	time.Sleep(10 * time.Millisecond)
	if err := stream.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	wg.Wait()

	if err := stream.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}
