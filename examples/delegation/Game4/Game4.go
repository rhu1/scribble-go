package Game4

import (
	"fmt"
	"log"

	"github.com/rhu1/scribble-go-runtime/runtime/session"
	"github.com/rhu1/scribble-go-runtime/runtime/transport"
)

const (
	Client_client = "Client client" // Client protocol client role
	Client_server = "Client server" // Client protocol server role
)

func NewClient_server(id, nclient, nserver int) (*Client_server_Init, error) {

	conn := make(map[string][]transport.Channel)
	conn[Client_client] = make([]transport.Channel, nclient)

	return &Client_server_Init{ept: session.NewEndpoint(id, nserver, conn)}, nil
}

type Client_server_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Client_server_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *Client_server_Init) Init() (*Client_server_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for i, conn := range st.ept.Conn[Client_client] {
		if conn == nil {
			return nil, fmt.Errorf("Invalid connection at Client_server %d", i)
		}
	}
	st.ept.ConnMu.Unlock()
	return &Client_server_1{ept: st.ept}, nil
}

func (st *Client_server_Init) Run(fn func(*Client_server_1) *Client_server_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	if st0 == nil {
		log.Fatal("Client_server is nil")
	}
	fn(st0)
}

type Client_server_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Client_server_1) Send_Play(args ...*Game_player_1tok_Init) *Client_server_End {
	st.Use()

	st.ept.ConnMu.RLock()
	for i, arg := range args {
		st.ept.Conn[Client_client][i].Send(arg.ept)
		arg.Use()
	}
	st.ept.ConnMu.RUnlock()
	return &Client_server_End{}
}

type Client_server_End struct {
}

func NewClient_client(id, nclient, nserver int) (*Client_client_Init, error) {

	conn := make(map[string][]transport.Channel)
	conn[Client_server] = make([]transport.Channel, nserver)

	return &Client_client_Init{ept: session.NewEndpoint(id, nclient, conn)}, nil
}

type Client_client_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Client_client_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *Client_client_Init) Init() (*Client_client_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for _, conn := range st.ept.Conn[Client_server] {
		if conn == nil {
			return nil, fmt.Errorf("Invalid connection at Client_client")
		}
	}
	st.ept.ConnMu.Unlock()
	return &Client_client_1{ept: st.ept}, nil
}

func (st *Client_client_Init) Run(fn func(*Client_client_1) *Client_client_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	if st0 == nil {
		log.Fatal("Client_client is nil")
	}
	fn(st0)
}

type Client_client_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Client_client_1) Recv_Play() ([]*Game_player_1tok_Init, *Client_client_End) {
	st.Use()

	sesss := make([]*Game_player_1tok_Init, len(st.ept.Conn[Client_server]))
	st.ept.ConnMu.RLock()
	for i, conn := range st.ept.Conn[Client_server] {
		var ept *session.Endpoint
		err := conn.Recv(&ept)
		if err != nil {
			log.Fatalf("Wrong value received..")
		}
		sesss[i] = &Game_player_1tok_Init{ept: ept} // Delegated session initialisation
	}
	st.ept.ConnMu.RUnlock()
	return sesss, &Client_client_End{}
}

type Client_client_End struct {
}

const (
	Game_player = "Game player"
)

func NewGame_player_1tok(id, nplayer int) (*Game_player_1tok_Init, error) {

	conn := make(map[string][]transport.Channel)
	conn[Game_player] = make([]transport.Channel, 0) //nplayer-1)

	return &Game_player_1tok_Init{ept: session.NewEndpoint(id, nplayer, conn)}, nil
}

type Game_player_1tok_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Game_player_1tok_Init) Ept() *session.Endpoint {
	return st.ept
}

func (st *Game_player_1tok_Init) Init() (*Game_player_1tok_1, error) {
	st.Use()

	st.ept.ConnMu.Lock()
	for role, conns := range st.ept.Conn {
		for i, conn := range conns {
			if conn == nil {
				return nil, fmt.Errorf("Invalid connection at Game@player: %s[%d] is nil", role, i)
			}
		}
	}
	st.ept.ConnMu.Unlock()
	return &Game_player_1tok_1{ept: st.ept}, nil
}

func (st *Game_player_1tok_Init) Run(fn func(*Game_player_1tok_1) *Game_player_1tok_End) {
	st.ept.CheckConnection()

	st0, err := st.Init()
	if err != nil {
		log.Fatalf("Failed to initialise the session: %v", err)
	}
	fn(st0)
}

type Game_player_1tok_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (st *Game_player_1tok_1) Game() *Game_player_1tok_End {
	st.Use()

	// Implementation of Game

	return &Game_player_1tok_End{}
}

type Game_player_1tok_End struct {
}
