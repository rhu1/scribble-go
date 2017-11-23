package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/nickng/scribble-go/examples/delegation/Game4"
	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport"
	"github.com/nickng/scribble-go/runtime/transport/shm"
)

const (
	nServer = 1
	nClient = 1
)

func main() {
	sharedCons := make([]transport.Transport, nClient)
	for i := range sharedCons {
		sharedCons[i] = shm.NewConnection()
	}

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go server(sharedCons, wg)
	go client(1, sharedCons[0], wg)
	//go client(2, sharedCons[1], wg)
	//go client(3, sharedCons[2], wg)
	//go client(4, sharedCons[3], wg)

	wg.Wait()
}

func client(id int, conn transport.Transport, wg *sync.WaitGroup) {
	defer wg.Done()
	client, err := Game4.NewClient_client(id, nClient, nServer)
	if err != nil {
		log.Fatalf("Cannot create Client %d: %v", id, err)
	}
	if err := session.Connect(client, Game4.Client_server, 1, conn); err != nil {
		log.Fatalf("Cannot connect to %s[%d]: %v", Game4.Client_server, 1, err)
	}

	client.Run(func(st *Game4.Client_client_1) *Game4.Client_client_End {
		games, st0 := st.Recv_Play()
		games[id-1].Run(func(st *Game4.Game_player_1tok_1) *Game4.Game_player_1tok_End {
			fmt.Println("Game4.")
			return st.Game()
		})
		return st0
	})
}

const (
	nPlayer = nClient
)

func server(conns []transport.Transport, wg *sync.WaitGroup) {
	defer wg.Done()
	server, err := Game4.NewClient_server(1, nClient, nServer)
	if err != nil {
		log.Fatalf("Cannot create server: %v", err)
	}
	for i := 1; i <= nClient; i++ {
		if err := session.Accept(server, Game4.Client_client, i, conns[i-1]); err != nil {
			log.Fatalf("Cannot accept from %s[%d]: %v", Game4.Client_client, i, err)
		}
	}

	Game_Players := make([]*Game4.Game_player_1tok_Init, nPlayer)
	for id := 1; id <= nPlayer; id++ {
		var err error
		Game_Players[id-1], err = Game4.NewGame_player_1tok(id, nPlayer)
		if err != nil {
			log.Fatalf("Cannot initialise Game at %s[%d]", Game4.Game_player, id-1)
		}
	}

	server.Run(func(st *Game4.Client_server_1) *Game4.Client_server_End {
		return st.Send_Play(Game_Players...)
	})
}
