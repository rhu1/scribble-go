package shm_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/nickng/scribble-go/runtime/transport/shm"
)

// This example should be exactly the same as example_test.go in tcp_test, but
// with a different "NewConnection".
func ExampleConn() {
	wg := new(sync.WaitGroup)  // wg is only needed to ensure the 2 goroutines
	wg.Add(2)                  // spawned below {server,client} completes.
	cfg := shm.NewConnection() // in the shared memory case this creates a
	// channel + machinery for sending/receiving messages
	server := func(cfg shm.ConnCfg, wg *sync.WaitGroup) {
		s := cfg.Accept() // Server accepting connection from client: id in shm

		var str string
		s.Recv(&str)
		fmt.Println(str)
		s.Send(str + " with reply")

		s.Close() // id in shm (channels do not need to be closed!)
		wg.Done() // Only needed for goroutine synchronisation.
	}
	client := func(cfg shm.ConnCfg, wg *sync.WaitGroup) {
		time.Sleep(10 * time.Millisecond)
		c := cfg.Connect() // Client connect to server: id in shm

		c.Send("hello")
		var result string
		c.Recv(&result)
		fmt.Println(result)

		c.Close() // id in shm (channels do not need to be closed!)
		wg.Done() // Only needed for goroutine synchronisation.
	}
	go server(cfg, wg)
	go client(cfg, wg)
	wg.Wait() // Block & wait for {server,client} to finish, emulates end of both endpoints.
	// Output:
	// hello
	// hello with reply
}
