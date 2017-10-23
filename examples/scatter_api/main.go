package main

import (
	"./scatter"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(11)

	serverCode := func() {
		serverEpt, err := scatter.NewServer(1, 1, 10)
		if err != nil {
			log.Fatalf("cannot create server endpoint: %s", err)
		}
		// One connection for each participant in the group
		for i := 1; i <= 10; i++ {
			err := serverEpt.ConnectionToW(i, "127.0.0.1", strconv.Itoa(33333+i))
			if err != nil {
				log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
			}
		}

		time.Sleep(150 * time.Millisecond)

		serverMain := mkservmain(10)
		serverEpt.Run(serverMain)
		wg.Done()
	}

	go serverCode()

	time.Sleep(100 * time.Millisecond)

	clientCode := func(i int) {
		clientEpt, err := scatter.NewWorker(i, 10, 1)
		if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}
		// One connection for each participant in the group
		err = clientEpt.ConnectionToS(1, "127.0.0.1", strconv.Itoa(33333+i))
		if err != nil {
			log.Fatalf("failed to create connection from participant %d of role 'worker': %s", i, err)
		}

		clientMain := mkworkermain(i)
		clientEpt.Run(clientMain)
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
		return st1.SendAll(payload)
	}
}

func mkworkermain(idx int) func(st1 *scatter.Worker_1To1_1) *scatter.Worker_1To1_End {
	return func(st1 *scatter.Worker_1To1_1) *scatter.Worker_1To1_End {
		r, st := st1.RecvAll()
		fmt.Println("Received payload at worker ", idx, "\t: ", r[0])
		return st
	}
}
