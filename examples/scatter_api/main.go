package main

import (
	"./scatter"
	"log"
	"strconv"
)

func main() {

	serverCode := func() {
		serverEpt, err := scatter.NewServer(1, 1, 10)
		if err != nil {
			log.Fatalf("cannot create server endpoint: %s", err)
		}
		// One connection for each participant in the group
		for i := 1; i <= 10; i++ {
			err := serverEpt.ConnectionToW(i, "127.0.0.1", strconv.Itoa(3333+i))
			if err != nil {
				log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
			}
		}

		serverMain := mkservmain(10)
		serverEpt.Run(serverMain)
	}

	go serverCode()

	clientCode := func(i int) {
		clientEpt, err := scatter.NewWorker(1, 10, 1)
		if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}
		// One connection for each participant in the group
		err = clientEpt.ConnectionToS(1, "127.0.0.1", strconv.Itoa(3333+i))
		if err != nil {
			log.Fatalf("failed to create connection from participant %d of role 'worker': %s", i, err)
		}

		clientMain := mkworkermain()
		clientEpt.Run(clientMain)
	}

	for i := 1; i <= 10; i++ {
		go clientCode(i)
	}
}

func mkservmain(nw int) func(st1 *scatter.Server_1To1_1) *scatter.Server_1To1_End {
	payload := make([]int, nw)
	for i := 0; i < nw; i++ {
		payload[i] = 42
	}
	return func(st1 *scatter.Server_1To1_1) *scatter.Server_1To1_End {
		return st1.SendAll(payload)
	}
}

func mkworkermain() func(st1 *scatter.Worker_1To1_1) *scatter.Worker_1To1_End {
	return func(st1 *scatter.Worker_1To1_1) *scatter.Worker_1To1_End {
		_, st := st1.RecvAll()
		return st
	}
}
