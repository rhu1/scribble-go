package shm

import (
	"fmt"
	"io"
	"sync"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

// sharedChan is an initial shared channel.
// The design is based on net.Pipe
type sharedChan struct {
	cb1 chan []byte
	cb2 chan []byte
	cn1 chan int
	cn2 chan int
	cp1 chan interface{}
	cp2 chan interface{}
}

func newSharedChan() *sharedChan {
	return &sharedChan{
		cb1: make(chan []byte),
		cb2: make(chan []byte),
		cn1: make(chan int),
		cn2: make(chan int),
		cp1: make(chan interface{}),
		cp2: make(chan interface{}),
	}
}

// Channel is a shared memory binary channel.
type Channel struct {
	wrMu sync.Mutex // Ensures atomic write

	rdRx  <-chan []byte
	rdTx  chan<- int
	rdPtr <-chan interface{}

	wrTx  chan<- []byte
	wrRx  <-chan int
	wrPtr chan<- interface{}
}

func (c *Channel) Read(b []byte) (n int, err error) {
	bw := <-c.rdRx
	nr := copy(b, bw)
	c.rdTx <- nr
	return nr, nil
}

func (c *Channel) Write(b []byte) (n int, err error) {
	c.wrMu.Lock()
	defer c.wrMu.Unlock()
	for once := true; once || len(b) > 0; once = false {
		c.wrTx <- b
		nw := <-c.wrRx
		b = b[nw:]
		n += nw
	}
	return n, nil
}

// Close channel terminates a channel.
func (c *Channel) Close() error {
	return nil
}

func (c *Channel) GetReader() io.Reader { return c }
func (c *Channel) GetWriter() io.Writer { return c }

// ReadPointer is for receiving pointer over an untyped channel.
func (c *Channel) ReadPointer(m *interface{}) {
	*m = <-c.rdPtr
}

// WritePointer is for sending a pointer over an untyped channel.
func (c *Channel) WritePointer(m interface{}) {
	c.wrPtr <- m
}

// Listener is a server-side shared memory listener
// which implements transport.ScribListener.
type Listener struct {
	port int
	ch   *sharedChan
}

func (ln *Listener) Accept() (transport2.BinChannel, error) {
	c := Channel{
		rdRx: ln.ch.cb1, rdTx: ln.ch.cn1, rdPtr: ln.ch.cp1,
		wrTx: ln.ch.cb2, wrRx: ln.ch.cn2, wrPtr: ln.ch.cp2,
	}
	return &c, nil
}

func (ln *Listener) Close() error {
	ports.mu.Lock()
	defer ports.mu.Unlock()
	delete(ports.chans, ln.port)
	return nil
}

// PortInUseError is the error used when Listen
// is called on a port which is already in use.
type PortInUseError struct {
	port int
}

func (e PortInUseError) Error() string {
	return fmt.Sprintf("cannot listen: shared memory port %d is in use", e.port)
}

type registry struct {
	mu    sync.Mutex
	chans map[int]*sharedChan
}

var ports *registry

func init() {
	ports = &registry{
		chans: make(map[int]*sharedChan),
	}
}

// Listen creates a new listener at with port as identifier.
func Listen(port int) (*Listener, error) {
	ports.mu.Lock()
	defer ports.mu.Unlock()
	if _, exists := ports.chans[port]; exists {
		return nil, PortInUseError{port}
	}
	shared := newSharedChan()
	ports.chans[port] = shared
	return &Listener{port: port, ch: shared}, nil
}

func Dial(_ string, port int) (transport2.BinChannel, error) {
	ports.mu.Lock()
	defer ports.mu.Unlock()
	ch, exists := ports.chans[port]
	if !exists {
		return nil, fmt.Errorf("shm: dial failed: port %d does not exist", port)
	}
	c := Channel{
		rdRx: ch.cb2, rdTx: ch.cn2, rdPtr: ch.cp2,
		wrTx: ch.cb1, wrRx: ch.cn1, wrPtr: ch.cp1,
	}
	return &c, nil
}
