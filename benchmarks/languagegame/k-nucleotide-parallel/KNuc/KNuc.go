package KNuc

import (
	"fmt"
	"github.com/nickng/scribble-go/runtime/session"
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

	return &A_Init{session.LinearResource{}, &session.Endpoint{id, numA, conn}}, nil
}

type A_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *A_Init) Init() (*A_1, error) {
	ini.Use()
	for _, v := range ini.ept.Conn {
		for _, c := range v {
			for c != nil {
			}
		}
	}
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
	for i, c := range ini.ept.Conn[S] {
		check(c.Send(pl[i]))
	}
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
	for i, c := range ini.ept.Conn[B] {
		check(c.Send(pl[i]))
	}
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

	for i, c := range ini.ept.Conn[S] {
		check(c.Recv(&tmp))
		pl[i] = tmp
	}
	return pl, &A_4{session.LinearResource{}, ini.ept}
}

type A_End struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *A_4) RecvB() ([]string, *A_End) {
	var tmp string
	pl := make([]string, len(ini.ept.Conn[B]))

	for i, c := range ini.ept.Conn[B] {
		check(c.Recv(&tmp))
		pl[i] = tmp
	}
	return pl, &A_End{}
}

func (ini *A_Init) Run(f func(*A_1) *A_End) {
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

	return &B_Init{session.LinearResource{}, &session.Endpoint{id, numB, conn}}, nil
}

type B_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *B_Init) Init() (*B_1, error) {
	ini.Use()
	for n, l := range ini.ept.Conn {
		for i, l := range l {
			if l == nil {
				return nil, fmt.Errorf("Invalid connection for worker %s[%d]", n, i)
			}
		}
	}
	return &B_1{session.LinearResource{}, ini.ept}, nil
}

type B_2 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *B_1) Recv_BA() (string, *B_2) {
	ini.Use()
	var tmp string

	check(ini.ept.Conn[A][0].Recv(&tmp))
	return tmp, &B_2{session.LinearResource{}, ini.ept}
}

type B_End struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *B_2) Send_BA(pl string) *B_End {
	ini.Use()

	check(ini.ept.Conn[A][0].Send(pl))
	return &B_End{}
}

func (ini *B_Init) Run(f func(*B_1) *B_End) {
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

	return &S_Init{session.LinearResource{}, &session.Endpoint{id, numS, conn}}, nil
}

type S_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *S_Init) Init() (*S_1, error) {
	ini.Use()
	for n, l := range ini.ept.Conn {
		for i, l := range l {
			if l == nil {
				return nil, fmt.Errorf("Invalid connection for worker %s[%d]", n, i)
			}
		}
	}
	return &S_1{session.LinearResource{}, ini.ept}, nil
}

type S_2 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *S_1) Recv_SA() (int, *S_2) {
	ini.Use()
	var tmp int

	check(ini.ept.Conn[A][0].Recv(&tmp))
	return tmp, &S_2{session.LinearResource{}, ini.ept}
}

type S_End struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *S_2) Send_SA(s string) *S_End {
	ini.Use()
	check(ini.ept.Conn[A][0].Send(s))
	return &S_End{}
}

func (ini *S_Init) Run(f func(*S_1) *S_End) {
	st1, err := ini.Init()
	check(err)
	f(st1)
}

/************ S API **********************************************************/
