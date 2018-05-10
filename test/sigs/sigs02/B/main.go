//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/sigs/sigs02B
//$ bin/B.exe

//go:generate scribblec-param.sh ../OneToMany.scr -d ../ -param Scatter -param-api Scatter B

package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/rhu1/scribble-go-runtime/test/sigs/sigs02/messages"
	"github.com/rhu1/scribble-go-runtime/test/sigs/sigs02/OneToMany/Scatter"
	B_1K "github.com/rhu1/scribble-go-runtime/test/sigs/sigs02/OneToMany/Scatter/B_1toK"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/tcp"
)

const (
	k = 2
)

func init() {
	var data messages.Data
	gob.Register(data)
}

func main() {
	s := Scatter.New()
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go b(s, 1, wg)
	go b(s, 2, wg)
	wg.Wait()
}

func b(s *Scatter.Scatter, id int, wg *sync.WaitGroup) {
	ln, err := tcp.Listen(3333 + id - 1)
	if err != nil {
		log.Fatalf("Cannot listen: %v", err)
	}
	B := s.New_B_1toK(k, id)
	if err := B.A_1to1_Accept(1, ln, new(session2.GobFormatter)); err != nil {
		log.Fatal(err)
	}
	B.Run(func(s *B_1K.Init_4) B_1K.End {
		d := make([]messages.Data, 1)
		end := s.A_1to1_Gather_Data(d)
		fmt.Println("B(" + strconv.Itoa(s.Ept.Self) + ") gathered:", d)
		return *end
	})
	wg.Done()
}
