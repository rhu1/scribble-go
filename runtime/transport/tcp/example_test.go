package tcp_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/nickng/scribble-go/runtime/transport/tcp"
)

// This example shows how to set up a connection, then send and receive message
// on the connection, where the message are user-level variables.
func ExampleConn() {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	cfg := tcp.NewConnection("127.0.0.1", "22222")
	server := func(cfg tcp.ConnCfg, wg *sync.WaitGroup) {
		s := cfg.Accept() // Server accepting connection from client.

		var str string
		s.Recv(&str)
		fmt.Println(str)
		s.Send(str + " with reply")

		s.Close()
		wg.Done()
	}
	client := func(cfg tcp.ConnCfg, wg *sync.WaitGroup) {
		time.Sleep(10 * time.Millisecond)
		c := cfg.Connect().(*tcp.Conn) // Client connect to server.

		c.Send("hello")
		var result string
		c.Recv(&result)
		fmt.Println(result)

		c.Close()
		wg.Done()
	}
	go server(cfg, wg)
	go client(cfg, wg)
	wg.Wait()
	// Output:
	// hello
	// hello with reply
}
