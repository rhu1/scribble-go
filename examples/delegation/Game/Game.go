package Game

import (
	"fmt"
	"log"

	"github.com/rhu1/scribble-go-runtime/runtime/session"
	"github.com/rhu1/scribble-go-runtime/runtime/transport"
)

const (
	ClientA_p = "ClientA p"
	ClientA_q = "ClientA q"
	ClientB_p = "ClientB p"
	ClientB_q = "ClientB q"
	ClientC_p = "ClientC p"
	ClientC_q = "ClientC q"

	Game_a = "Game a"
	Game_b = "Game b"
	Game_c = "Game c"
)

// global protocol ClientA, role p
func NewClientA_p() (*ClientA_p_Init, error) {
	// Omitted range check.

	conn := make(map[string][]transport.Channel)
	conn[ClientA_q] = make([]transport.Channel, 1)

	return &ClientA_p_Init{ept: session.NewEndpoint(1, 1, conn)}, nil
}

type ClientA_p_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientA_p_Init) Ept() *session.Endpoint {
	return st.ept
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
	if st0 == nil {
		log.Fatal("Client_A_p is nil")
	}
	fn(st0)
}

type ClientA_p_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientA_p_1) Recv_PlayA() ([]*Game_a_Init, *ClientA_p_End) {
	st.Use()

	sesss := make([]*Game_a_Init, len(st.ept.Conn[ClientA_q])) // Session structs
	st.ept.ConnMu.RLock()
	for i, conn := range st.ept.Conn[ClientA_q] {
		var ept *session.Endpoint
		err := conn.Recv(&ept)
		if err != nil {
			log.Fatalf("Wrong value received..")
		}
		sesss[i] = &Game_a_Init{ept: ept} // Delegated session initialisation
	}
	st.ept.ConnMu.RUnlock()
	return sesss, &ClientA_p_End{}
}

type ClientA_p_End struct {
}

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

func (st *ClientA_q_Init) Ept() *session.Endpoint {
	return st.ept
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

func (st *ClientA_q_1) Send_PlayA(args ...*Game_a_Init) *ClientA_q_End {
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

// global protocol ClientB, role p
func NewClientB_p() (*ClientB_p_Init, error) {
	// Omitted range check.

	conn := make(map[string][]transport.Channel)
	conn[ClientB_q] = make([]transport.Channel, 1)

	return &ClientB_p_Init{ept: session.NewEndpoint(1, 1, conn)}, nil
}

type ClientB_p_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientB_p_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *ClientB_p_Init) Init() (*ClientB_p_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for _, conn := range st.ept.Conn[ClientB_q] {
		if conn == nil {
			return nil, fmt.Errorf("Invalid connection at ClientB_p")
		}
	}
	st.ept.ConnMu.Unlock()
	return &ClientB_p_1{ept: st.ept}, nil
}

func (st *ClientB_p_Init) Run(fn func(*ClientB_p_1) *ClientB_p_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st0)
}

type ClientB_p_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientB_p_1) Recv_PlayB() ([]*Game_b_Init, *ClientB_p_End) {
	st.Use()

	sesss := make([]*Game_b_Init, len(st.ept.Conn[ClientB_q])) // Session structs
	st.ept.ConnMu.RLock()
	for i, conn := range st.ept.Conn[ClientB_q] {
		var ept *session.Endpoint
		err := conn.Recv(&ept)
		if err != nil {
			log.Fatalf("Wrong value received..")
		}
		sesss[i] = &Game_b_Init{ept: ept} // Delegated session initialisation
	}
	st.ept.ConnMu.RUnlock()

	return sesss, &ClientB_p_End{}
}

type ClientB_p_End struct {
}

// global protocol ClientB, role q
func NewClientB_q() (*ClientB_q_Init, error) {
	// Omitted range check.

	conn := make(map[string][]transport.Channel)
	conn[ClientB_p] = make([]transport.Channel, 1)

	return &ClientB_q_Init{ept: session.NewEndpoint(1, 1, conn)}, nil
}

type ClientB_q_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientB_q_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *ClientB_q_Init) Init() (*ClientB_q_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for _, conn := range st.ept.Conn[ClientB_p] {
		if conn == nil {
			return nil, fmt.Errorf("Invalid connection at ClientB_q")
		}
	}
	st.ept.ConnMu.Unlock()
	return &ClientB_q_1{ept: st.ept}, nil
}

func (st *ClientB_q_Init) Run(fn func(*ClientB_q_1) *ClientB_q_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st0)
}

type ClientB_q_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientB_q_1) Send_PlayB(args ...*Game_b_Init) *ClientB_q_End {
	st.Use()

	st.ept.ConnMu.RLock()
	for i, arg := range args {
		// Here we only send the *session.Endpoint connections.
		st.ept.Conn[ClientB_p][i].Send(arg.ept)
		arg.Use()
	}
	st.ept.ConnMu.RUnlock()
	return &ClientB_q_End{}
}

type ClientB_q_End struct {
}

// global protocol ClientC, role p
func NewClientC_p() (*ClientC_p_Init, error) {
	// Omitted range check.

	conn := make(map[string][]transport.Channel)
	conn[ClientC_q] = make([]transport.Channel, 1)

	return &ClientC_p_Init{ept: session.NewEndpoint(1, 1, conn)}, nil
}

type ClientC_p_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientC_p_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *ClientC_p_Init) Init() (*ClientC_p_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for _, conn := range st.ept.Conn[ClientC_q] {
		if conn == nil {
			return nil, fmt.Errorf("Invalid connection at ClientC_p")
		}
	}
	st.ept.ConnMu.Unlock()
	return &ClientC_p_1{ept: st.ept}, nil
}

func (st *ClientC_p_Init) Run(fn func(*ClientC_p_1) *ClientC_p_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st0)
}

type ClientC_p_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientC_p_1) Recv_PlayC() ([]*Game_c_Init, *ClientC_p_End) {
	st.Use()

	sesss := make([]*Game_c_Init, len(st.ept.Conn[ClientC_q])) // Session structs
	st.ept.ConnMu.RLock()
	for i, conn := range st.ept.Conn[ClientC_q] {
		var ept *session.Endpoint
		err := conn.Recv(&ept)
		if err != nil {
			log.Fatalf("Wrong value received..")
		}
		sesss[i] = &Game_c_Init{ept: ept} // Delegated session initialisation
	}
	st.ept.ConnMu.RUnlock()
	return sesss, &ClientC_p_End{}
}

type ClientC_p_End struct {
}

// global protocol ClientC, role q
func NewClientC_q() (*ClientC_q_Init, error) {
	// Omitted range check.

	conn := make(map[string][]transport.Channel)
	conn[ClientC_p] = make([]transport.Channel, 1)

	return &ClientC_q_Init{ept: session.NewEndpoint(1, 1, conn)}, nil
}

type ClientC_q_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientC_q_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *ClientC_q_Init) Init() (*ClientC_q_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for _, conn := range st.ept.Conn[ClientC_p] {
		if conn == nil {
			return nil, fmt.Errorf("Invalid connection at ClientC_q")
		}
	}
	st.ept.ConnMu.Unlock()
	return &ClientC_q_1{ept: st.ept}, nil
}

func (st *ClientC_q_Init) Run(fn func(*ClientC_q_1) *ClientC_q_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st0)
}

type ClientC_q_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *ClientC_q_1) Send_PlayC(args ...*Game_c_Init) *ClientC_q_End {
	st.Use()

	st.ept.ConnMu.RLock()
	for i, arg := range args {
		// Here we only send the *session.Endpoint connections.
		st.ept.Conn[ClientC_p][i].Send(arg.ept)
		arg.Use()
	}
	st.ept.ConnMu.RUnlock()
	return &ClientC_q_End{}
}

type ClientC_q_End struct {
}

func NewGame_a() (*Game_a_Init, error) {
	// Omitted range check.

	conn := make(map[string][]transport.Channel)
	conn[Game_b] = make([]transport.Channel, 1)
	conn[Game_c] = make([]transport.Channel, 1)

	return &Game_a_Init{ept: session.NewEndpoint(1, 1, conn)}, nil
}

// Game_a_Init is the initial state of the session to be delegated (Game@a).
// Assumes It's already been setup but not executed.
type Game_a_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Game_a_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *Game_a_Init) Init() (*Game_a_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for role, conns := range st.ept.Conn {
		for i, conn := range conns {
			if conn == nil {
				return nil, fmt.Errorf("Invalid connection at Game@a: %s[%d] is nil", role, i)
			}
		}
	}
	st.ept.ConnMu.Unlock()
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

func (st *Game_a_1) Game() *Game_a_End {
	st.Use()
	return &Game_a_End{}
}

type Game_a_End struct {
}

func NewGame_b() (*Game_b_Init, error) {
	// Omitted range check.

	conn := make(map[string][]transport.Channel)
	conn[Game_a] = make([]transport.Channel, 1)
	conn[Game_c] = make([]transport.Channel, 1)
	c := Game_b_Init{ept: session.NewEndpoint(1, 1, conn)}

	return &c, nil
}

// Game_b_Init is the initial state of the session to be delegated (Game@b).
// Assumes It's already been setup but not executed.
type Game_b_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Game_b_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *Game_b_Init) Init() (*Game_b_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for role, conns := range st.ept.Conn {
		for i, conn := range conns {
			if conn == nil {
				return nil, fmt.Errorf("Invalid connection at Game@b: %s[%d] is nil", role, i)
			}
		}
	}
	st.ept.ConnMu.Unlock()
	return &Game_b_1{ept: st.ept}, nil
}

func (st *Game_b_Init) Run(fn func(*Game_b_1) *Game_b_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st0)
}

type Game_b_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Game_b_1) Game() *Game_b_End {
	st.Use()
	return &Game_b_End{}
}

type Game_b_End struct {
}

func NewGame_c() (*Game_c_Init, error) {
	// Omitted range check.

	conn := make(map[string][]transport.Channel)
	conn[Game_a] = make([]transport.Channel, 1)
	conn[Game_b] = make([]transport.Channel, 1)
	c := Game_c_Init{ept: session.NewEndpoint(1, 1, conn)}

	return &c, nil
}

// Game_c_Init is the initial state of the session to be delegated (Game@b).
// Assumes It's already been setup but not executed.
type Game_c_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Game_c_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *Game_c_Init) Init() (*Game_c_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for role, conns := range st.ept.Conn {
		for i, conn := range conns {
			if conn == nil {
				return nil, fmt.Errorf("Invalid connection at Game@c: %s[%d] is nil", role, i)
			}
		}
	}
	st.ept.ConnMu.Unlock()
	return &Game_c_1{ept: st.ept}, nil
}

func (st *Game_c_Init) Run(fn func(*Game_c_1) *Game_c_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st0)
}

type Game_c_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Game_c_1) Game() *Game_c_End {
	st.Use()
	return &Game_c_End{}
}

type Game_c_End struct {
}
