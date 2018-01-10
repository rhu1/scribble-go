//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/auction/Auctioneer
//$ bin/Auctioneer.exe 8888 1

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"

	"github.com/rhu1/scribble-go-runtime/test/auction/Auction/Proto"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

//type myintslice = []int

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	args := os.Args[1:]
	port, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal(err)
	}
	k, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatal(err)
	}

	Proto := Proto.NewProto()
	auctioneer := Proto.NewProto_Auctioneer_1To1(k, 1)
	/*if err != nil {
		log.Fatalf("Cannot create Auctioneer: %v", err)
	}*/
	wg := new(sync.WaitGroup)
	wg.Add(k)
	for i := 1; i <= k; i++ {
		//err := //session.Accept(auctioneer, Auction.Bidder, i, conn)
		go func(j int) {
			p := port + j - 1
			fmt.Println("Waiting:", p)
			//auctioneer.Accept(Proto.Bidder, j, util.LOCALHOST, strconv.Itoa(p))
			conn := tcp.NewAcceptor(strconv.Itoa(p))
			auctioneer.Accept(Proto.Bidder, j, conn)
			wg.Done()
			fmt.Println("Done:", p)
		}(i)
		/*if err != nil {
			log.Fatalf("failed to create connection to Bidder %d: %v", i, err)
		}*/
	}
	wg.Wait()

	a1 := auctioneer.Init()
	auctioneerFn(a1)
}

func auctioneerFn(st *Proto.Proto_Auctioneer_1To1_1) *Proto.Proto_Auctioneer_1To1_End {
	fmt.Println("auctioneerFn")
	var end *Proto.Proto_Auctioneer_1To1_End

	var bids []int
	st2 := st.Recv_Bidder_1Tok_(&bids)
	var highest, winnerID int
	for i := range bids {
		if bids[i] > highest {
			highest = bids[i]
			winnerID = i + 1
		}
	}
	// bids -> intGen
	st3 := st2.Send_Bidder_1Tok_(highest, util.Copy)
BID_LOOP:
	for {
		var bidSkips []int
		st4 := st3.Recv_Bidder_1Tok_(&bidSkips)
		var bidCount int
		for i, bs := range bidSkips {
			if bs > -1 {
				if bs > highest {
					highest = bs
					winnerID = i + 1
				}
				bidCount++
			}
		}
		hasWinner := (bidCount == 1)
		if hasWinner {
			fmt.Println("Current highest bid:", highest, "bidding ends")
			st4.Send_Bidder_1Tok_winner(strconv.Itoa(winnerID), mystrdup)
			break BID_LOOP
		} else {
			fmt.Println("Current highest bid:", highest, "bidding continues")
			st3 = st4.Send_Bidder_1Tok_highest(highest, util.Copy)
		}
	}
	return end
}

/*func intGen(v int, count int) []int {
	ints := make([]int, count)
	for i := range ints {
		ints[i] = v
	}
	return ints
}

func stringGen(v string, count int) []string {
	strs := make([]string, count)
	for i := range strs {
		strs[i] = v
	}
	return strs
}*/

func mystrdup(data string, i int) string {
	return data
}
