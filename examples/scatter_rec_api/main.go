package main

import (
	"fmt"
	"github.com/nickng/scribble-go/examples/scatter_rec_api/scatter"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport/tcp"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	wg := new(sync.WaitGroup)
	wg.Add(11)

	serverCode := func() {
		serverIni, err := scatter.NewServer(1, 1, 10)
		if err != nil {
			log.Fatalf("cannot create server endpoint: %s", err)
		}
		// One connection for each participant in the group
		for i := 1; i <= 10; i++ {
			conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
			err := session.Accept(serverIni, scatter.Worker, i, conn)
			if err != nil {
				log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
			}
		}

		serverMain := mkservmain(10)
		serverIni.Run(serverMain)
		wg.Done()
	}

	go serverCode()

	time.Sleep(100 * time.Millisecond)

	clientCode := func(i int) {
		clientIni, err := scatter.NewWorker(i, 10, 1)
		if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}
		// One connection for each participant in the group
		conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
		err = session.Connect(clientIni, scatter.Server, 1, conn)
		if err != nil {
			log.Fatalf("failed to create connection from participant %d of role 'worker': %s", i, err)
		}

		clientMain := mkworkermain(i)
		clientIni.Run(clientMain)
		wg.Done()
	}

	for i := 1; i <= 10; i++ {
		go clientCode(i)
	}

	wg.Wait()
}

func mkservmain(nw int) func(st1 *scatter.Server_1To1_1) *scatter.Server_1To1_End {
	payload := make([]int, nw)
	for i := 0; i < nw; i++ {
		payload[i] = 42 + i
	}
	return func(st1 *scatter.Server_1To1_1) *scatter.Server_1To1_End {
		for i := 0; i < 100000; i++ {
			st1 = st1.Scatter(payload)
		}
		return st1.Quit()
	}
}

func mkworkermain(idx int) func(st1 *scatter.Worker_1Ton_1) *scatter.Worker_1Ton_End {
	return func(st1 *scatter.Worker_1Ton_1) *scatter.Worker_1Ton_End {
		var st2 *scatter.Worker_1Ton_End
		var pl int
		for {
			select {
			case st1 = <-st1.Scatter(&pl):
				fmt.Fprintf(os.Stdout, "Received payload at worker %d\t:\t%d\n", idx, pl)
			case st2 = <-st1.Quit():
				return st2
			}
		}
		return st2
	}
}
