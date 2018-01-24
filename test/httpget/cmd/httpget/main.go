package main

import (
	"bytes"
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/rhu1/scribble-go-runtime/runtime/transport"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/shm"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"

	"github.com/rhu1/scribble-go-runtime/test/httpget/HTTPget/Proto2"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

var (
	nFetcher int
	filePath string

	HTTPHost string
	HTTPPort int
)

func init() {
	flag.IntVar(&nFetcher, "fetcher", 2, "Number of Fetchers")
	flag.StringVar(&filePath, "filepath", "/main.go", "File path to download")

	flag.StringVar(&HTTPHost, "host", util.LOCALHOST, "HTTP server host")
	flag.IntVar(&HTTPPort, "port", 8100, "HTTP server port")
}

func main() {
	flag.Parse()

	// Shared memory connections.
	connsMu := new(sync.Mutex)
	connsMu.Lock()
	shmConns := make([]transport.Transport, nFetcher)
	for i := 0; i < nFetcher; i++ {
		shmConns[i] = shm.NewConnection()
	}
	connsMu.Unlock()

	wg := new(sync.WaitGroup)
	wg.Add(nFetcher + 1)

	go Master(shmConns, wg)

	waitFetcher := make(chan struct{})
	go Fetcher_1(shmConns[0], wg, waitFetcher)
	<-waitFetcher

	for i := 1; i < nFetcher; i++ {
		go Fetcher_2Ton(i+1, shmConns[i], wg)
	}
	wg.Wait()
}

func Fetcher_1(mastConn transport.Transport, wg *sync.WaitGroup, ch chan struct{}) *Proto2.Proto2_Fetcher_1To1and1Tok_End {
	debugf("[Fetcher 1] Fetcher_1()\n")
	defer wg.Done()

	servConn := tcp.NewRequestor(HTTPHost, strconv.Itoa(HTTPPort))
	servConn.SerialiseMeth = tcp.SerialiseWithPassthru
	servConn.DelimMeth = tcp.DelimitByCRLF

	P1 := Proto2.NewProto2()
	Fetcher := P1.NewProto2_Fetcher_1To1and1Tok(nFetcher, 1)
	Fetcher.Request(P1.Server, 1, servConn)
	Fetcher.Request(P1.Master, 1, mastConn)
	ch <- struct{}{}

	var filepath string
	f2 := Fetcher.Init().
		Reduce_Master_1To1_URL(&filepath, util.UnaryReduceString)
	debugf("[Fetcher 1] filepath: %s\n", filepath)

	headReq := httphead(filepath, HTTPHost, strconv.Itoa(HTTPPort))
	f3 := f2.Split_Server_1To1_(headReq, util.CopyString)
	debugf("[Fetcher 1] HTTP Request:\n%s\n\n", headReq)

	var filesize int
	reply := make([]byte, 4096)
	f4 := f3.Reduce_Server_1To1_(&reply, util.UnaryReduceBates)
	re := regexp.MustCompile(`Content-Length: (\d+)`)
	if matches := re.FindSubmatch(reply); len(matches) > 0 {
		i, err := strconv.Atoi(string(matches[1]))
		if err != nil {
			i = 0
		}
		filesize = i
	}
	debugf("[Fetcher 1] HTTP Response:\n%s\n\n", string(reply))
	debugf("[Fetcher 1] extracted filesize: %d\n", filesize)

	// Recv size range from Master.
	var start, end int
	f7 := f4.
		Split_Master_1To1_FileSize(filesize, util.Copy).
		Reduce_Master_1To1_start(&start, util.UnaryReduce).
		Reduce_Master_1To1_end(&end, util.UnaryReduce)

	getReq := httpget_chunked(filepath, HTTPHost, strconv.Itoa(HTTPPort), start, end)
	f8 := f7.Split_Server_1To1_(getReq, util.CopyString)
	debugf("[Fetcher 1] HTTP Request:\n%s\n\n", getReq)

	replyHead := make([]byte, 4096)
	f9 := f8.Reduce_Server_1To1_(&replyHead, util.UnaryReduceBates)
	debugf("[Fetcher 1] HTTP Response Header:\n%s\n\n", string(replyHead))

	replyBody := make([]byte, end-start)
	f10 := f9.Reduce_Server_1To1_(&replyBody, util.UnaryReduceBates)
	debugf("[Fetcher 1] HTTP Response Body (%d bytes):\n%s\n\n", len(replyBody), string(replyBody))

	// Send to master to merge.
	return f10.Split_Master_1To1_merge(string(replyBody), util.CopyString)
}

func Fetcher_2Ton(self int, mastConn transport.Transport, wg *sync.WaitGroup) *Proto2.Proto2_Fetcher_1Tok_not_1To1_End {
	debugf("[Fetcher %d/%d] Fetcher_2Ton()\n", self, nFetcher)
	defer wg.Done()

	P1 := Proto2.NewProto2()
	Fetcher := P1.NewProto2_Fetcher_1Tok_not_1To1(nFetcher, self)

	servConn := tcp.NewRequestor(HTTPHost, strconv.Itoa(HTTPPort))
	servConn.SerialiseMeth = tcp.SerialiseWithPassthru
	servConn.DelimMeth = tcp.DelimitByCRLF

	Fetcher.Request(P1.Server, 1, servConn) // Server[1]
	Fetcher.Request(P1.Master, 1, mastConn) // Master[1]
	debugf("[Fetcher %d/%d] connected\n", self, nFetcher)

	var filepath string
	f2 := Fetcher.Init().
		Reduce_Master_1To1_URL(&filepath, util.UnaryReduceString)
	debugf("[Fetcher %d/%d] filepath: %s\n", self, nFetcher, filepath)

	headReq := httphead(filepath, HTTPHost, strconv.Itoa(HTTPPort))
	f3 := f2.Split_Server_1To1_(headReq, util.CopyString)
	debugf("[Fetcher %d/%d] HTTP Request:\n%s\n\n", self, nFetcher, headReq)

	reply := make([]byte, 4096)
	f4 := f3.Reduce_Server_1To1_(&reply, util.UnaryReduceBates)

	debugf("[Fetcher %d/%d] HTTP Response:\n%s\n\n", self, nFetcher, string(reply))

	// Recv size range from Master.
	var start, end int
	f5 := f4.
		Reduce_Master_1To1_start(&start, util.UnaryReduce).
		Reduce_Master_1To1_end(&end, util.UnaryReduce)

	getReq := httpget_chunked(filepath, HTTPHost, strconv.Itoa(HTTPPort), start, end)
	f6 := f5.Split_Server_1To1_(getReq, util.CopyString)

	debugf("[Fetcher %d/%d] HTTP Request:\n%s\n\n", self, nFetcher, getReq)

	replyHead := make([]byte, 4096)
	f7 := f6.Reduce_Server_1To1_(&replyHead, util.UnaryReduceBates)

	debugf("[Fetcher %d/%d] HTTP Response Header:\n%s\n\n", self, nFetcher, string(replyHead))

	replyBody := make([]byte, end-start)
	f8 := f7.Reduce_Server_1To1_(&replyBody, util.UnaryReduceBates)

	debugf("[Fetcher %d/%d] HTTP Response Body (%d bytes):\n%s\n", self, nFetcher, len(replyBody), string(replyBody))

	// Send to master to merge.
	return f8.Split_Master_1To1_merge(string(replyBody), util.CopyString)
}

func Master(conns []transport.Transport, wg *sync.WaitGroup) *Proto2.Proto2_Master_1To1_End {
	defer wg.Done()

	P1 := Proto2.NewProto2()
	Master := P1.NewProto2_Master_1To1(nFetcher, 1)

	for i := 0; i < nFetcher; i++ {
		Master.Accept(P1.Fetcher, i+1, conns[i]) // Accept Fetcher[1..N]
	}

	return runMaster2(Master.Init())
}

func runMaster2(master *Proto2.Proto2_Master_1To1_1) *Proto2.Proto2_Master_1To1_End {
	debugf("[Master] filepath: %s\n", filePath)

	sizes := make([]int, nFetcher) // Pre-allocate size container
	m2 := master.Split_Fetcher_1Tok_URL(filePath, util.CopyString)
	debugf("[Master] sent filepath")

	m3 := m2.Recv_Fetcher_1To1_FileSize(&sizes)

	debugf("[Master] received file size %v\n", sizes)
	totalFilesize := sizes[0]
	chunkSize := totalFilesize / nFetcher
	debugf("[Master] totalFilesize: %d\n", totalFilesize)

	start, end := make([]int, nFetcher), make([]int, nFetcher)
	for i := 0; i < nFetcher; i++ {
		start[i] = i * chunkSize
		if i < nFetcher-1 {
			end[i] = (i+1)*chunkSize - 1
		} else {
			end[i] = totalFilesize
		}
		debugf("[Master] chunk for [Fetcher %d/%d]: %d--%d\n",
			i+1, nFetcher, start[i], end[i])
	}

	merges := make([]string, nFetcher)
	mEnd := m3.
		Send_Fetcher_1Tok_start(start).
		Send_Fetcher_1Tok_end(end).
		Recv_Fetcher_1Tok_merge(&merges)

	debugf("[Master] merged file:\n")
	fmt.Print(strings.Join(merges, ""))
	debugf("[Master] EOF merged file ==========\n")

	return mEnd
}

func httphead(path, host, port string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("HEAD %s HTTP/1.1\r\n", path))
	buf.WriteString(fmt.Sprintf("Host: %s:%s\r\n", host, port))
	buf.WriteString("Connection: keep-alive")
	return buf.String()
}

func httpget_chunked(path, host, port string, start, end int) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("GET %s HTTP/1.1\r\n", path))
	buf.WriteString(fmt.Sprintf("Host: %s:%s\r\n", host, port))
	buf.WriteString(fmt.Sprintf("Range: bytes=%d-%d", start, end))
	return buf.String()
}

func splitter(xs []int, i int) int {
	return xs[i]
}
