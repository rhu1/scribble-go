package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/rhu1/scribble-go-runtime/runtime/session"
	"github.com/rhu1/scribble-go-runtime/runtime/transport"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"

	"github.com/rhu1/scribble-go-runtime/test/util"
	"github.com/rhu1/scribble-go-runtime/test/httpget/HTTPget/Proto1"
)

const (
	/*nServer  = 1
	nMaster  = 1*/
	nFetcher = 2
)

func main() {
	// Shared memory connections.
	connsMu := new(sync.Mutex)
	connsMu.Lock()
	sharedConns := make([]transport.Transport, nFetcher)
	for i := 0; i < nFetcher; i++ {
		sharedConns[i] = shm.NewConnection()
	}
	connsMu.Unlock()

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go Master(sharedConns, wg)
	go Fetcher(1, sharedConns, wg)
	go Fetcher(2, sharedConns, wg)
	wg.Wait()
}

func Fetcher(self int, conns []transport.Transport, wg *sync.WaitGroup) {
	defer wg.Done()
	/*f, err := HTTPget.NewFetcher(id, nFetcher, nMaster, nServer)
	if err != nil {
		log.Fatalf("Cannot create new Fetcher: %v", err)
	}*/

	P1 := Proto1.NewProto1()
	Fetcher := P1.NewProto1_Fetcher_1Tok(nFetcher, self)

	/*svrConn := tcp.NewConnection("127.0.0.1", "8100")
	svrConn.SerialiseMeth = tcp.SerialiseWithPassthru
	svrConn.DelimMeth = tcp.DelimitByCRLF
	for i := 1; i <= nServer; i++ {
		if err := session.Connect(f, HTTPget.Server, i, svrConn); err != nil {
			log.Fatalf("Cannot connect to %s[%d]: %v", HTTPget.Server, i, err)
		}
	}*/
	conn := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(8100))
	Fetcher.Request(P1.Server, 1, conn)
	/*svrConn.SerialiseMeth = tcp.SerialiseWithPassthru
	svrConn.DelimMeth = tcp.DelimitByCRLF*/

	/*for i := 1; i <= nMaster; i++ {
		if err := session.Connect(f, HTTPget.Master, i, conns[id-1]); err != nil { // id - 1
			log.Fatalf("Cannot connect to %s[%d]: %v", HTTPget.Master, i, err)
		}
	}*/

	//Fetcher.Ept().CheckConnection()
	f1 := Fetcher.Init()
	//var end *Proto1.Proto1_Fetcher_1Tok_End

	var filepath string
	/*if err := Fetcher.Ept().Conn[Proto1.Master][0].Recv(&filepath); err != nil {
		log.Fatalf("Cannot receive: %v", err)
	}*/
	f2 := f1.Reduce_Master_1To1_URL(&filepath, util.UnaryReduceString)

	headCmd := fmt.Sprintf("HEAD %s HTTP/1.1\r\nHost: 127.0.0.1\r\nConnection: keep-alive", filepath)
	/*if err := f.Ept().Conn[HTTPget.Server][0].Send(headCmd); err != nil {
		log.Fatalf("Cannot send: %v", err)
	}*/
	f3 := f2.Send_Server_1To1_HEAD(headCmd, util.CopyString)

	fmt.Printf("Request:\n%s\n\n", headCmd)

	var fileSize int
	reply := make([]byte, 4096)
	/*if err := f.Ept().Conn[HTTPget.Server][0].Recv(&reply); err != nil {
		log.Fatalf("Cannot recv: %v", err)
	}*/
	f4 := f3.Reduce_Server_1To1_response(&reply, util.UnaryReduce)
	re := regexp.MustCompile(`Content-Length: (\d+)`)
	if matches := re.FindSubmatch(reply); len(matches) > 0 {
		i, err := strconv.Atoi(string(matches[1]))
		if err != nil {
			i = 0
		}
		fileSize = i
	}

	fmt.Printf("Response:\n%s\n\n", string(reply))

	// Send filesize to Master.
	//f.Ept().Conn[HTTPget.Master][0].Send(fileSize)
	f5 := f4.Send_Master_1To1_FileSize(fileSize, util.Copy)

	// Recv size range from Master.
	var start, end int
	/*f.Ept().Conn[HTTPget.Master][0].Recv(&start)
	f.Ept().Conn[HTTPget.Master][0].Recv(&end)*/
	f7 := f5.Reduce_Master_1To1_start(&start, util.UnaryReduce).Reduce_Master_1To1_end(&end, util.UnaryReduce)

	getCmd := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: 127.0.0.1\r\nRange: bytes=%d-%d", filepath, start, end)
	/*if err := f.Ept().Conn[HTTPget.Server][0].Send(getCmd); err != nil {
		log.Fatal("Cannot send:", err)
	}*/
	f8 := f7.Send_Server_1To1_GET(getCmd, util.CopyString)

	fmt.Printf("Request:\n%s\n\n", getCmd)

	replyHead := make([]byte, 4096)
	/*if err := f.Ept().Conn[HTTPget.Server][0].Recv(&replyHead); err != nil {
		log.Fatal("Cannot recv:", err)
	}*/
	f9 := f8.Reduce_Server_1To1_Response(&replyHead, util.UnaryReduce)

	fmt.Printf("Response HEAD:\n%s\n\n", string(replyHead))

	body := make([]byte, end-start)
	/*if err := f.Ept().Conn[HTTPget.Server][0].Recv(&body); err != nil {
		log.Fatal("Cannot recv:", err)
	}*/
	f10 := f9.Reduce_Server_1To1_Body(&body, util.UnaryReduceString)

	fmt.Printf("Response BODY:\n%d bytes\n\n", len(body))

	// Send to master to merge.
	//f.Ept().Conn[HTTPget.Master][0].Send(string(body))
	f10.Send_Master_1To1_merge(string(body), util.Copy)
}

func Master(conns []transport.Transport, wg *sync.WaitGroup) {
	defer wg.Done()
	m, err := HTTPget.NewMaster(1, nFetcher, nMaster, nServer)
	if err != nil {
		log.Fatalf("Cannot create new Master: %v", err)
	}
	for i := 1; i <= nFetcher; i++ {
		if err := session.Accept(m, HTTPget.Fetcher, i, conns[i-1]); err != nil {
			log.Fatalf("Cannot connect to %s[%d]: %v", HTTPget.Fetcher, i, err)
		}
	}

	m.Run(func(master *HTTPget.Master_1) *HTTPget.Master_End {
		URLs := make([]string, nFetcher)
		for i := 0; i < nFetcher; i++ {
			URLs[i] = "/main.go"
		}
		sizes, st3 := master.
			SendAll_Fetcher_URL(URLs).
			RecvAll_Fetcher_FileSize()
		fileSize := sizes[0]
		chunkSize := fileSize / nFetcher

		start := make([]int, nFetcher)
		end := make([]int, nFetcher)

		fmt.Printf("Master: fileSize=%d\n", fileSize)

		for i := 0; i < nFetcher; i++ {
			start[i] = i * chunkSize
			if i < nFetcher-1 {
				end[i] = (i+1)*chunkSize - 1
			} else {
				end[i] = fileSize
			}
			fmt.Printf("chunk %d: %d-%d\n", i, start[i], end[i])
		}

		merges, stEnd := st3.
			SendAll_Fetcher_start(start).
			SendAll_Fetcher_end(end).
			RecvAll_Fetcher_merge()

		fmt.Printf("\n-- merge --\n\n%v\n", strings.Join(merges, ""))

		return stEnd
	})
}
