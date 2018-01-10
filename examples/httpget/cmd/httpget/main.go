package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/rhu1/scribble-go-runtime/examples/httpget/httpget"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/shm"

	"github.com/rhu1/scribble-go-runtime/runtime/session"
	"github.com/rhu1/scribble-go-runtime/runtime/transport"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"
)

const (
	nServer  = 1
	nMaster  = 1
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

func Fetcher(id int, conns []transport.Transport, wg *sync.WaitGroup) {
	defer wg.Done()
	f, err := httpget.NewFetcher(id, nFetcher, nMaster, nServer)
	if err != nil {
		log.Fatalf("Cannot create new Fetcher: %v", err)
	}
	svrConn := tcp.NewConnection("127.0.0.1", "8100")
	svrConn.SerialiseMeth = tcp.SerialiseWithPassthru
	svrConn.DelimMeth = tcp.DelimitByCRLF
	for i := 1; i <= nServer; i++ {
		if err := session.Connect(f, httpget.Server, i, svrConn); err != nil {
			log.Fatalf("Cannot connect to %s[%d]: %v", httpget.Server, i, err)
		}
	}
	for i := 1; i <= nMaster; i++ {
		if err := session.Connect(f, httpget.Master, i, conns[id-1]); err != nil { // id - 1
			log.Fatalf("Cannot connect to %s[%d]: %v", httpget.Master, i, err)
		}
	}

	f.Ept().CheckConnection()
	var filepath string
	if err := f.Ept().Conn[httpget.Master][0].Recv(&filepath); err != nil {
		log.Fatalf("Cannot receive: %v", err)
	}

	headCmd := fmt.Sprintf("HEAD %s HTTP/1.1\r\nHost: 127.0.0.1\r\nConnection: keep-alive", filepath)
	if err := f.Ept().Conn[httpget.Server][0].Send(headCmd); err != nil {
		log.Fatalf("Cannot send: %v", err)
	}

	fmt.Printf("Request:\n%s\n\n", headCmd)

	var fileSize int
	reply := make([]byte, 4096)
	if err := f.Ept().Conn[httpget.Server][0].Recv(&reply); err != nil {
		log.Fatalf("Cannot recv: %v", err)
	}
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
	f.Ept().Conn[httpget.Master][0].Send(fileSize)

	// Recv size range from Master.
	var start, end int
	f.Ept().Conn[httpget.Master][0].Recv(&start)
	f.Ept().Conn[httpget.Master][0].Recv(&end)

	getCmd := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: 127.0.0.1\r\nRange: bytes=%d-%d", filepath, start, end)
	if err := f.Ept().Conn[httpget.Server][0].Send(getCmd); err != nil {
		log.Fatal("Cannot send:", err)
	}

	fmt.Printf("Request:\n%s\n\n", getCmd)

	replyHead := make([]byte, 4096)
	if err := f.Ept().Conn[httpget.Server][0].Recv(&replyHead); err != nil {
		log.Fatal("Cannot recv:", err)
	}

	fmt.Printf("Response HEAD:\n%s\n\n", string(replyHead))

	body := make([]byte, end-start)
	if err := f.Ept().Conn[httpget.Server][0].Recv(&body); err != nil {
		log.Fatal("Cannot recv:", err)
	}

	fmt.Printf("Response BODY:\n%d bytes\n\n", len(body))

	// Send to master to merge.
	f.Ept().Conn[httpget.Master][0].Send(string(body))
}

func Master(conns []transport.Transport, wg *sync.WaitGroup) {
	defer wg.Done()
	m, err := httpget.NewMaster(1, nFetcher, nMaster, nServer)
	if err != nil {
		log.Fatalf("Cannot create new Master: %v", err)
	}
	for i := 1; i <= nFetcher; i++ {
		if err := session.Accept(m, httpget.Fetcher, i, conns[i-1]); err != nil {
			log.Fatalf("Cannot connect to %s[%d]: %v", httpget.Fetcher, i, err)
		}
	}

	m.Run(func(master *httpget.Master_1) *httpget.Master_End {
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
