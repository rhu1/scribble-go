package shm_test

import (
	"fmt"
	"sync"

	"github.com/nickng/scribble-go/runtime/transport"
	"github.com/nickng/scribble-go/runtime/transport/shm"
)

// This example should be exactly the same as example_test.go in tcp_test, but
// with a different "NewConnection".
func ExampleShm() {
	wg := new(sync.WaitGroup)  // wg is only needed to ensure the 2 goroutines
	wg.Add(2)                  // spawned below {server,client} completes.
	cfg := shm.NewConnection() // in the shared memory case this creates a
	s, c := cfg.Endpoints()    // Setup two ends of the shared memory connection.

	// Machinery for sending/receiving messages
	server := func(s transport.Channel, wg *sync.WaitGroup) {
		var str string
		s.Recv(&str)
		fmt.Println(str)
		s.Send(str + " with reply")

		s.Close() // id in shm (channels do not need to be closed!)
		wg.Done() // Only needed for goroutine synchronisation.
	}
	client := func(c transport.Channel, wg *sync.WaitGroup) {
		c.Send("hello")
		var result string
		c.Recv(&result)
		fmt.Println(result)

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
