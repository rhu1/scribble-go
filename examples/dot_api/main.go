package main

import (
	"github.com/nickng/scribble-go-runtime/examples/dot_api/dot"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

/*
global protocol Foo(role A(n), role B(n) {
	(int) dot A[1..n] to B[1..n]
}
*/
func main() {
	wg := new(sync.WaitGroup)
	wg.Add(20)

	roleACode := func(i int) {
		roleAIni, err := dot.NewRoleA(i, 10, 10)
		if err != nil {
			log.Fatalf("cannot create roleA endpoint: %s", err)
		}
		// One connection for each participant in the group
		err = roleAIni.Accept(dot.RoleB, i, "127.0.0.1", strconv.Itoa(33333+i))
		if err != nil {
			log.Fatalf("failed to create connection to participant %d of role 'roleB': %s", i, err)
		}

		roleAMain := mkroleAmain(i)
		roleAIni.Run(roleAMain)
		wg.Done()
	}

	for n := 1; n <= 10; n++ {
		go roleACode(n)
	}

	time.Sleep(1000 * time.Millisecond)

	roleBCode := func(i int) {
		roleBIni, err := dot.NewRoleB(i, 10, 10)
		if err != nil {
			log.Fatalf("cannot create roleB endpoint: %s", err)
		}
		// One connection for each participant in the group
		err = roleBIni.Connect(dot.RoleA, i, "127.0.0.1", strconv.Itoa(33333+i))
		if err != nil {
			log.Fatalf("failed to create connection from participant %d of role 'roleB': %s", i, err)
		}

		roleBMain := mkroleBmain(i)
		roleBIni.Run(roleBMain)
		wg.Done()
	}

	for i := 1; i <= 10; i++ {
		go roleBCode(i)
	}

	wg.Wait()
}

func mkroleAmain(nw int) func(st1 *dot.RoleA_1To1_1) *dot.RoleA_1To1_End {
	return func(st1 *dot.RoleA_1To1_1) *dot.RoleA_1To1_End {
		return st1.DotSend(42 + nw)
	}
}

func mkroleBmain(idx int) func(st1 *dot.RoleB_1Ton_1) *dot.RoleB_1Ton_End {
	return func(st1 *dot.RoleB_1Ton_1) *dot.RoleB_1Ton_End {
		v, ste := st1.DotRecv()
		fmt.Println(idx, " received: ", v)
		return ste
	}
}
