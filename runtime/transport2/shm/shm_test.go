package shm_test

import (
	"bytes"
	"io"
	"sync"
	"testing"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
)

func TestRoundTrip(t *testing.T) {
	ln, err := shm.Listen(1)
	if err != nil {
		t.Error(err)
	}
	input := []byte("This is the round trip message.")
	go func() {
		server, err := ln.Accept()
		if err != nil {
			t.Error(err)
		}
		r := bytes.NewReader(input)
		inputLen := int64(len(input))
		n, err := io.CopyN(server, r, inputLen) // message to server
		if err != nil {
			t.Error(err)
		}
		t.Logf("Server received %d bytes", n)
		n, err = io.CopyN(server, server, inputLen) // server to client
		if err != nil {
			t.Error(err)
		}
		t.Logf("Server forwarded %d bytes", n)
	}()
	client, err := shm.Dial("", 1)
	if err != nil {
		t.Error(err)
	}

	inputLen := int64(len(input))
	n, err := io.CopyN(client, client, inputLen)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Client forwarded %d bytes", n)
	var b bytes.Buffer
	n, err = io.CopyN(&b, client, inputLen)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Client received %d bytes: %s", n, b.String())
	if want, got := string(input), b.String(); want != got {
		t.Errorf("Expected %s but got %s", want, got)
	}
	if err := ln.Close(); err != nil {
		t.Error(err)
	}
}

// msg is an implementation of session message.
type msg struct {
	V int
}

func (*msg) GetOp() string { return "msg" }

func TestRoundTripPointer(t *testing.T) {
	ln, err := shm.Listen(1)
	if err != nil {
		t.Error(err)
	}
	go func() {
		client, err := shm.Dial("", 1)
		if err != nil {
			t.Error(err)
		}
		var clientMsg interface{} = &msg{}
		client.(session2.PointerReader).ReadPointer(&clientMsg)
		clientMsg.(*msg).V++
		client.(session2.PointerWriter).WritePointer(clientMsg)
	}()

	server, err := ln.Accept()
	if err != nil {
		t.Error(err)
	}
	var serverMsg interface{} = &msg{V: 41}
	server.(session2.PointerWriter).WritePointer(serverMsg)
	server.(session2.PointerReader).ReadPointer(&serverMsg)
	if want, got := 42, serverMsg.(*msg).V; want != got {
		t.Errorf("Expecting %d but got %d", want, got)
	}
	t.Log(serverMsg.(*msg).V)
	if err := ln.Close(); err != nil {
		t.Error(err)
	}
}

func TestSynchronousConnection(t *testing.T) {
	port := 10
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func(port int, wg *sync.WaitGroup) {
		ch, err := shm.Dial("", port)
		if err != nil {
			t.Fatalf("shm: dial to :%d failed: %v", port, err)
		}
		defer func() {
			if err := ch.Close(); err != nil {
				t.Errorf("shm: close failed: %v", err)
			}
		}()
		wg.Done()
	}(port, wg)
	go func(port int, wg *sync.WaitGroup) {
		ln, err := shm.Listen(port)
		if err != nil {
			t.Fatalf("cannot create listener: %v", err)
		}
		ch, err := ln.Accept()
		if err != nil {
			t.Fatalf("shm: cannot accept: %v", err)
		}
		defer func() {
			if err := ln.Close(); err != nil {
				t.Errorf("shm: listener close failed: %v", err)
			}
		}()
		defer func() {
			if err := ch.Close(); err != nil {
				t.Errorf("shm: close failed: %v", err)
			}
		}()
		wg.Done()
	}(port, wg)
	wg.Wait()
}
