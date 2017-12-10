package gather

import (
	"fmt"
	"github.com/nickng/scribble-go-runtime/runtime/session"
	"github.com/nickng/scribble-go-runtime/runtime/transport"
	"github.com/nickng/scribble-go-runtime/runtime/transport/tcp"
	"log"
)

const Master = "master"
const Worker = "worker"

type Master_1To1_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func NewMaster(id, nmaster, nworker int) (*Master_1To1_Init, error) {
	if id > nmaster || id < 1 {
		return nil, fmt.Errorf("'master' id not in range [1, %d]", nmaster)
	}
	if nworker < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'worker': %d", nworker)
	}
	conn := make(map[string][]transport.Channel)
	conn[Worker] = make([]transport.Channel, nworker)

	return &Master_1To1_Init{ept: session.NewEndpoint(id, nmaster, conn)}, nil
}

func (ini *Master_1To1_Init) Accept(rolename string, id int, addr, port string) error {
	cn, ok := ini.ept.Conn[rolename]
	if !ok {
		return fmt.Errorf("rolename '%s' does not exist", rolename)
	}
	if id < 1 || id > len(cn) {
		return fmt.Errorf("participant %d of role '%s' out of bounds", id, rolename)
	}
	go func(i int, addr, port string) {
		ini.ept.Conn[rolename][i-1] = tcp.NewConnection(addr, port).Accept().(*tcp.Conn)
	}(id, addr, port)
	return nil
}

type Master_1To1_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

// Session hasn't started, so an error is returned if anything 'goes wrong'
// For the master, wait until a connection for each participant is available.
// FIXME: inefficient spinlock.
// TODO: Assumption, rolename and
func (ini *Master_1To1_Init) Init() (*Master_1To1_1, error) {
	ini.Use()
	conn := ini.ept.Conn[Worker]
	n_worker := len(conn)

	// FIXME
	for i := 0; i < n_worker; i++ {
		for ini.ept.Conn[Worker][i] == nil {
		}
	}

	return &Master_1To1_1{session.LinearResource{}, ini.ept}, nil
}

type Master_1To1_End struct {
}

// Session has started, so if an error occurs, then a runtime error is produced
// and the program exits
func (st1 *Master_1To1_1) RecvAll() ([]int, *Master_1To1_End) {
	st1.Use()

	var tmp int
	res := make([]int, len(st1.ept.Conn[Worker]))
	for i, conn := range st1.ept.Conn[Worker] {
		err := conn.Recv(&tmp)
		if err != nil {
			log.Fatalf("wrong value from master at %d: %s", st1.ept.Id, err)
		}
		res[i] = tmp
	}
	return res, &Master_1To1_End{}
}

// Convenience to check that user implements the full protocol
func (ini *Master_1To1_Init) Run(f func(*Master_1To1_1) *Master_1To1_End) {

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

func NewWorker(id, nworker, nmaster int) (*Worker_1Ton_Init, error) {
	if id > nworker || id < 1 {
		return nil, fmt.Errorf("'worker' id not in range [1, %d]", nworker)
	}
	if nmaster < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'master': %d", nmaster)
	}
	conn := make(map[string][]transport.Channel)
	conn[Master] = make([]transport.Channel, nmaster)

	return &Worker_1Ton_Init{session.LinearResource{}, session.NewEndpoint(id, nworker, conn)}, nil
}

func (ini *Worker_1Ton_Init) Connect(rolename string, id int, addr, port string) error {
	cn, ok := ini.ept.Conn[rolename]
	if !ok {
		return fmt.Errorf("rolename '%s' does not exist", rolename)
	}
	if id < 1 || id > len(cn) {
		return fmt.Errorf("participant %d of role '%s' out of bounds", id, rolename)
	}
	// Probably a good idea to use tcp.NewConnectionWithRetry
	ini.ept.Conn[rolename][id-1] = tcp.NewConnection(addr, port).Connect()
	return nil
}

type Worker_1Ton_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

// Session hasn't started, so an error is returned if anything 'goes wrong'
func (ini *Worker_1Ton_Init) Init() (*Worker_1Ton_1, error) {
	n_master := len(ini.ept.Conn[Master])
	for i := 0; i < n_master; i++ {
		if ini.ept.Conn[Master][i] == nil {
			return nil, fmt.Errorf("invalid connection from 'worker[%d]' to 'master[%d]'", ini.ept.Id, i)
		}
	}
	return &Worker_1Ton_1{session.LinearResource{}, ini.ept}, nil
}

type Worker_1Ton_End struct {
}

func (st1 *Worker_1Ton_1) Send(v int) *Worker_1Ton_End {
	st1.Use()
	st1.ept.Conn[Master][0].Send(v)
	return &Worker_1Ton_End{}
}

func (ini *Worker_1Ton_Init) Run(f func(*Worker_1Ton_1) *Worker_1Ton_End) {
	st1, err := ini.Init()
	if err != nil {
		log.Fatalf("failed to initialise the session: %s", err)
	}
	f(st1)
}
