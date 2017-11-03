package scatter

/***
* Protocol:

global protocol scatter(role Server(n), role Worker(n)){
  choice at Server[1..1] {
	scatter(int) from Server[1..1] to Worker[1..n];
	do scatter(Server, Worker);
  } or {
	quit() from Server[1..1] to Worker[1..n];
  }
}

*/

import (
	"fmt"
	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport"
	"log"
)

const Server = "server"
const Worker = "worker"

const LScatter = 1 // Or: "scatter"
const LQuit = 2    // or: "quit"

type Server_1To1_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *Server_1To1_Init) Ept() *session.Endpoint {
	return ini.ept
}

func NewServer(id, nserver, nworker int) (*Server_1To1_Init, error) {
	if id > nserver || id < 1 {
		return nil, fmt.Errorf("'server' id not in range [1, %d]", nserver)
	}
	if nworker < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'worker': %d", nworker)
	}
	conn := make(map[string][]transport.Channel)
	conn[Worker] = make([]transport.Channel, nworker)

	return &Server_1To1_Init{session.LinearResource{}, &session.Endpoint{id, nserver, conn}}, nil
}

type Server_1To1_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

// Session hasn't started, so an error is returned if anything 'goes wrong'
// For the server, wait until a connection for each participant is available.
// FIXME: inefficient spinlock.
// TODO: Assumption, rolename and
func (ini *Server_1To1_Init) Init() (*Server_1To1_1, error) {
	ini.Use()
	conn := ini.ept.Conn[Worker]
	n_worker := len(conn)

	// FIXME
	for i := 0; i < n_worker; i++ {
		for ini.ept.Conn[Worker][i] == nil {
		}
	}

	return &Server_1To1_1{session.LinearResource{}, ini.ept}, nil
}

func (st1 *Server_1To1_1) Scatter(pl []int) *Server_1To1_1 {
	if len(pl) != len(st1.ept.Conn[Worker]) {
		log.Fatalf("sending wrong number of arguments in 'st1': %d != %d", len(st1.ept.Conn[Worker]), len(pl))
	}
	st1.Use()

	for i, v := range pl {
		st1.ept.Conn[Worker][i].Send(LScatter)
		st1.ept.Conn[Worker][i].Send(v)
	}
	return &Server_1To1_1{session.LinearResource{}, st1.ept}
}

type Server_1To1_End struct {
}

func (st1 *Server_1To1_1) Quit() *Server_1To1_End {
	st1.Use()

	for _, v := range st1.ept.Conn[Worker] {
		v.Send(LQuit)
	}
	return &Server_1To1_End{}
}

// Convenience to check that user implements the full protocol
func (ini *Server_1To1_Init) Run(f func(*Server_1To1_1) *Server_1To1_End) {

	st1, err := ini.Init()

	if err != nil {
		log.Fatalf("failed to initialise the session: %s", err)
	}

	f(st1)
}

type Worker_1Ton_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *Worker_1Ton_Init) Ept() *session.Endpoint {
	return ini.ept
}

func NewWorker(id, nworker, nserver int) (*Worker_1Ton_Init, error) {
	if id > nworker || id < 1 {
		return nil, fmt.Errorf("'worker' id not in range [1, %d]", nworker)
	}
	if nserver < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'server': %d", nserver)
	}
	conn := make(map[string][]transport.Channel)
	conn[Server] = make([]transport.Channel, nserver)

	return &Worker_1Ton_Init{session.LinearResource{}, session.NewEndpoint(id, nworker, conn)}, nil
}

type Worker_1Ton_1 struct {
	session.LinearResource
	ept        *session.Endpoint
	data       chan int
	ch_scatter chan chan *Worker_1Ton_1
	ch_quit    chan chan *Worker_1Ton_End
}

func newWorker_1Ton_1(ept *session.Endpoint) *Worker_1Ton_1 {
	ch_res := make(chan int, 1)
	ch_st1 := make(chan chan *Worker_1Ton_1, 1)
	ch_st2 := make(chan chan *Worker_1Ton_End, 1)
	st1 := &Worker_1Ton_1{session.LinearResource{}, ept, ch_res, ch_st1, ch_st2}
	go st1.scatterOrQuit(ch_res, ch_st1, ch_st2)
	return st1
}

// Session hasn't started, so an error is returned if anything 'goes wrong'
func (ini *Worker_1Ton_Init) Init() (*Worker_1Ton_1, error) {
	ini.Use()
	n_server := len(ini.ept.Conn[Server])
	for i := 0; i < n_server; i++ {
		if ini.ept.Conn[Server][i] == nil {
			return nil, fmt.Errorf("invalid connection from 'worker[%d]' to 'server[%d]'", ini.ept.Id, i)
		}
	}
	return newWorker_1Ton_1(ini.ept), nil
}

type Worker_1Ton_End struct {
}

func (st1 *Worker_1Ton_1) Scatter(res *int) <-chan *Worker_1Ton_1 {
	ch, selected := <-st1.ch_scatter
	if !selected {
		return nil
	}
	*res = <-st1.data
	return ch
}

func (st1 *Worker_1Ton_1) Quit() <-chan *Worker_1Ton_End {
	ch, selected := <-st1.ch_quit
	if !selected {
		return nil
	}
	return ch
}

func (st1 *Worker_1Ton_1) scatterOrQuit(data chan int, st2 chan chan *Worker_1Ton_1, st3 chan chan *Worker_1Ton_End) {
	st1.Use()
	var lbl int
	var res int

	conn := st1.ept.Conn[Server][0]

	err := conn.Recv(&lbl)
	if err != nil {
		log.Fatalf("wrong value from server at %d: %s", st1.ept.Id, err)
	}

	if lbl == LScatter {
		ch := make(chan *Worker_1Ton_1, 1)
		err = conn.Recv(&res)
		if err != nil {
			log.Fatalf("wrong value from server at %d: %s", st1.ept.Id, err)
		}
		data <- res
		ch <- newWorker_1Ton_1(st1.ept)
		st2 <- ch
		close(st3)
		return
	}
	if lbl == LQuit {
		ch := make(chan *Worker_1Ton_End, 1)
		ch <- &Worker_1Ton_End{}
		st3 <- ch
		close(st2)
		return
	}
	log.Fatalf("wrong value from server at %d: %s", st1.ept.Id, err)
}

func (ini *Worker_1Ton_Init) Run(f func(*Worker_1Ton_1) *Worker_1Ton_End) {
	st1, err := ini.Init()
	if err != nil {
		log.Fatalf("failed to initialise the session: %s", err)
	}
	f(st1)
}
