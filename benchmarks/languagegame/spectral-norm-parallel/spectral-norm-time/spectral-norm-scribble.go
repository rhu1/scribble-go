//

/*
Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

    * Redistributions of source code must retain the above copyright
    notice, this list of conditions and the following disclaimer.

    * Redistributions in binary form must reproduce the above copyright
    notice, this list of conditions and the following disclaimer in the
    documentation and/or other materials provided with the distribution.

    * Neither the name of "The Computer Language Benchmarks Game" nor the
    name of "The Computer Language Shootout Benchmarks" nor the names of
    its contributors may be used to endorse or promote products derived
    from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED.  IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

/* The Computer Language Benchmarks Game
 * http://shootout.alioth.debian.org/
 *
 * contributed by The Go Authors.
 * Based on spectral-norm.c by Sebastien Loisel
 */

package main

import (
	"flag"
	"fmt"
	"github.com/nickng/scribble-go/benchmarks/languagegame/spectral-norm-parallel/SN" // Protocol API
	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport"
	"github.com/nickng/scribble-go/runtime/transport/shm"
	"runtime"
	"time"
)

var n = flag.Int("n", 2000, "count")
var nCPU = flag.Int("ncpu", 4, "number of cpus")

func evalA(i, j int) float64 { return 1 / float64(((i+j)*(i+j+1)/2 + i + 1)) }

type Vec []float64

func (v Vec) Times(i, n int, u Vec) {
	for ; i < n; i++ {
		v[i] = 0
		for j := 0; j < len(u); j++ {
			v[i] += evalA(i, j) * u[j]
		}
	}
}

func (v Vec) TimesTransp(i, n int, u Vec) {
	for ; i < n; i++ {
		v[i] = 0
		for j := 0; j < len(u); j++ {
			v[i] += evalA(j, i) * u[j]
		}
	}
}

func main() {
	run_startt := time.Now()

	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	// create inputs for program
	N := *n
	u := make(Vec, N)
	for i := 0; i < N; i++ {
		u[i] = 1
	}
	v := make(Vec, N)

	// create connections

	conns := make([]transport.Transport, *nCPU)
	for i, _ := range conns {
		conns[i] = shm.NewBufferedConnection(100)
	}

	// instantiate protocol
	mini, _ := SN.NewA(1, 1, *nCPU)
	for i, _ := range conns {
		session.Accept(mini, SN.B, i+1, conns[i])
	}

	var x Vec

	// main session initiated, main function created
	mmain := func() {
		mini.Run(master(N, u, v, &x))
	}

	// instantiate workers with sub-roles
	workerInitialise := func(idx int) func() {
		ini, _ := SN.NewB(idx+1, *nCPU, 1)
		session.Connect(ini, SN.A, 1, conns[idx])
		mainf := worker(idx, u, v, &x)
		return func() {
			ini.Run(mainf)
		}
	}

	workers := make([]func(), *nCPU)
	for i := 0; i < *nCPU; i++ {
		workers[i] = workerInitialise(i)
	}

	for i := 0; i < *nCPU; i++ {
		go workers[i]()
	}

	mmain()
	run_endt := time.Now()

	fmt.Println(N, "\t", *nCPU, "\t", run_endt.Sub(run_startt).Nanoseconds())
}

func worker(i int, u, v Vec, x *Vec) func(*SN.B_1) *SN.B_End {
	return func(st1 *SN.B_1) *SN.B_End {
		var pl int
		var st2 *SN.B_2
		var st3 *SN.B_3
		var st4 *SN.B_4
		var st5 *SN.B_5
		var st6 *SN.B_6
		var st7 *SN.B_7
		var st8 *SN.B_8
		var ste *SN.B_End
		for {
			select {
			case st2 = <-st1.RecvTimes(&pl):
				// x.Times u followed by v.TimesTransp x
				(*x).Times(i*len(v) / *nCPU, (i+1)*len(v) / *nCPU, u)
				// tell master we are done
				st3 = st2.SendDone(0)
				_, st4 = st3.RecvNext()
				v.TimesTransp(i*len(v) / *nCPU, (i+1)*len(v) / *nCPU, *x)
				st5 = st4.SendDone(0)

				// now we are doing a u.TimesTransp(v), so u and v should be reversed in
				// the operations. Also, x should be fresh and of length (v)
				_, st6 = st5.RecvTimesTr()
				(*x).Times(i*len(u) / *nCPU, (i+1)*len(u) / *nCPU, v)
				st7 = st6.SendDone(0)
				_, st8 = st7.RecvNext()
				u.TimesTransp(i*len(u) / *nCPU, (i+1)*len(u) / *nCPU, *x)
				st1 = st8.SendDone(0)
			case ste = <-st1.RecvEnd(&pl):
				// last iteration
				return ste
			}
		}
	}
}

func master(N int, u, v Vec, x *Vec) func(*SN.A_1) *SN.A_End {
	return func(st1 *SN.A_1) *SN.A_End {
		var st2 *SN.A_2
		var st3 *SN.A_3
		var st4 *SN.A_4
		var st5 *SN.A_5
		var st6 *SN.A_6
		var st7 *SN.A_7
		var st8 *SN.A_8
		pl := make([]int, *nCPU)
		for i := 0; i < *nCPU; i++ {
			pl[i] = 1 + i
		}
		for i := 0; i < 10; i++ {
			// v.ATimesTransp(u)
			*x = make(Vec, len(u))
			st2 = st1.SendTimes(pl)
			_, st3 = st2.RecvDone()
			st4 = st3.SendNext(pl)
			_, st5 = st4.RecvDone()
			// u.ATimesTransp(v)
			*x = make(Vec, len(v))
			st6 = st5.SendTimesTr(pl)
			_, st7 = st6.RecvDone()
			st8 = st7.SendNext(pl)
			_, st1 = st8.RecvDone()
		}

		// after synchronisation finishes, continue with local computation
		var vBv, vv float64
		for i := 0; i < N; i++ {
			vBv += u[i] * v[i]
			vv += v[i] * v[i]
		}
		// fmt.Printf("%0.9f\n", math.Sqrt(vBv/vv))

		// finalise
		return st1.SendEnd(pl)
	}
}
