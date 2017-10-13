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

// This tests sending and receiving multiple byte streams over TCP.
// The test uses delimited {Reader,Writer} and since they are not exported in
// Conn, external Delimited {Reader,Writer} are created and used for sending and
// receiving.
//
// This test does not consider serialisation nor deserialisation.
func TestConnectionSendMultiBytes(t *testing.T) {
	const (
		Host = "127.0.0.1"
		Port = "12347"
		N    = 10
	)
	c := tcp.NewConnectionWithRetry(Host, Port, 10*time.Microsecond)
	wg := new(sync.WaitGroup)
	wg.Add(2)

	var sSent, sRcvd int
	var cSent, cRcvd int

	go func(c tcp.ConnCfg) {
		defer wg.Done()
		server := c.Accept()
		p1 := make([]byte, 256)
		p2 := make([]byte, 256)
		r := tcp.NewDelimReader(server.(*tcp.Conn), c.DelimMeth)
		w := tcp.NewDelimWriter(server.(*tcp.Conn), c.DelimMeth)
		for i := 0; i < N; i++ {
			n1, err := r.Read(p1)
			if err != nil {
				t.Errorf("receive failed: %v", err)
			}
			sRcvd += n1
			n2, err := r.Read(p2)
			if err != nil {
				t.Errorf("receive failed: %v", err)
			}
			sRcvd += n2
			t.Logf("S: Rcvd[%d] %v + %v", i, p1[:n1], p2[:n2])
			n1, err = w.Write(p1[:n1])
			if err != nil {
				t.Errorf("send failed: %v", err)
			}
			sSent += n1
			n2, err = w.Write(p2[:n2])
			if err != nil {
				t.Errorf("send failed: %v", err)
			}
			sSent += n2
		}
		t.Logf("Sent/Received %d/%d bytes", sSent, sRcvd)
		server.Close()
	}(c)

	go func(c tcp.ConnCfg) {
		defer wg.Done()
		client := c.Connect()
		b1 := []byte("hello ")
		b2 := []byte("world?")
		var p1, p2 []byte
		var n1, n2 int
		var err error
		r := tcp.NewDelimReader(client.(*tcp.Conn), c.DelimMeth)
		w := tcp.NewDelimWriter(client.(*tcp.Conn), c.DelimMeth)
		for i := 0; i < N; i++ {
			if i == 0 {
				n1, err = w.Write(b1)
				if err != nil {
					t.Errorf("send failed: %v", err)
				}
				cSent += n1
				p1 = make([]byte, 256)
				n2, err = w.Write(b2)
				if err != nil {
					t.Errorf("send failed: %v", err)
				}
				cSent += n2
				p2 = make([]byte, 256)
			} else {
				n1, err = w.Write(p1[:n1])
				if err != nil {
					t.Errorf("send failed: %v", err)
				}
				cSent += n1
				n1, err = w.Write(p2[:n2])
				if err != nil {
					t.Errorf("send failed: %v", err)
				}
				cSent += n2
			}
			n1, err = r.Read(p1)
			if err != nil {
				t.Errorf("receive failed: %v", err)
			}
			cRcvd += n1
			n2, err = r.Read(p2)
			if err != nil {
				t.Errorf("receive failed: %v", err)
			}
			cRcvd += n2
			t.Logf("C: Rcvd[%d] %v + %v", i, p1[:n1], p2[:n2])
		}
		t.Logf("Sent/Received %d/%d bytes", cSent, cRcvd)
		client.Close()
	}(c)

	wg.Wait()
	if sSent != cRcvd {
		t.Errorf("data transfer mismatch: S sent %d, C received %d", sSent, cRcvd)
	}
	if cSent != sRcvd {
		t.Errorf("data transfer mismatch: C sent %d, S received %d", cSent, sRcvd)
	}
}
