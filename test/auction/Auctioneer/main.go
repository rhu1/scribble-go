package main

import (
	"fmt"
	//"log"
	"strconv"

	//"github.com/nickng/scribble-go-runtime/runtime/session"
	//"github.com/nickng/scribble-go-runtime/runtime/transport/tcp"

	"github.com/rhu1/scribble-go-runtime/test/auction/Auction/Proto"
)

type myintslice = []int

const nBidder = 1

func main() {
	Proto := Proto.NewProto()
	auctioneer := Proto.NewProto_Auctioneer_1To1(1, 1)
	/*if err != nil {
		log.Fatalf("Cannot create Auctioneer: %v", err)
	}*/
	for i := 1; i <= nBidder; i++ {
		//err := //session.Accept(auctioneer, Auction.Bidder, i, conn)
		auctioneer.Accept(Proto.Bidder, i, "127.0.0.1", strconv.Itoa(33333+i))
		/*if err != nil {
			log.Fatalf("failed to create connection to Bidder %d: %v", i, err)
		}*/
	}

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
			winnerID = i
		}
	}
	// bids -> intGen
	st3 := st2.Send_Bidder_1Tok_(highest, mydup)
BID_LOOP:
	for {
		var bidSkips []int
		st4 := st3.Recv_Bidder_1Tok_(&bidSkips)
		var bidCount int
		for i, bs := range bidSkips {
			if bs > -1 {
				if bs > highest {
					highest = bs
					winnerID = i
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
			st3 = st4.Send_Bidder_1Tok_highest(highest, mydup)
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

func mysum(xs []int) int {
	res := 0
	for j := 0; j < len(xs); j++ {
		res = res + xs[j]	
	}
	return res
}

func mydup(data int, i int) int {
	return data
}

func mystrdup(data string, i int) string {
	return data
}
