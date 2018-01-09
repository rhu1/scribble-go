package httpget

import (
	"fmt"
	"log"

	"github.com/rhu1/scribble-go-runtime/runtime/session"
	"github.com/rhu1/scribble-go-runtime/runtime/transport"
)

const (
	Fetcher = "Fetcher"
	Master  = "Master"
	Server  = "Server"
)

func NewFetcher(id, nFetcher, nMaster, nServer int) (*Fetcher_Init, error) {
	if id > nFetcher || id < 1 {
		return nil, fmt.Errorf("'Fetcher' ID not in range [1, %d]", nFetcher)
	}
	if nMaster < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'Master': %d", nMaster)
	}
	if nServer < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'Server': %d", nServer)
	}
	conn := make(map[string][]transport.Channel)
	conn[Master] = make([]transport.Channel, nMaster)
	conn[Server] = make([]transport.Channel, nServer)

	return &Fetcher_Init{ept: session.NewEndpoint(id, nFetcher, conn)}, nil
}

type Fetcher_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Fetcher_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *Fetcher_Init) Init() (*Fetcher_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for i, conn := range st.ept.Conn[Master] {
		if conn == nil {
			return nil, fmt.Errorf("invalid connection Fetcher[%d] ↔ Master[%d]", st.ept.Id, i)
		}
	}
	st.ept.ConnMu.Unlock()

	return &Fetcher_1{ept: st.ept}, nil
}

func (st *Fetcher_Init) Run(fn func(*Fetcher_1) *Master_End) {
	st.ept.CheckConnection()

	st1, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st1)
}

type Fetcher_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func NewMaster(id, nFetcher, nMaster, nServer int) (*Master_Init, error) {
	if nFetcher < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'Fetcher': %d", nFetcher)
	}
	if id > nMaster || id < 1 {
		return nil, fmt.Errorf("'Master' ID not in range [1, %d]", nMaster)
	}
	if nServer < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'Server': %d", nServer)
	}
	conn := make(map[string][]transport.Channel)
	conn[Fetcher] = make([]transport.Channel, nFetcher)
	conn[Server] = make([]transport.Channel, nServer)

	return &Master_Init{ept: session.NewEndpoint(id, nMaster, conn)}, nil
}

type Master_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Master_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *Master_Init) Init() (*Master_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	defer st.ept.ConnMu.Unlock()
	for i, conn := range st.ept.Conn[Fetcher] {
		if conn == nil {
			return nil, fmt.Errorf("invalid connection from Master[%d] to %s[%d]", st.ept.Id, Fetcher, i)
		}
	}
	return &Master_1{ept: st.ept}, nil
}

func (st *Master_Init) Run(fn func(*Master_1) *Master_End) {
	st.ept.CheckConnection()

	st1, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st1)
}

type Master_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Master_1) SendAll_Fetcher_URL(args []string) *Master_2 {
	st.Use()

	st.ept.ConnMu.RLock()
	for i, arg := range args {
		st.ept.Conn[Fetcher][i].Send(arg)
	}
	st.ept.ConnMu.RUnlock()
	return &Master_2{ept: st.ept}
}

type Master_2 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Master_2) RecvAll_Fetcher_FileSize() ([]int, *Master_3) {
	st.Use()

	res := make([]int, len(st.ept.Conn[Fetcher]))
	st.ept.ConnMu.RLock()
	for i, conn := range st.ept.Conn[Fetcher] {
		err := conn.Recv(&res[i])
		if err != nil {
			log.Fatalf("Wrong value from %s[%d] at Master[%d]: %v", Fetcher, i, st.ept.Id, err)
		}
	}
	st.ept.ConnMu.RUnlock()
	return res, &Master_3{ept: st.ept}
}

type Master_3 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Master_3) SendAll_Fetcher_start(args []int) *Master_4 {
	st.Use()

	st.ept.ConnMu.RLock()
	for i, arg := range args {
		st.ept.Conn[Fetcher][i].Send(arg)
	}
	st.ept.ConnMu.RUnlock()
	return &Master_4{ept: st.ept}
}

type Master_4 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Master_4) SendAll_Fetcher_end(args []int) *Master_5 {
	st.Use()

	st.ept.ConnMu.RLock()
	for i, arg := range args {
		st.ept.Conn[Fetcher][i].Send(arg)
	}
	st.ept.ConnMu.RUnlock()
	return &Master_5{ept: st.ept}
}

type Master_5 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Master_5) RecvAll_Fetcher_merge() ([]string, *Master_End) {
	st.Use()

	res := make([]string, len(st.ept.Conn[Fetcher]))
	st.ept.ConnMu.RLock()
	for i, conn := range st.ept.Conn[Fetcher] {
		err := conn.Recv(&res[i])
		if err != nil {
			log.Fatalf("Wrong value from %s[%d] at Master[%d]: %v", Fetcher, i, st.ept.Id, err)
		}
	}
	st.ept.ConnMu.RUnlock()
	return res, &Master_End{}
}

type Master_End struct {
}

func NewServer(id, nFetcher, nMaster, nServer int) (*Server_Init, error) {
	if nFetcher < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'Fetcher': %d", nFetcher)
	}
	if nMaster < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'Master': %d", nMaster)
	}
	if id > nServer || id < 1 {
		return nil, fmt.Errorf("'Server' ID not in range [1, %d]", nServer)
	}
	conn := make(map[string][]transport.Channel)
	conn[Fetcher] = make([]transport.Channel, nFetcher)
	conn[Master] = make([]transport.Channel, nMaster)

	return &Server_Init{ept: session.NewEndpoint(id, nServer, conn)}, nil
}

type Server_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

/* Server implementation ignored */
