package Regex

import (
	"fmt"
	"github.com/nickng/scribble-go-runtime/runtime/session"
	"log"
)

const A = "A"
const B = "B"
const C = "C"

func check(e error) {
	if e != nil {
		log.Panic(e.Error())
	}
}

/*****************************************************************************/
/************ A API **********************************************************/
/*****************************************************************************/
type A_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *A_Init) Ept() *session.Endpoint {
	return ini.ept
}

func NewA(id, numA, numB, numC int) (*A_Init, error) {
	session.RoleRange(id, numA)
	conn, err := session.NewConn([]session.ParamRole{{B, numB}, {C, numC}})
	if err != nil {
		return nil, err
	}

	return &A_Init{session.LinearResource{}, session.NewEndpoint(id, numA, conn)}, nil
}

type A_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *A_Init) Init() (*A_1, error) {
	ini.Use()
	ini.ept.ConnMu.Lock()
	for n, _ := range ini.ept.Conn {
		for j, _ := range ini.ept.Conn[n] {
			for ini.ept.Conn[n][j] == nil {
			}
		}
	}
	ini.ept.ConnMu.Unlock()
	return &A_1{session.LinearResource{}, ini.ept}, nil
}

type A_2 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *A_1) Count(pl []string) *A_2 {
	ini.Use()
	if len(pl) != len(ini.ept.Conn[B]) {
		log.Panicf("Incorrect number of arguments to role 'A' Count")
	}
	ini.ept.ConnMu.RLock()
	for i, c := range ini.ept.Conn[B] {
		check(c.Send(pl[i]))
	}
	ini.ept.ConnMu.RUnlock()
	return &A_2{session.LinearResource{}, ini.ept}
}

type A_3 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *A_2) Measure(pl int) *A_3 {
	ini.Use()
	if 1 != len(ini.ept.Conn[C]) {
		log.Panicf("Incorrect number of arguments to role 'C' Measure")
	}
	ini.ept.ConnMu.RLock()
	check(ini.ept.Conn[C][0].Send(pl))
	ini.ept.ConnMu.RUnlock()
	return &A_3{session.LinearResource{}, ini.ept}
}

type A_4 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *A_3) Donec() ([]int, *A_4) {
	ini.Use()
	var tmp int
	pl := make([]int, len(ini.ept.Conn[B]))

	ini.ept.ConnMu.RLock()
	for i, c := range ini.ept.Conn[B] {
		check(c.Recv(&tmp))
		pl[i] = tmp
	}
	ini.ept.ConnMu.RUnlock()
	return pl, &A_4{session.LinearResource{}, ini.ept}
}

type A_End struct {
}

func (ini *A_4) Len() (int, *A_End) {
	var tmp int

	ini.ept.ConnMu.RLock()
	check(ini.ept.Conn[C][0].Recv(&tmp))
	ini.ept.ConnMu.RUnlock()

	return tmp, &A_End{}
}

func (ini *A_Init) Run(f func(*A_1) *A_End) {
	ini.ept.CheckConnection()
	st1, err := ini.Init()
	check(err)
	f(st1)
}

/************ A API **********************************************************/

/*****************************************************************************/
/************ B API **********************************************************/
/*****************************************************************************/
type B_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *B_Init) Ept() *session.Endpoint {
	return ini.ept
}

func NewB(id, numB, numA int) (*B_Init, error) {
	session.RoleRange(id, numB)
	conn, err := session.NewConn([]session.ParamRole{{A, numA}})
	if err != nil {
		return nil, err
	}

	return &B_Init{session.LinearResource{}, session.NewEndpoint(id, numB, conn)}, nil
}

type B_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *B_Init) Init() (*B_1, error) {
	ini.Use()
	ini.ept.ConnMu.Lock()
	for n, l := range ini.ept.Conn {
		for i, c := range l {
			if c == nil {
				return nil, fmt.Errorf("nvalid connection for worker %s[%d] at B[%d]", n, i, ini.Ept().Id)
			}
		}
	}
	ini.ept.ConnMu.Unlock()
	return &B_1{session.LinearResource{}, ini.ept}, nil
}

type B_2 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *B_1) Count() (string, *B_2) {
	ini.Use()
	var tmp string

	ini.ept.ConnMu.RLock()
	check(ini.ept.Conn[A][0].Recv(&tmp))
	ini.ept.ConnMu.RUnlock()
	return tmp, &B_2{session.LinearResource{}, ini.ept}
}

type B_End struct {
}

func (ini *B_2) Donec(pl int) *B_End {
	ini.Use()

	ini.ept.ConnMu.RLock()
	check(ini.ept.Conn[A][0].Send(pl))
	ini.ept.ConnMu.RUnlock()
	return &B_End{}
}

func (ini *B_Init) Run(f func(*B_1) *B_End) {
	ini.ept.CheckConnection()
	st1, err := ini.Init()
	check(err)
	f(st1)
}

/************ B API **********************************************************/

/*****************************************************************************/
/************ C API **********************************************************/
/*****************************************************************************/
type C_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *C_Init) Ept() *session.Endpoint {
	return ini.ept
}

func NewC(id, numS, numA int) (*C_Init, error) {
	session.RoleRange(id, numS)
	conn, err := session.NewConn([]session.ParamRole{{A, numA}})
	if err != nil {
		return nil, err
	}

	return &C_Init{session.LinearResource{}, session.NewEndpoint(id, numS, conn)}, nil
}

type C_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *C_Init) Init() (*C_1, error) {
	ini.Use()
	ini.ept.ConnMu.Lock()
	for n, l := range ini.ept.Conn {
		for i, l := range l {
			if l == nil {
				return nil, fmt.Errorf("Invalid connection for worker %s[%d]", n, i)
			}
		}
	}
	ini.ept.ConnMu.Unlock()
	return &C_1{session.LinearResource{}, ini.ept}, nil
}

type C_2 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *C_1) Measure() (int, *C_2) {
	ini.Use()
	var tmp int

	ini.ept.ConnMu.RLock()
	check(ini.ept.Conn[A][0].Recv(&tmp))
	ini.ept.ConnMu.RUnlock()
	return tmp, &C_2{session.LinearResource{}, ini.ept}
}

type C_End struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *C_2) Len(s int) *C_End {
	ini.Use()
	ini.ept.ConnMu.RLock()
	check(ini.ept.Conn[A][0].Send(s))
	ini.ept.ConnMu.RUnlock()
	return &C_End{}
}

func (ini *C_Init) Run(f func(*C_1) *C_End) {
	st1, err := ini.Init()
	check(err)
	f(st1)
}

/************ C API **********************************************************/
