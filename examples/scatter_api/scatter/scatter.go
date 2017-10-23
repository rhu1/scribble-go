package scatter

import (
	"fmt"
	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport/tcp"
	"log"
)

type ServerEpt struct {
	roleId   int
	numRoles int

	n_worker    int
	conn_worker []*tcp.Conn
}

func NewServer(id, nserver, nworker int) (*ServerEpt, error) {
	if id > nserver || id < 1 {
		return nil, fmt.Errorf("Server id not in range [1, %d]", nserver)
	}
	if nworker < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'worker': %d", nworker)
	}
	return (&ServerEpt{id, nserver, nworker, make([]*tcp.Conn, nworker)}), nil
}

func (ept *ServerEpt) ConnectionToW(i int, addr, port string) error {
	if i < 1 || i > ept.n_worker {
		return fmt.Errorf("participant %d of role 'worker' out of bounds", i)
	}
	go func(i int, addr, port string) {
		ept.conn_worker[i-1] = tcp.NewConnection(addr, port).Accept().(*tcp.Conn)
	}(i, addr, port)
	return nil
}

type Server_1To1_1 struct {
	session.LinearResource
	ept *ServerEpt
}

// Session hasn't started, so an error is returned if anything 'goes wrong'
func (ept *ServerEpt) Init() (*Server_1To1_1, error) {
	for i := 0; i < ept.n_worker; i++ {
		if ept.conn_worker[i] == nil {
			return nil, fmt.Errorf("invalid connection from 'server' to 'worker' participant %d", i)
		}
	}
	return &Server_1To1_1{session.LinearResource{}, ept}, nil
}

type Server_1To1_End struct {
}

// Session has started, so if an error occurs, then a runtime error is produced
// and the program exits
func (st1 *Server_1To1_1) SendAll(pl []int) *Server_1To1_End {
	if len(pl) != st1.ept.n_worker {
		log.Fatalf("error, sending wrong number of arguments in 'st1': %d != %d", st1.ept.n_worker, len(pl))
	}
	st1.Use()

	for i, v := range pl {
		st1.ept.conn_worker[i].Send(v)
	}
	return &Server_1To1_End{}
}

// Convenience to check that user implements the full protocol
func (ept *ServerEpt) Run(f func(*Server_1To1_1) *Server_1To1_End) {

	st1, err := ept.Init()

	if err != nil {
		log.Fatalf("failed to initialise the session: %s", err)
	}

	f(st1)
}

// One way of generalising this could be describing an interface for endpoints
// instead of structs?
type WorkerEpt struct {
	roleId   int
	numRoles int

	n_server    int
	conn_server []*tcp.Conn
}

func NewWorker(id, nworker, nserver int) (*WorkerEpt, error) {
	if id > nworker || id < 1 {
		return nil, fmt.Errorf("Worker id not in range [1, %d]", nworker)
	}
	if nserver < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'server': %d", nserver)
	}
	return (&WorkerEpt{id, nworker, nserver, make([]*tcp.Conn, nserver)}), nil
}

func (ept *WorkerEpt) ConnectionToS(i int, addr, port string) error {
	if i < 1 || i > ept.n_server {
		return fmt.Errorf("participant %d of role 'server' out of bounds", i)
	}
	// Difference with ConnectionToW is in the use of 'Connect' instead of
	// 'Accept'. If we want to generalise this, we'd need to sort this out.
	// Also, probably a good idea to use tcp.NewConnectionWithRetry
	ept.conn_server[i-1] = tcp.NewConnection(addr, port).Connect().(*tcp.Conn)
	return nil
}

type Worker_1To1_1 struct {
	session.LinearResource
	ept *WorkerEpt
}

// Session hasn't started, so an error is returned if anything 'goes wrong'
func (ept *WorkerEpt) Init() (*Worker_1To1_1, error) {
	for i := 0; i < ept.n_server; i++ {
		if ept.conn_server[i] == nil {
			return nil, fmt.Errorf("invalid connection from 'worker' %d to 'server' participant %d", ept.roleId, i)
		}
	}
	return &Worker_1To1_1{session.LinearResource{}, ept}, nil
}

type Worker_1To1_End struct {
}

func (st1 *Worker_1To1_1) RecvAll() ([]int, *Worker_1To1_End) {
	var tmp int
	st1.Use()

	res := make([]int, st1.ept.n_server)
	for i, conn := range st1.ept.conn_server {
		err := conn.Recv(&tmp)
		if err != nil {
			log.Fatalf("wrong value from server at %d: %s", st1.ept.roleId, err)
		}
		res[i] = tmp
	}
	return res, &Worker_1To1_End{}
}

func (ept *WorkerEpt) Run(f func(*Worker_1To1_1) *Worker_1To1_End) {
	st1, err := ept.Init()
	if err != nil {
		log.Fatalf("failed to initialise the session: %s", err)
	}
	f(st1)
}
