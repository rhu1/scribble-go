package KNuc

import (
	"fmt"
	"github.com/rhu1/scribble-go-runtime/runtime/session"
	"log"
)

const A = "A"
const B = "B"
const S = "S"

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

func NewA(id, numA, numS, numB int) (*A_Init, error) {
	session.RoleRange(id, numA)
	conn, err := session.NewConn([]session.ParamRole{{S, numS}, {B, numB}})
	if err != nil {
		return nil, err
	}

	//return &A_Init{session.LinearResource{}, session.NewEndpoint(id, numA, conn)}, nil
	return &A_Init{session.LinearResource{}, session.NewEndpoint(id, conn)}, nil
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

func (ini *A_1) SendS(pl []int) *A_2 {
	ini.Use()
	if len(pl) != len(ini.ept.Conn[S]) {
		log.Panicf("Incorrect number of arguments to role 'A' SendS")
	}
	ini.ept.ConnMu.RLock()
	for i, c := range ini.ept.Conn[S] {
		check(c.Send(pl[i]))
	}
	ini.ept.ConnMu.RUnlock()
	return &A_2{session.LinearResource{}, ini.ept}
}

type A_3 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *A_2) SendB(pl []string) *A_3 {
	ini.Use()
	if len(pl) != len(ini.ept.Conn[B]) {
		log.Panicf("Incorrect number of arguments to role 'B' SendB")
	}
	ini.ept.ConnMu.RLock()
	for i, c := range ini.ept.Conn[B] {
		check(c.Send(pl[i]))
	}
	ini.ept.ConnMu.RUnlock()
	return &A_3{session.LinearResource{}, ini.ept}
}

type A_4 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *A_3) RecvS() ([]int, *A_4) {
	ini.Use()
	var tmp int
	pl := make([]int, len(ini.ept.Conn[S]))

	ini.ept.ConnMu.RLock()
	for i, c := range ini.ept.Conn[S] {
		check(c.Recv(&tmp))
		pl[i] = tmp
	}
	ini.ept.ConnMu.RUnlock()
	return pl, &A_4{session.LinearResource{}, ini.ept}
}

type A_End struct {
}

func (ini *A_4) RecvB() ([]string, *A_End) {
	var tmp string
	pl := make([]string, len(ini.ept.Conn[B]))

	ini.ept.ConnMu.RLock()
	for i, c := range ini.ept.Conn[B] {
		check(c.Recv(&tmp))
		pl[i] = tmp
	}
	ini.ept.ConnMu.RUnlock()
	return pl, &A_End{}
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

	//return &B_Init{session.LinearResource{}, session.NewEndpoint(id, numB, conn)}, nil
	return &B_Init{session.LinearResource{}, session.NewEndpoint(id, conn)}, nil
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
				return nil, fmt.Errorf("Invalid connection for worker %s[%d]", n, i)
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

func (ini *B_1) Recv_BA() (string, *B_2) {
	ini.Use()
	var tmp string

	ini.ept.ConnMu.RLock()
	check(ini.ept.Conn[A][0].Recv(&tmp))
	ini.ept.ConnMu.RUnlock()
	return tmp, &B_2{session.LinearResource{}, ini.ept}
}

type B_End struct {
}

func (ini *B_2) Send_BA(pl string) *B_End {
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
/************ S API **********************************************************/
/*****************************************************************************/
type S_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *S_Init) Ept() *session.Endpoint {
	return ini.ept
}

func NewS(id, numS, numA int) (*S_Init, error) {
	session.RoleRange(id, numS)
	conn, err := session.NewConn([]session.ParamRole{{A, numA}})
	if err != nil {
		return nil, err
	}

	//return &S_Init{session.LinearResource{}, session.NewEndpoint(id, numS, conn)}, nil
	return &S_Init{session.LinearResource{}, session.NewEndpoint(id, conn)}, nil
}

type S_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *S_Init) Init() (*S_1, error) {
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
	return &S_1{session.LinearResource{}, ini.ept}, nil
}

type S_2 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *S_1) Recv_SA() (int, *S_2) {
	ini.Use()
	var tmp int

	ini.ept.ConnMu.RLock()
	check(ini.ept.Conn[A][0].Recv(&tmp))
	ini.ept.ConnMu.RUnlock()
	return tmp, &S_2{session.LinearResource{}, ini.ept}
}

type S_End struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *S_2) Send_SA(s int) *S_End {
	ini.Use()
	ini.ept.ConnMu.RLock()
	check(ini.ept.Conn[A][0].Send(s))
	ini.ept.ConnMu.RUnlock()
	return &S_End{}
}

func (ini *S_Init) Run(f func(*S_1) *S_End) {
	st1, err := ini.Init()
	check(err)
	f(st1)
}

/************ S API **********************************************************/
