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
 */

package main

import (
	"../Regex"
	"flag"
	"fmt"
	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport"
	"github.com/nickng/scribble-go/runtime/transport/shm"
	"io/ioutil"
	"os"
	"regexp"
	"runtime"
	"time"
)

var allvariants = []string{
	"agggtaaa|tttaccct",
	"[cgt]gggtaaa|tttaccc[acg]",
	"a[act]ggtaaa|tttacc[agt]t",
	"ag[act]gtaaa|tttac[agt]ct",
	"agg[act]taaa|ttta[agt]cct",
	"aggg[acg]aaa|ttt[cgt]ccct",
	"agggt[cgt]aa|tt[acg]accct",
	"agggta[cgt]a|t[acg]taccct",
	"agggtaa[cgt]|[acg]ttaccct",
	"agggtaaa|tttaccct",
	"agggtaaa|tttaccct",
	"[cgt]gggtaaa|tttaccc[acg]",
	"a[act]ggtaaa|tttacc[agt]t",
	"ag[act]gtaaa|tttac[agt]ct",
	"agg[act]taaa|ttta[agt]cct",
	"aggg[acg]aaa|ttt[cgt]ccct",
	"agggt[cgt]aa|tt[acg]accct",
	"agggta[cgt]a|t[acg]taccct",
	"agggtaa[cgt]|[acg]ttaccct",
	"[cgt]gggtaaa|tttaccc[acg]",
	"a[act]ggtaaa|tttacc[agt]t",
	"ag[act]gtaaa|tttac[agt]ct",
	"agg[act]taaa|ttta[agt]cct",
	"aggg[acg]aaa|ttt[cgt]ccct",
	"agggt[cgt]aa|tt[acg]accct",
	"agggta[cgt]a|t[acg]taccct",
	"agggtaa[cgt]|[acg]ttaccct",
}

type Subst struct {
	pat, repl string
}

var substs = []Subst{
	Subst{"B", "(c|g|t)"},
	Subst{"D", "(a|g|t)"},
	Subst{"H", "(a|c|t)"},
	Subst{"K", "(g|t)"},
	Subst{"M", "(a|c)"},
	Subst{"N", "(a|c|g|t)"},
	Subst{"R", "(a|g)"},
	Subst{"S", "(c|g)"},
	Subst{"V", "(a|c|g)"},
	Subst{"W", "(a|t)"},
	Subst{"Y", "(c|t)"},
}

func countMatches(pat string, bytes []byte) int {
	re := regexp.MustCompile(pat)
	n := 0
	for {
		e := re.FindIndex(bytes)
		if e == nil {
			break
		}
		n++
		bytes = bytes[e[1]:]
	}
	return n
}

func main() {
	run_startt := time.Now()
	var nCPU int
	flag.IntVar(&nCPU, "ncpu", 8, "num goroutines")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	variants := allvariants[:nCPU]
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't read input: %s\n", err)
		os.Exit(2)
	}
	ilen := len(bytes)
	// Delete the comment lines and newlines
	bytes = regexp.MustCompile("(>[^\n]+)?\n").ReplaceAll(bytes, []byte{})
	clen := len(bytes)

	// Create connections
	connB := make([]transport.Transport, nCPU)
	for i := 0; i < nCPU; i++ {
		connB[i] = shm.NewConnection()
	}
	connC := shm.NewConnection()

	// instantiate protocol
	mini, _ := Regex.NewA(1, 1, nCPU, 1)
	for i, _ := range connB {
		session.Accept(mini, Regex.B, i+1, connB[i])
	}
	session.Accept(mini, Regex.C, 1, connC)

	// main session initiated, main function created
	mmain := func() {
		mini.Run(master(ilen, clen, variants))
	}

	// initialise C session
	bb := bytes

	cini, _ := Regex.NewC(1, 1, 1)
	session.Connect(cini, Regex.A, 1, connC)

	// C main function
	cmain := func() {
		cini.Run(substr(bb))
	}

	mkbmain := func(idx int) func() {
		bini, _ := Regex.NewB(1, idx+1, 1)
		session.Connect(bini, Regex.A, 1, connB[idx])
		return func() {
			bini.Run(worker(bytes))
		}
	}

	bmains := make([]func(), nCPU)
	for idx := 0; idx < nCPU; idx++ {
		bmains[idx] = mkbmain(idx)
	}

	go cmain()
	for idx := 0; idx < nCPU; idx++ {
		go bmains[idx]()
	}

	mmain()
	run_endt := time.Now()
	fmt.Println(ilen, "\t", nCPU, "\t", run_endt.Sub(run_startt).Nanoseconds())
}

func substr(bb []byte) func(*Regex.C_1) *Regex.C_End {
	return func(st1 *Regex.C_1) *Regex.C_End {
		_, st2 := st1.Measure()
		for _, sub := range substs {
			bb = regexp.MustCompile(sub.pat).ReplaceAll(bb, []byte(sub.repl))
		}
		return st2.Len(len(bb))
	}
}

func worker(bytes []byte) func(*Regex.B_1) *Regex.B_End {
	return func(st1 *Regex.B_1) *Regex.B_End {
		s, st2 := st1.Count()
		return st2.Donec(countMatches(s, bytes))
	}
}

func master(ilen, clen int, variants []string) func(*Regex.A_1) *Regex.A_End {
	return func(st1 *Regex.A_1) *Regex.A_End {

		st2 := st1.Count(variants)

		st4 := st2.Measure(0)

		rs, st3 := st4.Donec()
		a, ste := st3.Len()

		for i, c := range rs {
			//fmt.Printf("%s %d\n", variants[i], c)
			ioutil.Discard.Write(([]byte)(fmt.Sprintf("%s %d\n", variants[i], c)))
		}

		//fmt.Printf("\n%d\n%d\n%d\n", ilen, clen, a)
		ioutil.Discard.Write(([]byte)(fmt.Sprintf("\n%d\n%d\n%d\n", ilen, clen, a)))
		return ste
	}
}
