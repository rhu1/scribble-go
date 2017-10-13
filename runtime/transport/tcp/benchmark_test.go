package tcp

import (
	"bufio"
	"encoding/gob"
	"net"
	"sync"
	"testing"
)

// This benchmarks manual message passing performance.
// The net.TCPConn connection is mocked with net.Pipe
// for a more stable overhead measurement.
//
// The aim of this benchmark is to measure the message passing overhead.
// This measures repeated sending and receiving and excludes connection
// setup time, as this is the critical path of session-based runtime,
// i.e. single connection, repeated message passing in a session.
func BenchmarkManualMsgPassing(b *testing.B) {
	c, s := net.Pipe() // As obtained from Connect, Accept.
	wg := new(sync.WaitGroup)
	wg.Add(2)

	type ClientConn struct {
		Conn net.Conn
		Bufw *bufio.Writer
		Enc  *gob.Encoder
	}
	cln := ClientConn{Conn: c, Bufw: bufio.NewWriter(c)}
	cln.Enc = gob.NewEncoder(cln.Bufw)
	type ServerConn struct {
		Conn net.Conn
		Bufr *bufio.Reader
		Dec  *gob.Decoder
	}
	svr := ServerConn{Conn: s, Bufr: bufio.NewReader(s)}
	svr.Dec = gob.NewDecoder(svr.Bufr)
	client := func(N int, cc ClientConn) {
		for i := 0; i < N; i++ {
			str := "hello scribble"
			cc.Enc.Encode(str)
			cc.Bufw.Flush()
		}
		wg.Done()
	}
	server := func(N int, sc ServerConn) {
		var str string
		for i := 0; i < N; i++ {
			sc.Dec.Decode(&str)
			b.Logf("Received #%d: %s", i, str)
		}
		wg.Done()
	}

	go client(b.N, cln)
	go server(b.N, svr)
	wg.Wait()
}

// This benchmarks runtime message passing passing performance.
// tcp.Conn connection is mocked with net.Pipe
// for a more stable overhead measurement.
//
// The aim of this benchmark is to measure the message passing overhead.
// This measures repeated sending and receiving and excludes connection
// setup time, as this is the critical path of session-based runtime,
// i.e. single connection, repeated message passing in a session.
func BenchmarkRuntimeMsgPassing(b *testing.B) {
	c, s := net.Pipe()                      // As obtained from Connect, Accept.
	cfg := NewConnection("127.0.0.1", ":0") // Mock connection.
	wg := new(sync.WaitGroup)
	wg.Add(2)

	client := func(N int, conn *Conn) {
		for i := 0; i < N; i++ {
			conn.Send("hello scribble")
		}
		wg.Done()
	}
	server := func(N int, conn *Conn) {
		var str string
		for i := 0; i < N; i++ {
			conn.Recv(&str)
			b.Logf("Received #%d: %s", i, str)
		}
		wg.Done()
	}

	go client(b.N, cfg.newConn(c))
	go server(b.N, cfg.newConn(s))
	wg.Wait()
}
