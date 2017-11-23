package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nickng/scribble-go/examples/delegation/Game"
	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport"
	"github.com/nickng/scribble-go/runtime/transport/shm"
)

const (
	ClientA_Server = iota
	ClientB_Server
	ClientC_Server
)

func main() {
	sharedCons := make([]transport.Transport, 3)
	sharedCons[ClientA_Server] = shm.NewConnection()
	sharedCons[ClientB_Server] = shm.NewConnection()
	sharedCons[ClientC_Server] = shm.NewConnection()

	wg := new(sync.WaitGroup)
	wg.Add(4)

	go server(sharedCons, wg)
	go clientA(sharedCons[ClientA_Server], wg)
	go clientB(sharedCons[ClientB_Server], wg)
	go clientC(sharedCons[ClientC_Server], wg)
	wg.Wait()
}

func clientA(conn transport.Transport, wg *sync.WaitGroup) {
	defer wg.Done()
	A_p, err := Game.NewClientA_p()
	if err != nil {
		log.Fatalf("Cannot create ClientA p: %v", err)
	}
	if err := session.Connect(A_p, Game.ClientA_q, 1, conn); err != nil {
		log.Fatalf("Cannot connect to %s[%d]: %v", Game.ClientA_q, 1, err)
	}

	A_p.Run(func(st *Game.ClientA_p_1) *Game.ClientA_p_End {
		gameA, stEnd := st.Recv_PlayA()

		gameA[0].Run(func(st *Game.Game_a_1) *Game.Game_a_End {
			fmt.Println("Game@A")
			return st.Game()
		})

		return stEnd
	})
}

func clientB(conn transport.Transport, wg *sync.WaitGroup) {
	defer wg.Done()
	B_p, err := Game.NewClientB_p()
	if err != nil {
		log.Fatalf("Cannot create ClientB p: %v", err)
	}
	if err := session.Connect(B_p, Game.ClientB_q, 1, conn); err != nil {
		log.Fatalf("Cannot connect to %s[%d]: %v", Game.ClientB_q, 1, err)
	}

	B_p.Run(func(st *Game.ClientB_p_1) *Game.ClientB_p_End {
		gameB, stEnd := st.Recv_PlayB()

		gameB[0].Run(func(st *Game.Game_b_1) *Game.Game_b_End {
			fmt.Println("Game@B")
			return st.Game()
		})

		return stEnd
	})
}

func clientC(conn transport.Transport, wg *sync.WaitGroup) {
	defer wg.Done()
	C_p, err := Game.NewClientC_p()
	if err != nil {
		log.Fatalf("Cannot create ClientC p: %v", err)
	}
	if err := session.Connect(C_p, Game.ClientC_q, 1, conn); err != nil {
		log.Fatalf("Cannot connect to %s[%d]: %v", Game.ClientC_q, 1, err)
	}

	C_p.Run(func(st *Game.ClientC_p_1) *Game.ClientC_p_End {
		gameC, stEnd := st.Recv_PlayC()

		gameC[0].Run(func(st *Game.Game_c_1) *Game.Game_c_End {
			fmt.Println("Game@C")
			return st.Game()
		})

		return stEnd
	})
}

func server(conns []transport.Transport, wg *sync.WaitGroup) {
	defer wg.Done()
	A_q, err := Game.NewClientA_q()
	if err != nil {
		log.Fatalf("Cannot create ClientA q: %v", err)
	}
	if err := session.Accept(A_q, Game.ClientA_p, 1, conns[ClientA_Server]); err != nil {
		log.Fatalf("Connect connect from %s[%d]: %v", Game.ClientA_p, 1, err)
	}

	B_q, err := Game.NewClientB_q()
	if err != nil {
		log.Fatalf("Cannot create ClientB q: %v", err)
	}
	if err := session.Accept(B_q, Game.ClientB_p, 1, conns[ClientB_Server]); err != nil {
		log.Fatalf("Connect connect from %s[%d]: %v", Game.ClientB_p, 1, err)
	}

	C_q, err := Game.NewClientC_q()
	if err != nil {
		log.Fatalf("Cannot create ClientC q: %v", err)
	}
	if err := session.Accept(C_q, Game.ClientC_p, 1, conns[ClientC_Server]); err != nil {
		log.Fatalf("Connect connect from %s[%d]: %v", Game.ClientC_p, 1, err)
	}

	log.Println("------------------- Clients connected ---------------------")

	// Now setup the Game.
	AB := shm.NewConnection()
	BC := shm.NewConnection()
	CA := shm.NewConnection()

	Game_a, err := Game.NewGame_a()
	if err != nil {
		log.Fatalf("Cannot create Game a: %v", err)
	}
	Game_b, err := Game.NewGame_b()
	if err != nil {
		log.Fatalf("Cannot create Game a: %v", err)
	}
	Game_c, err := Game.NewGame_c()
	if err != nil {
		log.Fatalf("Cannot create Game a: %v", err)
	}

	if err := session.Accept(Game_b, Game.Game_a, 1, AB); err != nil {
		log.Fatalf("Cannot accept Game B from A")
	}
	if err := session.Accept(Game_c, Game.Game_b, 1, BC); err != nil {
		log.Fatalf("Cannot accept Game C from B")
	}
	if err := session.Accept(Game_a, Game.Game_c, 1, CA); err != nil {
		log.Fatalf("Cannot accept Game C from A")
	}

	if err := session.Connect(Game_a, Game.Game_b, 1, AB); err != nil {
		log.Fatalf("Cannot connect Game A to B")
	}
	if err := session.Connect(Game_b, Game.Game_c, 1, BC); err != nil {
		log.Fatalf("Cannot connect Game B to C")
	}
	if err := session.Connect(Game_c, Game.Game_a, 1, CA); err != nil {
		log.Fatalf("Cannot connect Game C to A")
	}

	log.Println("------------------- Setup done ----------------------------")
	time.Sleep(1 * time.Second)

	// Delegate the Game.

	A_q.Run(func(st *Game.ClientA_q_1) *Game.ClientA_q_End {
		return st.Send_PlayA(Game_a)
	})
	B_q.Run(func(st *Game.ClientB_q_1) *Game.ClientB_q_End {
		return st.Send_PlayB(Game_b)
	})
	C_q.Run(func(st *Game.ClientC_q_1) *Game.ClientC_q_End {
		return st.Send_PlayC(Game_c)
	})
}
