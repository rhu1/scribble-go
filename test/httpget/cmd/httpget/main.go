package main

import (
	"fmt"
	//"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

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
	shmConns := make([]transport.Transport, nFetcher)
	for i := 0; i < nFetcher; i++ {
		shmConns[i] = shm.NewConnection()
	}
	connsMu.Unlock()

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go Master(shmConns, wg)

	ch := make(chan interface{})
	go Fetcher_1(shmConns[0], wg, ch)
	<-ch

	for i := 1; i < nFetcher; i++ {
		go Fetcher_2Ton(i+1, shmConns[i], wg)
	}
	wg.Wait()
}

func Fetcher_1(mastConn transport.Transport, wg *sync.WaitGroup, ch chan interface{}) *Proto1.Proto1_Fetcher_1To1and1Tok_End {
	defer wg.Done()
	/*f, err := HTTPget.NewFetcher(id, nFetcher, nMaster, nServer)
	if err != nil {
		log.Fatalf("Cannot create new Fetcher: %v", err)
	}*/

	P1 := Proto1.NewProto1()
	Fetcher := P1.NewProto1_Fetcher_1To1and1Tok(nFetcher, 1)

	/*for i := 1; i <= nServer; i++ {
		if err := session.Connect(f, HTTPget.Server, i, svrConn); err != nil {
			log.Fatalf("Cannot connect to %s[%d]: %v", HTTPget.Server, i, err)
		}
	}*/
	servConn := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(8100))
	servConn.SerialiseMeth = tcp.SerialiseWithPassthru
	servConn.DelimMeth = tcp.DelimitByCRLF
	Fetcher.Request(P1.Server, 1, servConn)
	/*svrConn.SerialiseMeth = tcp.SerialiseWithPassthru
	svrConn.DelimMeth = tcp.DelimitByCRLF*/

	/*for i := 1; i <= nMaster; i++ {
		if err := session.Connect(f, HTTPget.Master, i, conns[id-1]); err != nil { // id - 1
			log.Fatalf("Cannot connect to %s[%d]: %v", HTTPget.Master, i, err)
		}
	}*/
	Fetcher.Request(P1.Master, 1, mastConn)
	ch <- nil

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
	f3 := f2.Split_Server_1To1_HEAD(headCmd, util.CopyString)

	fmt.Printf("Request:\n%s\n\n", headCmd)

	var fileSize int
	reply := make([]byte, 4096)
	/*if err := f.Ept().Conn[HTTPget.Server][0].Recv(&reply); err != nil {
		log.Fatalf("Cannot recv: %v", err)
	}*/
	f4 := f3.Reduce_Server_1To1_response(&reply, util.UnaryReduceBates)
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
	f5 := f4.Split_Master_1To1_FileSize(fileSize, util.Copy)

	// Recv size range from Master.
	var start, end int
	/*f.Ept().Conn[HTTPget.Master][0].Recv(&start)
	f.Ept().Conn[HTTPget.Master][0].Recv(&end)*/
	f7 := f5.Reduce_Master_1To1_start(&start, util.UnaryReduce).Reduce_Master_1To1_end(&end, util.UnaryReduce)

	getCmd := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: 127.0.0.1\r\nRange: bytes=%d-%d", filepath, start, end)
	/*if err := f.Ept().Conn[HTTPget.Server][0].Send(getCmd); err != nil {
		log.Fatal("Cannot send:", err)
	}*/
	f8 := f7.Split_Server_1To1_GET(getCmd, util.CopyString)

	fmt.Printf("Request:\n%s\n\n", getCmd)

	replyHead := make([]byte, 4096)
	/*if err := f.Ept().Conn[HTTPget.Server][0].Recv(&replyHead); err != nil {
		log.Fatal("Cannot recv:", err)
	}*/
	f9 := f8.Reduce_Server_1To1_Response(&replyHead, util.UnaryReduceBates)

	fmt.Printf("Response HEAD:\n%s\n\n", string(replyHead))

	body := make([]byte, end-start)
	/*if err := f.Ept().Conn[HTTPget.Server][0].Recv(&body); err != nil {
		log.Fatal("Cannot recv:", err)
	}*/
	f10 := f9.Reduce_Server_1To1_Body(&body, util.UnaryReduceBates)

	fmt.Printf("Response BODY:\n%d bytes\n\n", len(body))

	// Send to master to merge.
	//f.Ept().Conn[HTTPget.Master][0].Send(string(body))
	return f10.Split_Master_1To1_merge(string(body), util.CopyString)
}

func Fetcher_2Ton(self int, mastConn transport.Transport, wg *sync.WaitGroup) *Proto1.Proto1_Fetcher_1Tok_not_1To1_End {
	defer wg.Done()

	P1 := Proto1.NewProto1()
	Fetcher := P1.NewProto1_Fetcher_1Tok_not_1To1(nFetcher, self)

	servConn := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(8100))
	servConn.SerialiseMeth = tcp.SerialiseWithPassthru
	servConn.DelimMeth = tcp.DelimitByCRLF
	Fetcher.Request(P1.Server, 1, servConn)

	Fetcher.Request(P1.Master, 1, mastConn)

	f1 := Fetcher.Init()
	//var end *Proto1.Proto1_Fetcher_1Tok_End

	var filepath string
	f2 := f1.Reduce_Master_1To1_URL(&filepath, util.UnaryReduceString)

	headCmd := fmt.Sprintf("HEAD %s HTTP/1.1\r\nHost: 127.0.0.1\r\nConnection: keep-alive", filepath)
	f3 := f2.Split_Server_1To1_HEAD(headCmd, util.CopyString)

	fmt.Printf("Request:\n%s\n\n", headCmd)

	reply := make([]byte, 4096)
	f4 := f3.Reduce_Server_1To1_response(&reply, util.UnaryReduceBates)

	fmt.Printf("Response:\n%s\n\n", string(reply))

	// Recv size range from Master.
	var start, end int
	f5 := f4.Reduce_Master_1To1_start(&start, util.UnaryReduce).Reduce_Master_1To1_end(&end, util.UnaryReduce)

	getCmd := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: 127.0.0.1\r\nRange: bytes=%d-%d", filepath, start, end)
	f6 := f5.Split_Server_1To1_GET(getCmd, util.CopyString)

	fmt.Printf("Request:\n%s\n\n", getCmd)

	replyHead := make([]byte, 4096)
	f7 := f6.Reduce_Server_1To1_Response(&replyHead, util.UnaryReduceBates)

	fmt.Printf("Response HEAD:\n%s\n\n", string(replyHead))

	body := make([]byte, end-start)
	f8 := f7.Reduce_Server_1To1_Body(&body, util.UnaryReduceBates)

	fmt.Printf("Response BODY:\n%d bytes\n\n", len(body))

	// Send to master to merge.
	return f8.Split_Master_1To1_merge(string(body), util.CopyString)
}

func Master(conns []transport.Transport, wg *sync.WaitGroup) *Proto1.Proto1_Master_1To1_End {
	defer wg.Done()
	/*m, err := HTTPget.NewMaster(1, nFetcher, nMaster, nServer)
	if err != nil {
		log.Fatalf("Cannot create new Master: %v", err)
	}*/
	P1 := Proto1.NewProto1()
	Master := P1.NewProto1_Master_1To1(nFetcher, 1)

	for i := 1; i <= nFetcher; i++ {
		/*if err := session.Accept(m, HTTPget.Fetcher, i, conns[i-1]); err != nil {
			log.Fatalf("Cannot connect to %s[%d]: %v", HTTPget.Fetcher, i, err)
		}*/
		Master.Accept(P1.Fetcher, i, conns[i])
	}

	m := Master.Init()
	//var end *Proto1.Proto1_Master_1To1_End
	
	return runMaster(m)
}

func runMaster(master *Proto1.Proto1_Master_1To1_1) *Proto1.Proto1_Master_1To1_End {
	/*URLs := make([]string, nFetcher)
	for i := 0; i < nFetcher; i++ {
		URLs[i] = "/main.go"
	}*/
	var sizes []int
	//sizes,
	st3 := master.
		/*SendAll_Fetcher_URL(URLs).
		RecvAll_Fetcher_FileSize()*/
		Split_Fetcher_1Tok_URL("/main.go", util.CopyString).
		Recv_Fetcher_1To1_FileSize(&sizes)
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

	var merges []string
	//merges,
	stEnd := st3.
		/*SendAll_Fetcher_start(start).
		SendAll_Fetcher_end(end).
		RecvAll_Fetcher_merge()*/
		Send_Fetcher_1Tok_start(start).
		Send_Fetcher_1Tok_end(end).
		Recv_Fetcher_1Tok_merge(&merges)

	fmt.Printf("\n-- merge --\n\n%v\n", strings.Join(merges, ""))

	return stEnd
}

func splitter(xs []int, i int) int {
	return xs[i]	
}
