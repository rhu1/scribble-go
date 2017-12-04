package shm_test

import (
	"sync"
	"testing"

	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport"
	"github.com/nickng/scribble-go/runtime/transport/shm"
)

func TestDataTypes(t *testing.T) {
	const (
		EptID = 123
	)
	wg := new(sync.WaitGroup)
	wg.Add(2)
	cfg := shm.NewConnection()
	s, c := cfg.Endpoints()

	// channel + machinery for sending/receiving messages
	server := func(s transport.Channel, wg *sync.WaitGroup) {
		var str string
		s.Recv(&str)
		if want, got := "hello", str; want != got {
			t.Errorf("Expecting string %s but got %s", want, got)
		}
		s.Send(len(str))
		var i8r, i8 int8
		s.Recv(&i8r)
		i8 = 2
		if want, got := i8, i8r; want != got {
			t.Errorf("Expecting int8 %d but got %d", want, got)
		}
		var ept *session.Endpoint
		ept = new(session.Endpoint) // ept needs to be pre-allocated
		connBefore := ept.Conn
		s.Recv(&ept)
		connAfter := ept.Conn
		if want, got := EptID, ept.Id; want != got {
			t.Errorf("Expecting *session.Endpoint to be passed but received unmatched: (got %v)", ept)
		}
		if !(connBefore == nil && connAfter != nil) {
			t.Errorf("session.Endpoint received did not get modified: %v vs %v", connBefore, connAfter)
		}

		s.Close() // id in shm (channels do not need to be closed!)
		wg.Done() // Only needed for goroutine synchronisation.
	}
	client := func(c transport.Channel, wg *sync.WaitGroup) {
		c.Send("hello")
		var length int
		c.Recv(&length)
		if want, got := len("hello"), length; want != got {
			t.Errorf("Expecting int %d but got %d", want, got)
		}
		var i8 int8 = 2
		c.Send(i8)
		ept := new(session.Endpoint)
		ept.Id = EptID                                  // For checking that values are passed
		ept.Conn = make(map[string][]transport.Channel) // For checking references are passed
		ept.Conn["test"] = make([]transport.Channel, 0)
		c.Send(ept)

		c.Close() // id in shm (channels do not need to be closed!)
		wg.Done() // Only needed for goroutine synchronisation.
	}
	go server(s, wg)
	go client(c, wg)
	wg.Wait() // Block & wait for {server,client} to finish, emulates end of both endpoints.
	// Output:
	// hello
	// hello with reply
}
