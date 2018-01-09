package scatter

import (
	"fmt"
	"github.com/rhu1/scribble-go-runtime/runtime/session"
	"github.com/rhu1/scribble-go-runtime/runtime/transport"
	"log"
)

const Server = "server"
const Worker = "worker"

type Server_1To1_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func NewServer(id, nserver, nworker int) (*Server_1To1_Init, error) {
	session.RoleRange(id, nserver)
	conn, err := session.NewConn([]session.ParamRole{{Worker, nworker}})

	if err != nil {
		return nil, err
	}

	return &Server_1To1_Init{session.LinearResource{}, session.NewEndpoint(id, nserver, conn)}, nil
}

func (ini *Server_1To1_Init) Ept() *session.Endpoint { return ini.ept }

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
	ini.Ept().ConnMu.RLock()
	for i, conn := range ini.ept.Conn[Worker] {
		if conn == nil { // ini.ept.Conn[Worker][i]
			return nil, fmt.Errorf("invalid connection from 'server[%d]' to 'worker[%d]'", ini.ept.Id, i)
		}
	}
	ini.Ept().ConnMu.RUnlock()

	return &Server_1To1_1{session.LinearResource{}, ini.ept}, nil
}

type Server_1To1_End struct {
}

// Session has started, so if an error occurs, then a runtime error is produced
// and the program exits
func (st1 *Server_1To1_1) SendAll(pl []int) *Server_1To1_1 {
	if len(pl) != len(st1.ept.Conn[Worker]) {
		log.Fatalf("sending wrong number of arguments in 'st1': %d != %d", len(st1.ept.Conn[Worker]), len(pl))
	}
	st1.Use()

	st1.ept.ConnMu.RLock()
	for i, v := range pl {
		st1.ept.Conn[Worker][i].Send(v)
	}
	st1.ept.ConnMu.RUnlock()
	return &Server_1To1_1{session.LinearResource{}, st1.ept}
}

// Convenience to check that user implements the full protocol
func (ini *Server_1To1_Init) Run(f func(*Server_1To1_1) *Server_1To1_End) {
	ini.ept.CheckConnection()
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

func (ini *Worker_1Ton_Init) Ept() *session.Endpoint { return ini.ept }

type Worker_1Ton_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

// Session hasn't started, so an error is returned if anything 'goes wrong'
func (ini *Worker_1Ton_Init) Init() (*Worker_1Ton_1, error) {
	ini.Use()
	ini.ept.ConnMu.Lock()
	defer ini.ept.ConnMu.Unlock()
	for i, conn := range ini.ept.Conn[Server] {
		if conn == nil { // ini.ept.Conn[Server][i]
			return nil, fmt.Errorf("invalid connection from 'worker[%d]' to 'server[%d]'", ini.ept.Id, i)
		}
	}
	return &Worker_1Ton_1{session.LinearResource{}, ini.ept}, nil
}

type Worker_1Ton_End struct {
}

func (st1 *Worker_1Ton_1) RecvAll() ([]int, *Worker_1Ton_1) {
	st1.Use()

	res := make([]int, len(st1.ept.Conn[Server]))
	st1.ept.ConnMu.Lock()
	defer st1.ept.ConnMu.Unlock()
	for i, conn := range st1.ept.Conn[Server] {
		err := conn.Recv(&res[i])
		if err != nil {
			log.Fatalf("wrong value from server at %d: %s", st1.ept.Id, err)
		}
	}
	return res, &Worker_1Ton_1{session.LinearResource{}, st1.ept}
}

func (ini *Worker_1Ton_Init) Run(f func(*Worker_1Ton_1) *Worker_1Ton_End) {
	ini.ept.CheckConnection()
	st1, err := ini.Init()
	if err != nil {
		log.Fatalf("failed to initialise the session: %s", err)
	}
	f(st1)
}
