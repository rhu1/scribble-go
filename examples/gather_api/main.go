package main

import (
	"fmt"
	"github.com/nickng/scribble-go-runtime/examples/gather_api/gather"
	"log"
	"strconv"
	"sync"
	"time"
)

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(11)

	masterCode := func() {
		masterIni, err := gather.NewMaster(1, 1, 10)
		if err != nil {
			log.Fatalf("cannot create master endpoint: %s", err)
		}
		// One connection for each participant in the group
		for i := 1; i <= 10; i++ {
			err := masterIni.Accept(gather.Worker, i, "127.0.0.1", strconv.Itoa(33333+i))
			if err != nil {
				log.Fatalf("failed to create connection to participant %d of role 'worker': %s", i, err)
			}
		}

		masterMain := mkservmain(10)
		masterIni.Run(masterMain)
		wg.Done()
	}

	go masterCode()

	time.Sleep(1000 * time.Millisecond)

	clientCode := func(i int) {
		clientIni, err := gather.NewWorker(i, 10, 1)
		if err != nil {
			log.Fatalf("cannot create client endpoint: %s", err)
		}
		// One connection for each participant in the group
		err = clientIni.Connect(gather.Master, 1, "127.0.0.1", strconv.Itoa(33333+i))
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

func mkservmain(nw int) func(st1 *gather.Master_1To1_1) *gather.Master_1To1_End {
	return func(st1 *gather.Master_1To1_1) *gather.Master_1To1_End {
		res, ste := st1.RecvAll()
		fmt.Println("Received: ", res)
		return ste
	}
}

func mkworkermain(idx int) func(st1 *gather.Worker_1Ton_1) *gather.Worker_1Ton_End {
	return func(st1 *gather.Worker_1Ton_1) *gather.Worker_1Ton_End {
		return st1.Send(42 + idx)
	}
}
