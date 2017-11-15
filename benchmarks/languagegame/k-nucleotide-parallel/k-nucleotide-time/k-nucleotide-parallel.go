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
	"../KNuc" // Protocol API
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport"
	"github.com/nickng/scribble-go/runtime/transport/shm"
	"io/ioutil"
	"os"
	"runtime"
	"sort"
	"time"
)

func count(data string, n int) map[string]int {
	counts := make(map[string]int)
	top := len(data) - n
	for i := 0; i <= top; i++ {
		s := data[i : i+n]
		counts[s]++
	}
	return counts
}

func countOne(data string, s string) int {
	return count(data, len(s))[s]
}

type kNuc struct {
	name  string
	count int
}

type kNucArray []kNuc

func (kn kNucArray) Len() int      { return len(kn) }
func (kn kNucArray) Swap(i, j int) { kn[i], kn[j] = kn[j], kn[i] }
func (kn kNucArray) Less(i, j int) bool {
	if kn[i].count == kn[j].count {
		return kn[i].name > kn[j].name // sort down
	}
	return kn[i].count > kn[j].count
}

func sortedArray(m map[string]int) kNucArray {
	kn := make(kNucArray, len(m))
	i := 0
	for k, v := range m {
		kn[i] = kNuc{k, v}
		i++
	}
	sort.Sort(kn)
	return kn
}

func printKnucs(a kNucArray) {
	sum := 0
	for _, kn := range a {
		sum += kn.count
	}
	for _, kn := range a {
		ioutil.Discard.Write(([]byte)(fmt.Sprintf("%s %.3f\n", kn.name, 100*float64(kn.count)/float64(sum))))
		//fmt.Printf("%s %.3f\n", kn.name, 100*float64(kn.count)/float64(sum))
	}
	// fmt.Print("\n")
}

var nCPU int

func main() {
	run_startt := time.Now()
	flag.IntVar(&nCPU, "ncpu", 8, "GOMAXPROCS")
	flag.Parse()
	runtime.GOMAXPROCS(8)

	in := bufio.NewReader(os.Stdin)
	three := []byte(">THREE ")
	for {
		line, err := in.ReadSlice('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "ReadLine err:", err)
			os.Exit(2)
		}
		if line[0] == '>' && bytes.Equal(line[0:len(three)], three) {
			break
		}
	}
	data, err := ioutil.ReadAll(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ReadAll err:", err)
		os.Exit(2)
	}
	// delete the newlines and convert to upper case
	j := 0
	for i := 0; i < len(data); i++ {
		if data[i] != '\n' {
			data[j] = data[i] &^ ' ' // upper case
			j++
		}
	}
	str := string(data[0:j])

	interests := []string{"GGT", "GGTA", "GGTATT", "GGTATTTTAATT",
		"GGT", "GGTA", "GGTATT", "GGTATTTTAATT",
		"GGT", "GGTA", "GGTATT", "GGTATTTTAATT",
		"GGT", "GGTA", "GGTATT", "GGTATTTTAATT",
		"GGT", "GGTA", "GGTATT", "GGTATTTTAATT"}

	var arr1, arr2 kNucArray

	// Create connections
	connS := make([]transport.Transport, 2)
	for i := 0; i < 2; i++ {
		connS[i] = shm.NewConnection()
	}
	connB := make([]transport.Transport, nCPU)
	for i := 0; i < nCPU; i++ {
		connB[i] = shm.NewConnection()
	}

	// instantiate protocol
	mini, _ := KNuc.NewA(1, 1, 2, nCPU)
	for i, _ := range connS {
		session.Accept(mini, KNuc.S, i+1, connS[i])
	}
	for i, _ := range connB {
		session.Accept(mini, KNuc.B, i+1, connB[i])
	}

	// main session initiated, main function created
	mmain := func() {
		mini.Run(master(&arr1, &arr2, interests))
	}

	inp := [2]*kNucArray{&arr1, &arr2}

	sorterInitialise := func(idx int) func() {
		ini, _ := KNuc.NewS(idx+1, 2, 1)
		session.Connect(ini, KNuc.A, 1, connS[idx])
		mainf := sorter(idx, str, inp[idx])
		return func() {
			ini.Run(mainf)
		}
	}
	sorter1 := sorterInitialise(0)
	sorter2 := sorterInitialise(1)

	workerInitialise := func(idx int) func() {
		ini, _ := KNuc.NewB(idx+1, nCPU, 1)
		session.Connect(ini, KNuc.A, 1, connB[idx])
		mainf := worker(str)
		return func() {
			ini.Run(mainf)
		}
	}
	workers := make([]func(), nCPU)
	for idx := 0; idx < nCPU; idx++ {
		workers[idx] = workerInitialise(idx)
	}

	// run sorters + workers

	go sorter1()
	go sorter2()
	for idx := 0; idx < nCPU; idx++ {
		go workers[idx]()
	}
	mmain()
	run_endt := time.Now()

	fmt.Println(len(data), "\t", nCPU, "\t", run_endt.Sub(run_startt).Nanoseconds())
}

func worker(str string) func(*KNuc.B_1) *KNuc.B_End {
	return func(st1 *KNuc.B_1) *KNuc.B_End {

		ss, st2 := st1.Recv_BA()

		result := fmt.Sprintf("%d %s\n", countOne(str, ss), ss)

		return st2.Send_BA(result)
	}
}

func sorter(i int, str string, arr *kNucArray) func(*KNuc.S_1) *KNuc.S_End {
	return func(st1 *KNuc.S_1) *KNuc.S_End {

		_, st2 := st1.Recv_SA()

		*arr = sortedArray(count(str, i+1))

		return st2.Send_SA(i)
	}
}

func master(arr1, arr2 *kNucArray, interests []string) func(*KNuc.A_1) *KNuc.A_End {

	return func(st1 *KNuc.A_1) *KNuc.A_End {

		ids := []int{1, 2}

		_, st2 := st1.SendS(ids).SendB(interests[:nCPU]).RecvS()

		printKnucs(*arr1)
		printKnucs(*arr2)

		res, ste := st2.RecvB()

		for _, rc := range res {
			ioutil.Discard.Write(([]byte)(rc))
			// fmt.Print(rc)
		}

		return ste
	}
}
