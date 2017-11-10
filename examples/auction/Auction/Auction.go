package Auction

import (
	"fmt"
	"log"

	"github.com/nickng/scribble-go/runtime/session"
	"github.com/nickng/scribble-go/runtime/transport"
)

const (
	Auctioneer = "Auctioneer"
	Bidder     = "Bidder"
)

func NewBidder(id, nAuctioneer, nBidder int) (*Bidder_1Ton_Init, error) {
	if id > nBidder || id < 1 {
		return nil, fmt.Errorf("'Bidder' ID not in range [1, %d]", nBidder)
	}
	if nAuctioneer < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'Auctioneer': %d", nAuctioneer)
	}
	conn := make(map[string][]transport.Channel)
	conn[Auctioneer] = make([]transport.Channel, nAuctioneer)

	return &Bidder_1Ton_Init{ept: session.NewEndpoint(id, nBidder, conn)}, nil
}

type Bidder_1Ton_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *Bidder_1Ton_Init) Ept() *session.Endpoint {
	return ini.ept
}

func (ini *Bidder_1Ton_Init) Init() (*Bidder_1Ton_1, error) {
	ini.Use()

	ini.ept.ConnMu.Lock()
	defer ini.ept.ConnMu.Unlock()
	for i, conn := range ini.ept.Conn[Auctioneer] {
		if conn == nil {
			return nil, fmt.Errorf("invalid connection from 'Bidder[%d]' to 'Auctioneer[%d]'", ini.ept.Id, i)
		}
	}
	return &Bidder_1Ton_1{ept: ini.ept}, nil
}

func (ini *Bidder_1Ton_Init) Run(fn func(*Bidder_1Ton_1) *Bidder_1Ton_End) {
	ini.ept.CheckConnection()

	st, err := ini.Init()
	if err != nil {
		log.Fatalf("failed to initialise the session: %v", err)
	}
	fn(st)
}

type Bidder_1Ton_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (b1 *Bidder_1Ton_1) SendAll(args []int) *Bidder_1Ton_2 {
	if want, got := len(args), len(b1.ept.Conn[Auctioneer]); want != got {
		log.Fatalf("Sending wrong number of arguments in 'b1': %d != %d", want, got)
	}
	b1.Use()

	b1.ept.ConnMu.RLock()
	for i, arg := range args {
		b1.ept.Conn[Auctioneer][i].Send(arg)
	}
	b1.ept.ConnMu.RUnlock()
	return &Bidder_1Ton_2{ept: b1.ept}
}

type Bidder_1Ton_2 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (b2 *Bidder_1Ton_2) RecvAll() ([]int, *Bidder_1Ton_3) {
	b2.Use()

	res := make([]int, len(b2.ept.Conn[Auctioneer]))
	b2.ept.ConnMu.RLock()
	for i, conn := range b2.ept.Conn[Auctioneer] {
		err := conn.Recv(&res[i])
		if err != nil {
			log.Fatalf("Wrong value from 'Auctioneer[%d]' at 'Bidder[%d]': %v", i, b2.ept.Id, err)
		}
	}
	b2.ept.ConnMu.RUnlock()
	return res, &Bidder_1Ton_3{ept: b2.ept}
}

type Bidder_1Ton_3 struct {
	session.LinearResource
	ept *session.Endpoint
}

type IntOrBool interface {
	_IntOrBool()
}

type Int int

func (Int) _IntOrBool() {}

type Bool bool

func (Bool) _IntOrBool() {}

const (
	tagInt = iota + 1
	tagBool
)

func (b3 *Bidder_1Ton_3) SendAll(args []IntOrBool) *Bidder_1Ton_4 {
	if want, got := len(args), len(b3.ept.Conn[Auctioneer]); want != got {
		log.Fatalf("Sending wrong number of arguments in 'b3': %d != %d", want, got)
	}
	b3.Use()

	b3.ept.ConnMu.RLock()
	for i, arg := range args {
		switch arg.(type) {
		case Int:
			b3.ept.Conn[Auctioneer][i].Send(tagInt)
		case Bool:
			b3.ept.Conn[Auctioneer][i].Send(tagBool)
		}
		b3.ept.Conn[Auctioneer][i].Send(arg)
	}
	b3.ept.ConnMu.RUnlock()
	return &Bidder_1Ton_4{ept: b3.ept,
		IntChan: make(chan int, 1), StringChan: make(chan string, 1)} // Channel must be created at state creation.
}

const (
	tagint = iota + 1
	tagstring
)

type Bidder_1Ton_4 struct {
	session.LinearResource
	ept *session.Endpoint

	IntChan    chan int
	StringChan chan string
}

func (b4 *Bidder_1Ton_4) RecvAll() *Bidder_1Ton_4_Select {
	b4.ept.ConnMu.RLock()
	for i, conn := range b4.ept.Conn[Auctioneer] {
		var tag int
		err := conn.Recv(&tag)
		if err != nil {
			log.Fatalf("Wrong value from 'Auctioneer[%d]' at 'Bidder[%d]': %v", i, b4.ept.Id, err)
		}
		switch tag {
		case tagint:
			var v int
			err = conn.Recv(&v)
			if err != nil {
				log.Fatalf("Wrong value from 'Auctioneer[%d]' at 'Bidder[%d]': %v", i, b4.ept.Id, err)
			}
			b4.IntChan <- v
			close(b4.StringChan)
		case tagstring:
			var v string
			err = conn.Recv(&v)
			if err != nil {
				log.Fatalf("Wrong value from 'Auctioneer[%d]' at 'Bidder[%d]': %v", i, b4.ept.Id, err)
			}
			close(b4.IntChan)
			b4.StringChan <- v
		}
	}
	b4.ept.ConnMu.RUnlock()

	return &Bidder_1Ton_4_Select{ept: b4.ept,
		IntChan: b4.IntChan, StringChan: b4.StringChan}
}

type Bidder_1Ton_4_Select struct {
	// Only 'used' when method selected.
	session.LinearResource
	ept *session.Endpoint

	IntChan    chan int
	StringChan chan string
}

func (b4 *Bidder_1Ton_4_Select) Int(arg *int) <-chan *Bidder_1Ton_3 {
	if v, ok := <-b4.IntChan; ok {
		b4.Use()
		*arg = v

		ch := make(chan *Bidder_1Ton_3, 1)
		ch <- &Bidder_1Ton_3{ept: b4.ept}
		return ch
	}
	return nil
}

func (b4 *Bidder_1Ton_4_Select) String(arg *string) <-chan *Bidder_1Ton_End {
	if v, ok := <-b4.StringChan; ok {
		b4.Use()
		*arg = v

		ch := make(chan *Bidder_1Ton_End, 1)
		ch <- &Bidder_1Ton_End{}
		return ch
	}
	return nil
}

type Bidder_1Ton_End struct {
}

func NewAuctioneer(id, nAuctioneer, nBidder int) (*Auctioneer_1To1_Init, error) {
	if id > nAuctioneer || id < 1 {
		return nil, fmt.Errorf("'Auctioneer' ID not in range [1, %d]", nAuctioneer)
	}
	if nBidder < 1 {
		return nil, fmt.Errorf("Wrong number of participants of role 'Bidder': %d", nBidder)
	}
	conn := make(map[string][]transport.Channel)
	conn[Bidder] = make([]transport.Channel, nBidder)

	return &Auctioneer_1To1_Init{ept: session.NewEndpoint(id, nAuctioneer, conn)}, nil
}

type Auctioneer_1To1_Init struct {
	session.LinearResource
	ept *session.Endpoint
}

func (ini *Auctioneer_1To1_Init) Ept() *session.Endpoint {
	return ini.ept
}

func (ini *Auctioneer_1To1_Init) Init() (*Auctioneer_1To1_1, error) {
	ini.Use()

	ini.ept.ConnMu.Lock()
	defer ini.ept.ConnMu.Unlock()
	for i, conn := range ini.ept.Conn[Bidder] {
		if conn == nil { // ini.ept.Conn[Bidder][i]
			return nil, fmt.Errorf("invalid connection from 'Bidder[%d]' to 'Auctioneer[%d]'", ini.ept.Id, i)
		}
	}
	return &Auctioneer_1To1_1{ept: ini.ept}, nil
}

func (ini *Auctioneer_1To1_Init) Run(fn func(*Auctioneer_1To1_1) *Auctioneer_1To1_End) {
	ini.ept.CheckConnection()

	st, err := ini.Init()
	if err != nil {
		log.Fatalf("failed to initialise the session: %v", err)
	}
	fn(st)
}

type Auctioneer_1To1_1 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (a1 *Auctioneer_1To1_1) RecvAll() ([]int, *Auctioneer_1To1_2) {
	a1.Use()

	res := make([]int, len(a1.ept.Conn[Bidder]))
	a1.ept.ConnMu.RLock()
	for i, conn := range a1.ept.Conn[Bidder] {
		err := conn.Recv(&res[i])
		if err != nil {
			log.Fatalf("Wrong value from 'Bidder[%d]' at 'Auctioneer[%d]': %v", i, a1.ept.Id, err)
		}
	}
	a1.ept.ConnMu.RUnlock()
	return res, &Auctioneer_1To1_2{ept: a1.ept}
}

type Auctioneer_1To1_2 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (a2 *Auctioneer_1To1_2) SendAll(args []int) *Auctioneer_1To1_3 {
	if want, got := len(args), len(a2.ept.Conn[Bidder]); want != got {
		log.Fatal("Sending wrong number of arguments in 'a2': %d != %d", want, got)
	}
	a2.Use()

	a2.ept.ConnMu.RLock()
	for i, arg := range args {
		a2.ept.Conn[Bidder][i].Send(arg)
	}
	a2.ept.ConnMu.RUnlock()
	return &Auctioneer_1To1_3{ept: a2.ept}
}

type Auctioneer_1To1_3 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (a3 *Auctioneer_1To1_3) RecvAll() ([]IntOrBool, *Auctioneer_1To1_4) {
	a3.Use()

	// Receive non-diverging choice.
	res := make([]IntOrBool, len(a3.ept.Conn[Bidder]))
	a3.ept.ConnMu.RLock()
	for i, conn := range a3.ept.Conn[Bidder] {
		var tag int
		err := conn.Recv(&tag)
		if err != nil {
			log.Fatalf("Wrong value from 'Bidder[%d]' at 'Auctioneer[%d]': %v", i, a3.ept.Id, err)
		}
		switch tag {
		case tagInt:
			var v Int
			err = conn.Recv(&v)
			if err != nil {
				log.Fatalf("Wrong value from 'Bidder[%d]' at 'Auctioneer[%d]': %v", i, a3.ept.Id, err)
			}
			res[i] = v
		case tagBool:
			var v Bool
			err = conn.Recv(&v)
			if err != nil {
				log.Fatalf("Wrong value from 'Bidder[%d]' at 'Auctioneer[%d]': %v", i, a3.ept.Id, err)
			}
			res[i] = v
		}
	}
	a3.ept.ConnMu.RUnlock()
	return res, &Auctioneer_1To1_4{ept: a3.ept}
}

type Auctioneer_1To1_4 struct {
	session.LinearResource
	ept *session.Endpoint
}

func (a4 *Auctioneer_1To1_4) SendAll_int(args []int) *Auctioneer_1To1_3 {
	if want, got := len(args), len(a4.ept.Conn[Bidder]); want != got {
		log.Fatalf("Sending wrong number of arguments in 'a4': %d != %d", want, got)
	}
	a4.Use()

	a4.ept.ConnMu.RLock()
	for i, arg := range args {
		a4.ept.Conn[Bidder][i].Send(tagint)
		a4.ept.Conn[Bidder][i].Send(arg)
	}
	return &Auctioneer_1To1_3{ept: a4.ept}
}

func (a4 *Auctioneer_1To1_4) SendAll_string(args []string) *Auctioneer_1To1_End {
	if want, got := len(args), len(a4.ept.Conn[Bidder]); want != got {
		log.Fatalf("Sending wrong number of arguments in 'a4': %d != %d", want, got)
	}
	a4.Use()

	a4.ept.ConnMu.RLock()
	for i, arg := range args {
		a4.ept.Conn[Bidder][i].Send(tagstring)
		a4.ept.Conn[Bidder][i].Send(arg)
	}
	a4.ept.ConnMu.RUnlock()
	return &Auctioneer_1To1_End{}
}

type Auctioneer_1To1_End struct {
}
