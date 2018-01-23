//rhu@HZHL4 ~/code/go
//$ go install github.com/rhu1/scribble-go-runtime/test/foo7
//$ bin/foo7.exe

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/tcp"
	"github.com/rhu1/scribble-go-runtime/runtime/transport/shm"

	"github.com/rhu1/scribble-go-runtime/test/foo/foo7/Foo7/Proto1"
	"github.com/rhu1/scribble-go-runtime/test/util"
)

// Bypass bloody annoying Go "unused import" errors
var _ = strconv.Itoa
var _ = tcp.NewAcceptor
var _ = shm.NewConnector

const PORT = 8888

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	n := 2

	wg := new(sync.WaitGroup)
	wg.Add(n + 1)

	as := make([]transport.Transport, n)
	for i := 1; i <= n; i++ {
		as[i-1] = tcp.NewAcceptor(strconv.Itoa(PORT+i))
		/*x := as[i-1].(tcp.ConnCfg)
		x.SerialiseMeth = tcp.SerialiseWithPassthru*/
		//as[i-1] = shm.NewConnector()
	}
	go serverCode(wg, n, as)

	time.Sleep(100 * time.Millisecond)

	for i := 1; i <= n; i++ {
		conn := tcp.NewRequestor(util.LOCALHOST, strconv.Itoa(PORT+i))
		/*conn.SerialiseMeth = tcp.SerialiseWithPassthru*/
		//conn := as[i-1]
		go clientCode(wg, n, i, conn)
	}

	wg.Wait()
}

func serverCode(wg *sync.WaitGroup, n int, conns []transport.Transport) *Proto1.Proto1_S_1To1_End {
	P1 := Proto1.NewProto1()

	S := P1.NewProto1_S_1To1(n, 1)
	for i := 1; i <= n; i++ {
		S.Accept(P1.W, i, conns[i-1])
	}
	s1 := S.Init()
	var end *Proto1.Proto1_S_1To1_End

	var bs []byte
	s2 := s1.Split_W_1Ton_norman([]byte {5,6,7,8}, util.CopyBates)
	end = s2.Reduce_W_1Ton_mum(&bs, foo)
	fmt.Println("S received:", bs)

	wg.Done()
	return end
}

func foo(bss [][]byte) []byte {
	bs := make([]byte, len(bss[0]))
	copy(bs, bss[0])
	for i := 1; i < len(bss); i++ {
		tmp := bss[i]
		for	j := 0; j < len(tmp); j++ {
			bs[j] = bs[j] + tmp[j]	
		}
	}
	return bs
}

func clientCode(wg *sync.WaitGroup, n int, self int, conn transport.Transport) *Proto1.Proto1_W_1Ton_End {
	P1 := Proto1.NewProto1()

	W := P1.NewProto1_W_1Ton(1, self)
	W.Request(P1.S, 1, conn)
	w1 := W.Init()
	var end *Proto1.Proto1_W_1Ton_End

	var bs []byte
	w2 := w1.Reduce_S_1To1_norman(&bs, util.UnaryReduceBates)
	fmt.Println("W" + strconv.Itoa(self) + ":", bs)
	end = w2.Split_S_1To1_mum([]byte{1, 2, 3, 4}, util.CopyBates)	

	wg.Done()
	return end
}
