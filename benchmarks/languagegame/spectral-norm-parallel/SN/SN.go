package SN

import (
	"fmt"
	"github.com/nickng/scribble-go/runtime/session"
	"log"
)

const A = "A"
const B = "B"

const LTimes = 1
const LEnd = 2

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

type A_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

type A_2 struct {
	session.LinearResource
	ept *session.Endpoint
}

type A_3 struct {
	session.LinearResource
	ept *session.Endpoint
}

type A_4 struct {
	session.LinearResource
	ept *session.Endpoint
}

type A_5 struct {
	session.LinearResource
	ept *session.Endpoint
}

type A_6 struct {
	session.LinearResource
	ept *session.Endpoint
}

type A_7 struct {
	session.LinearResource
	ept *session.Endpoint
}

type A_8 struct {
	session.LinearResource
	ept *session.Endpoint
}

type A_End struct {
}

func NewA(id, numA, numB int) (*A_Init, error) {
	session.RoleRange(id, numA)
	conn, err := session.NewConn([]session.ParamRole{{B, numB}})
	if err != nil {
		return nil, err
	}

	return &A_Init{session.LinearResource{}, &session.Endpoint{id, numA, conn}}, nil
}

func (ini *A_Init) Init() (*A_1, error) {
	ini.Use()
	for n, _ := range ini.ept.Conn {
		for j, _ := range ini.ept.Conn[n] {
			for ini.ept.Conn[n][j] == nil {
			}
		}
	}
	return &A_1{session.LinearResource{}, ini.ept}, nil
}

func (ini *A_1) SendTimes(pl []int) *A_2 {
	ini.Use()
	if len(pl) != len(ini.ept.Conn[B]) {
		log.Panicf("Incorrect number of arguments to role 'A' SendS")
	}
	for i, c := range ini.ept.Conn[B] {
		check(c.Send(LTimes))
		check(c.Send(pl[i]))
	}
	return &A_2{session.LinearResource{}, ini.ept}
}

func (ini *A_1) SendEnd(pl []int) *A_End {
	ini.Use()
	if len(pl) != len(ini.ept.Conn[B]) {
		log.Panicf("Incorrect number of arguments to role 'A' SendS")
	}
	for i, c := range ini.ept.Conn[B] {
		check(c.Send(LEnd))
		check(c.Send(pl[i]))
	}
	return &A_End{}
}

func (ini *A_2) RecvDone() ([]int, *A_3) {
	var tmp int
	pl := make([]int, len(ini.ept.Conn[B]))

	for i, c := range ini.ept.Conn[B] {
		check(c.Recv(&tmp))
		pl[i] = tmp
	}
	return pl, &A_3{session.LinearResource{}, ini.ept}
}

func (ini *A_3) SendNext(pl []int) *A_4 {
	ini.Use()
	if len(pl) != len(ini.ept.Conn[B]) {
		log.Panicf("Incorrect number of arguments to role 'A' SendS")
	}
	for i, c := range ini.ept.Conn[B] {
		check(c.Send(pl[i]))
	}
	return &A_4{session.LinearResource{}, ini.ept}
}

func (ini *A_4) RecvDone() ([]int, *A_5) {
	var tmp int
	pl := make([]int, len(ini.ept.Conn[B]))

	for i, c := range ini.ept.Conn[B] {
		check(c.Recv(&tmp))
		pl[i] = tmp
	}
	return pl, &A_5{session.LinearResource{}, ini.ept}
}

func (ini *A_5) SendTimesTr(pl []int) *A_6 {
	ini.Use()
	if len(pl) != len(ini.ept.Conn[B]) {
		log.Panicf("Incorrect number of arguments to role 'A' SendS")
	}
	for i, c := range ini.ept.Conn[B] {
		check(c.Send(pl[i]))
	}
	return &A_6{session.LinearResource{}, ini.ept}
}

func (ini *A_6) RecvDone() ([]int, *A_7) {
	var tmp int
	pl := make([]int, len(ini.ept.Conn[B]))

	for i, c := range ini.ept.Conn[B] {
		check(c.Recv(&tmp))
		pl[i] = tmp
	}
	return pl, &A_7{session.LinearResource{}, ini.ept}
}

func (ini *A_7) SendNext(pl []int) *A_8 {
	ini.Use()
	if len(pl) != len(ini.ept.Conn[B]) {
		log.Panicf("Incorrect number of arguments to role 'A' SendS")
	}
	for i, c := range ini.ept.Conn[B] {
		check(c.Send(pl[i]))
	}
	return &A_8{session.LinearResource{}, ini.ept}
}

func (ini *A_8) RecvDone() ([]int, *A_1) {
	var tmp int
	pl := make([]int, len(ini.ept.Conn[B]))

	for i, c := range ini.ept.Conn[B] {
		check(c.Recv(&tmp))
		pl[i] = tmp
	}
	return pl, &A_1{session.LinearResource{}, ini.ept}
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
	ept    *session.Endpoint
	data   chan int
	ctimes chan chan *B_2
	cend   chan chan *B_End
}

type B_2 struct {
	session.LinearResource
	ept *session.Endpoint
}

type B_3 struct {
	session.LinearResource
	ept *session.Endpoint
}

type B_4 struct {
	session.LinearResource
	ept *session.Endpoint
}

type B_5 struct {
	session.LinearResource
	ept *session.Endpoint
}

type B_6 struct {
	session.LinearResource
	ept *session.Endpoint
}

type B_7 struct {
	session.LinearResource
	ept *session.Endpoint
}

type B_8 struct {
	session.LinearResource
	ept *session.Endpoint
}

type B_End struct {
}

func (st1 *B_1) timesOrEnd(data chan int, st2 chan chan *B_2, st3 chan chan *B_End) {
	st1.Use()
	var lbl int
	var res int

	conn := st1.ept.Conn[A][0]

	err := conn.Recv(&lbl)
	if err != nil {
		log.Fatalf("wrong label from server at %d: %s", st1.ept.Id, err)
	}

	if lbl == LTimes {
		ch := make(chan *B_2, 1)
		err = conn.Recv(&res)
		if err != nil {
			log.Fatalf("wrong value(times) from server at %d: %s", st1.ept.Id, err)
		}
		data <- res
		ch <- &B_2{session.LinearResource{}, st1.ept}
		st2 <- ch
		close(st3)
		return
	}
	if lbl == LEnd {
		ch := make(chan *B_End, 1)
		err = conn.Recv(&res)
		if err != nil {
			log.Fatalf("wrong value(end) from server at %d: %s", st1.ept.Id, err)
		}
		data <- res
		ch <- &B_End{}
		st3 <- ch
		close(st2)
		return
	}
	log.Fatalf("wrong value(unknown) from server at %d: %d", st1.ept.Id, lbl)
}

func mkB_1(ept *session.Endpoint) *B_1 {
	ch_res := make(chan int, 1)
	ch_st1 := make(chan chan *B_2, 1)
	ch_st2 := make(chan chan *B_End, 1)
	st1 := &B_1{session.LinearResource{}, ept, ch_res, ch_st1, ch_st2}
	go st1.timesOrEnd(ch_res, ch_st1, ch_st2)
	return st1
}

func (ini *B_Init) Init() (*B_1, error) {
	ini.Use()
	for n, l := range ini.ept.Conn {
		for i, c := range l {
			if c == nil {
				return nil, fmt.Errorf("Invalid connection for worker %s[%d]", n, i)
			}
		}
	}
	return mkB_1(ini.ept), nil
}

func (st1 *B_1) RecvTimes(res *int) <-chan *B_2 {
	ch, selected := <-st1.ctimes
	if !selected {
		return nil
	}
	*res = <-st1.data
	return ch
}

func (st1 *B_1) RecvEnd(res *int) <-chan *B_End {
	ch, selected := <-st1.cend
	if !selected {
		return nil
	}
	*res = <-st1.data
	return ch
}

func (ini *B_2) SendDone(pl int) *B_3 {
	ini.Use()

	check(ini.ept.Conn[A][0].Send(pl))
	return &B_3{session.LinearResource{}, ini.ept}
}

func (ini *B_3) RecvNext() (int, *B_4) {
	ini.Use()
	var tmp int

	check(ini.ept.Conn[A][0].Recv(&tmp))
	return tmp, &B_4{session.LinearResource{}, ini.ept}
}

func (ini *B_4) SendDone(pl int) *B_5 {
	ini.Use()

	check(ini.ept.Conn[A][0].Send(pl))
	return &B_5{session.LinearResource{}, ini.ept}
}

func (ini *B_5) RecvTimesTr() (int, *B_6) {
	ini.Use()
	var tmp int

	check(ini.ept.Conn[A][0].Recv(&tmp))
	return tmp, &B_6{session.LinearResource{}, ini.ept}
}

func (ini *B_6) SendDone(pl int) *B_7 {
	ini.Use()

	check(ini.ept.Conn[A][0].Send(pl))
	return &B_7{session.LinearResource{}, ini.ept}
}

func (ini *B_7) RecvNext() (int, *B_8) {
	ini.Use()
	var tmp int

	check(ini.ept.Conn[A][0].Recv(&tmp))
	return tmp, &B_8{session.LinearResource{}, ini.ept}
}

func (ini *B_8) SendDone(pl int) *B_1 {
	ini.Use()

	check(ini.ept.Conn[A][0].Send(pl))
	return mkB_1(ini.ept)
}

func (ini *B_Init) Run(f func(*B_1) *B_End) {
	st1, err := ini.Init()
	check(err)
	f(st1)
}

/************ B API **********************************************************/
