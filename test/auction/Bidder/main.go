//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/auction/Bidder
//$ bin/Bidder.exe 8888 2 1

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/rhu1/scribble-go-runtime/test/util"
	"github.com/rhu1/scribble-go-runtime/test/auction/Auction/Proto"
)


const nAuctioneer = 1


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
	self, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatal(err)
	}

	Proto := Proto.NewProto()
	bidder := Proto.NewProto_Bidder_1Tok(k, self)
	/*if err != nil {
		log.Fatalf("Cannot create Bidder: %v", err)
	}*/
	for i := 1; i <= nAuctioneer; i++ {
		//err := 
		bidder.Connect(Proto.Auctioneer, i, util.LOCALHOST, strconv.Itoa(port+self))
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

	var bids []int
	b3 := st.Send_Auctioneer_1To1_(10, util.Copy).Recv_Auctioneer_1To1_(&bids)
	highest = bids[0]

BID_LOOP:
	for {
		fmt.Println("Current highest bid:", highest)
		var b4 *Proto.Proto_Bidder_1Tok_4
		giveUp := (highest > MAXBID)
		if giveUp {
			b4 = b3.Send_Auctioneer_1To1_(-1, util.Copy)
		} else {
			raised := highest+1
			b4 = b3.Send_Auctioneer_1To1_(raised, util.Copy)
			fmt.Println("Raised bid:", raised)
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
