package tcp_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/nickng/scribble-go/runtime/transport/tcp"
)

// This example shows how to set up a connection, then send and receive message
// on the connection, where the message are user-level variables.
//
// This example uses goroutines to emulate a distributed environment
// sync.WaitGroup ensures the example goroutines complete before the
// parent function.
// In a real distributed set up, sync.WaitGroup is not necessary.
func ExampleConn() {
	wg := new(sync.WaitGroup) // wg is only needed to ensure the 2 goroutines
	wg.Add(2)                 // spawned below {server,client} completes.
	cfg := tcp.NewConnection("127.0.0.1", "22222")
	server := func(cfg tcp.ConnCfg, wg *sync.WaitGroup) {
		s := cfg.Accept() // Server accepting connection from client.

		var str string
		s.Recv(&str)
		fmt.Println(str)
		s.Send(str + " with reply")

		s.Close()
		wg.Done() // Only needed for goroutine synchronisation.
	}
	client := func(cfg tcp.ConnCfg, wg *sync.WaitGroup) {
		time.Sleep(10 * time.Millisecond)
		c := cfg.Connect().(*tcp.Conn) // Client connect to server.

		c.Send("hello")
		var result string
		c.Recv(&result)
		fmt.Println(result)

		c.Close()
		wg.Done() // Only needed for goroutine synchronisation.
	}
	go server(cfg, wg)
	go client(cfg, wg)
	wg.Wait() // Block & wait for {server,client} to finish, emulates end of both endpoints.
	// Output:
	// hello
	// hello with reply
}
