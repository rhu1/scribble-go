package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/nickng/scribble-go-runtime/examples/auction/Auction"
	"github.com/rhu1/scribble-go-runtime/runtime/session"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"
)

const nBidder = 1

func main() {
	auctioneer, err := Auction.NewAuctioneer(1, 1, 1)
	if err != nil {
		log.Fatalf("Cannot create Auctioneer: %v", err)
	}
	for i := 1; i <= nBidder; i++ {
		conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
		err := session.Accept(auctioneer, Auction.Bidder, i, conn)
		if err != nil {
			log.Fatalf("failed to create connection to Bidder %d: %v", i, err)
		}
	}

	auctioneer.Run(auctioneerFn)
}

func auctioneerFn(st *Auction.Auctioneer_1To1_1) *Auction.Auctioneer_1To1_End {
	fmt.Println("auctioneerFn")
	var end *Auction.Auctioneer_1To1_End

	bids, st2 := st.RecvAll()
	var highest, winnerID int
	for i := range bids {
		if bids[i] > highest {
			highest = bids[i]
			winnerID = i
		}
	}
	// bids -> intGen
	st3 := st2.SendAll(intGen(highest, 1))
BID_LOOP:
	for {
		bidSkips, st4 := st3.RecvAll()
		var bidCount int
		for i, bs := range bidSkips {
			switch bid := bs.(type) {
			case Auction.Int:
				if int(bid) > highest {
					highest = int(bid)
					winnerID = i
				}
				bidCount++
			case Auction.Bool:
			}
		}
		hasWinner := (bidCount == 1)
		if hasWinner {
			fmt.Println("Current highest bid:", highest, "bidding ends")
			st4.SendAll_string(stringGen(strconv.Itoa(winnerID), 1))
			break BID_LOOP
		} else {
			fmt.Println("Current highest bid:", highest, "bidding continues")
			st3 = st4.SendAll_int(intGen(highest, 1))
		}
	}
	return end
}

func intGen(v int, count int) []int {
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
}
