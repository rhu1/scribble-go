package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/nickng/scribble-go-runtime/examples/auction/Auction"
	"github.com/rhu1/scribble-go-runtime/runtime/session"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"
)

const nAuctioneer = 1

func main() {
	bidder, err := Auction.NewBidder(1, 1, 1)
	if err != nil {
		log.Fatalf("Cannot create Bidder: %v", err)
	}
	for i := 1; i <= nAuctioneer; i++ {
		conn := tcp.NewConnection("127.0.0.1", strconv.Itoa(33333+i))
		err := session.Connect(bidder, Auction.Auctioneer, i, conn)
		if err != nil {
			log.Fatalf("failed to create connection to Auctioneer: %v", err)
		}
	}

	bidder.Run(bidderFn)
}

const MAXBID = 100

func bidderFn(st *Auction.Bidder_1Ton_1) *Auction.Bidder_1Ton_End {
	fmt.Println("bidderFn")
	var end *Auction.Bidder_1Ton_End
	var highest int
	var winner string

	highests, b3 := st.SendAll(intGen(10, nAuctioneer)).RecvAll()
	highest = highests[0]
BID_LOOP:
	for {
		fmt.Println("Current highest bid:", highest)
		var b4 *Auction.Bidder_1Ton_4_Select
		giveUp := (highest > MAXBID)
		if giveUp {
			b4 = b3.SendAll([]Auction.IntOrBool{Auction.Bool(true)}).RecvAll()
		} else {
			b4 = b3.SendAll([]Auction.IntOrBool{Auction.Int(highest + 1)}).RecvAll()
		}
		select {
		case b3 = <-b4.Int(&highest):
			fmt.Println("Got bid:", highest)
		case end = <-b4.String(&winner):
			fmt.Println("Got winner:", winner)
			break BID_LOOP
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
