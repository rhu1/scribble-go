//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/auction/Bidder
//$ bin/Bidder.exe 8888 2

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/rhu1/scribble-go-runtime/test/util"
	"github.com/rhu1/scribble-go-runtime/test/auction/Auction/Proto"
)


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
	/*self, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatal(err)
	}*/

	wg := new(sync.WaitGroup)
	wg.Add(k)
	Proto := Proto.NewProto()
	for i := 1; i <= k; i++ {
		bidder := Proto.NewProto_Bidder_1Tok(k, i)
		/*if err != nil {
			log.Fatalf("Cannot create Bidder: %v", err)
		}*/

		p := port+i-1
		fmt.Println("Requesting", (strconv.Itoa(i) + ":"), p)
		//err := 
		bidder.Connect(Proto.Auctioneer, 1, util.LOCALHOST, strconv.Itoa(p))
		/*if err != nil {
			log.Fatalf("failed to create connection to Auctioneer: %v", err)
		}*/

		b1 := bidder.Init()
		go bidderFn(wg, b1, i, 100+i)
	}
	wg.Wait()
}


func bidderFn(wg *sync.WaitGroup, st *Proto.Proto_Bidder_1Tok_1, self int, MAXBID int) *Proto.Proto_Bidder_1Tok_End {
	fmt.Println(("(" + strconv.Itoa(self) + ")"), "bidderFn")
	var end *Proto.Proto_Bidder_1Tok_End
	var highest int
	var winner string

	var bids []int
	b3 := st.Send_Auctioneer_1To1_(10, util.Copy).Recv_Auctioneer_1To1_(&bids)
	highest = bids[0]

BID_LOOP:
	for {
		fmt.Println(("(" + strconv.Itoa(self) + ")"), "Current highest bid:", highest)
		var b4 *Proto.Proto_Bidder_1Tok_4
		giveUp := (highest > MAXBID)
		if giveUp {
			b4 = b3.Send_Auctioneer_1To1_(-1, util.Copy)
			fmt.Println(("(" + strconv.Itoa(self) + ")"), "Too high:", highest)
		} else {
			raised := highest+1
			b4 = b3.Send_Auctioneer_1To1_(raised, util.Copy)
			fmt.Println(("(" + strconv.Itoa(self) + ")"), "Raised bid:", raised)
		}
		select {
		case b3 = <-b4.Recv_Auctioneer_1To1_highest(&highest):
			fmt.Println(("(" + strconv.Itoa(self) + ")"), "Got bid:", highest)
		case end = <-b4.Recv_Auctioneer_1To1_winner(&winner):
			fmt.Println(("(" + strconv.Itoa(self) + ")"), "Got winner:", winner)
			break BID_LOOP
		}
	}
	wg.Done()
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
