package Game

import (
	"fmt"
	"log"

	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport"
)

const (
	ClientA_q = "ClientA q"
	ClientA_p = "ClientA p"
)

// global protocol ClientA, role q
func NewClientA_q() (*ClientA_q_Init, error) {
	// Omitted range check.

	conn := make(map[string][]transport.Channel)
	conn[ClientA_p] = make([]transport.Channel, 1)

	return &ClientA_q_Init{ept: session.NewEndpoint(1, 1, conn)}, nil
}

type ClientA_q_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientA_q_Init) Init() (*ClientA_q_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for _, conn := range st.ept.Conn[ClientA_p] {
		if conn == nil {
			return nil, fmt.Errorf("Invalid connection at ClientA_q")
		}
	}
	st.ept.ConnMu.Unlock()
	return &ClientA_q_1{ept: st.ept}, nil
}

func (st *ClientA_q_Init) Run(fn func(*ClientA_q_1) *ClientA_q_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st0)
}

type ClientA_q_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

// Send_PlayA sends a session from to p.
func (st *ClientA_q_1) Send_PlayA(args []*Game_a_Init) *ClientA_q_End {
	st.Use()

	st.ept.ConnMu.RLock()
	for i, arg := range args {
		// Here we only send the *session.Endpoint connections.
		st.ept.Conn[ClientA_p][i].Send(arg.ept)
		arg.Use()
	}
	st.ept.ConnMu.RUnlock()

	return &ClientA_q_End{}
}

type ClientA_q_End struct {
}

// global protocol ClientA, role p
func NewClientA_p() (*ClientA_p_Init, error) {
	// Omitted range check.

	conn := make(map[string][]transport.Channel)
	conn[ClientA_q] = make([]transport.Channel, 1)

	return &ClientA_p_Init{}, nil
}

type ClientA_p_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientA_p_Init) Init() (*ClientA_p_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for _, conn := range st.ept.Conn[ClientA_q] {
		if conn == nil {
			return nil, fmt.Errorf("Invalid connection at ClientA_p")
		}
	}
	st.ept.ConnMu.Unlock()
	return &ClientA_p_1{ept: st.ept}, nil
}

func (st *ClientA_p_Init) Run(fn func(*ClientA_p_1) *ClientA_p_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st0)
}

type ClientA_p_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientA_p_1) Recv_PlayA() ([]*Game_a_Init, *ClientA_p_End) {
	st.Use()

	sesss := make([]*Game_a_Init, len(st.ept.Conn[ClientA_q]))     // Session structs
	epts := make([]*session.Endpoint, len(st.ept.Conn[ClientA_q])) // Endpoints
	st.ept.ConnMu.RLock()
	for i, conn := range st.ept.Conn[ClientA_q] {
		err := conn.Recv(&epts[i])
		if err != nil {
			log.Fatalf("Wrong value received..")
		}
		sesss[i] = &Game_a_Init{ept: epts[i]} // Delegated session initialisation
	}
	st.ept.ConnMu.RUnlock()

	return sesss, &ClientA_p_End{}
}

type ClientA_p_End struct {
}

// Game_a_Init is the initial state of the session to be delegated (Game@a).
// Assumes It's already been setup but not executed.
type Game_a_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Game_a_Init) Init() (*Game_a_1, error) {
	st.Use()

	// --- Test connection ---

	return &Game_a_1{ept: st.ept}, nil
}

func (st *Game_a_Init) Run(fn func(*Game_a_1) *Game_a_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st0)
}

type Game_a_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

type Game_a_End struct {
}
