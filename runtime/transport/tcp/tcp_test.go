package tcp_test

import (
	"sync"
	"testing"
	"time"

	"github.com/nickng/scribble-go/runtime/transport/tcp"
)

// This tests establishing connection.
func TestConnection(t *testing.T) {
	const (
		Host = "localhost"
		Port = "12345"
	)
	c := tcp.NewConnection(Host, Port)
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(c tcp.ConnCfg) {
		defer wg.Done()
		server := c.Accept()
		server.Close()
	}(c)

	go func(c tcp.ConnCfg) {
		defer wg.Done()
		client := c.Connect()
		client.Close()
	}(c)

	wg.Wait()
	t.Logf("Connection established")
}

// This tests sending and receiving a single byte stream over TCP.
// The test only uses raw stream buffered {Reader,Writer}.
//
// This test does not consider serialisation nor deserialisation.
func TestConnectionSendRawByte(t *testing.T) {
	const (
		Host = "localhost"
		Port = "12346"
	)
	c := tcp.NewConnectionWithRetry(Host, Port, 10*time.Microsecond)
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func(c tcp.ConnCfg) {
		defer wg.Done()
		server := c.Accept()
		p := make([]byte, 11)
		n, err := server.Read(p)
		if err != nil {
			t.Errorf("receive failed: %v", err)
		}
		t.Logf("Received %d bytes: %s", n, string(p))
		server.Close()
	}(c)

	go func(c tcp.ConnCfg) {
		defer wg.Done()
		client := c.Connect()
		p := []byte("hello world")
		n, err := client.Write(p)
		if err != nil {
			t.Errorf("send failed: %v", err)
		}
		t.Logf("Sent %d bytes: %s", n, string(p))
		client.Close()
	}(c)

	wg.Wait()
}
