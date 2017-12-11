//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/auction/Bidder
//$ bin/Bidder.exe

package main

import (
	"fmt"
	"log"
	"strconv"

	//"github.com/rhu1/scribble-go-runtime/runtime/session"
	//"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"

	"github.com/rhu1/scribble-go-runtime/test/auction/Auction/Proto"
)

const nAuctioneer = 1

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	Proto := Proto.NewProto()
	bidder := Proto.NewProto_Bidder_1Tok(1, 1)
	/*if err != nil {
		log.Fatalf("Cannot create Bidder: %v", err)
	}*/
	for i := 1; i <= nAuctioneer; i++ {
		//err := 
		bidder.Connect(Proto.Auctioneer, i, "127.0.0.1", strconv.Itoa(33333+i))
		/*if err != nil {
			log.Fatalf("failed to create connection to Auctioneer: %v", err)
		}*/
	}

	b1 := bidder.Init()
	bidderFn(b1)
}

const MAXBID = 100

func bidderFn(st *Proto.Proto_Bidder_1Tok_1) *Proto.Proto_Bidder_1Tok_End {
	fmt.Println("bidderFn")
	var end *Proto.Proto_Bidder_1Tok_End
	var highest int
	var winner string

	b3 := st.Send_Auctioneer_1To1_(10, mydup).Reduce_Auctioneer_1To1_(&highest, mysum)

BID_LOOP:
	for {
		fmt.Println("Current highest bid:", highest)
		var b4 *Proto.Proto_Bidder_1Tok_4
		giveUp := (highest > MAXBID)
		if giveUp {
			b4 = b3.Send_Auctioneer_1To1_(-1, mydup)
		} else {
			b4 = b3.Send_Auctioneer_1To1_(highest+1, mydup)
		}
		select {
		case b3 = <-b4.Recv_Auctioneer_1To1_highest(&highest):
			fmt.Println("Got bid:", highest)
		case end = <-b4.Recv_Auctioneer_1To1_winner(&winner):
			fmt.Println("Got winner:", winner)
			break BID_LOOP
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
