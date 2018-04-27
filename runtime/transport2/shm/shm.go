package shm


import (
	//"bufio"
	"errors"
	"fmt"
	"io"
	//"log"
	//"net"
	"strconv"
	"sync"
	//"time"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

var _ = fmt.Print

var _lock *sync.Mutex = new(sync.Mutex)
var _ports map[int]chan interface{} = make(map[int]chan interface{})

type ShmListener struct {
	port int	
	c chan interface{}
}

func (ss *ShmListener) Accept() (transport2.BinChannel, error) {
	rr := (<-ss.c).(chan interface{})
	ww := (<-ss.c).(chan interface{})
	r := (<-ss.c).(chan []byte)
	w := (<-ss.c).(chan []byte)
	return newShmChannel(rr, ww, r, w), nil
}

func (ss *ShmListener) Close() error {
	_lock.Lock()
	defer _lock.Unlock()
	delete(_ports, ss.port)
	return nil
}

func Listen(port int)	(*ShmListener, error) {
	_lock.Lock()
	defer _lock.Unlock()
	if _, has := _ports[port]; has {
		return nil, errors.New("[shm] Port in use: " + strconv.Itoa(port))
	}
	c := make(chan interface{})
	_ports[port] = c
	return &ShmListener{port: port, c: c}, nil
}

// FIXME: separate Shm and GobShm
func Dial(host string, port int) (transport2.BinChannel, error) {
	_lock.Lock()
	p, has := _ports[port]
	_lock.Unlock()  // Subsequent comms can race with listener Close -- same as distributed case (and similar to unopen port error)
	if !has {
		return nil, errors.New("[shm] Port not open: " + strconv.Itoa(port))
	}

	rr := make(chan interface{}, 1024)
	ww := make(chan interface{}, 1024)
	r := make(chan []byte, 1024)
	w := make(chan []byte, 1024)
	p <- ww
	p <- rr
	p <- w
	p <- r
	return newShmChannel(rr, ww, r, w), nil
}

type ShmChannel struct {
	rr chan interface{}
	ww chan interface{}
	r *ShmReader
	w *ShmWriter
}

func (c *ShmChannel) WritePointer(m interface{}) {
	c.ww <- m	
}

func (c *ShmChannel) ReadPointer(m *interface{}) {
	*m = <-c.rr	
}

func newShmChannel(rr chan interface{}, ww chan interface{}, r chan []byte, w chan[] byte) *ShmChannel {
	buff := make([]byte, 1024)
	return &ShmChannel{rr: rr, ww: ww, r: &ShmReader{c:r, overflow:buff}, w: &ShmWriter{c:w}}	
}

// FIXME: use a *bytes.Buffer (its a Reader/Writer)
func (c *ShmChannel) GetReader() io.Reader {
	return c.r
}

func (c *ShmChannel) GetWriter() io.Writer {
	return c.w
}

func (c *ShmChannel) Close() error {
	return nil
}

type ShmReader struct {
	c chan []byte
	overflow []byte
}

//q.Items= q.Items[1:]
//append(q.Items,item)	
func (r *ShmReader) Read(p []byte) (n int, err error) {
	bs := <-r.c
	fmt.Println("R read: ", len(bs))
	copy(p, bs)				
	return len(bs), nil

	/*d := p
	for {
		n := len(d)  // Remaining to read
		fmt.Println("R remaining: ", n)
		avail := len(r.overflow)
		fmt.Println("R avail: ", avail)
		if avail > 0 {
			copy(p, r.overflow)
			if avail >= n {
					r.overflow = r.overflow[n:]  // Mem leak?  Re-slicing doesn't free up?
					fmt.Println("R done: ", len(r.overflow))
					return avail, nil
			}
			r.overflow = r.overflow[avail:]
			d = p[avail:]
		}
		bs := <-r.c
		fmt.Println("R read: ", len(bs))
		r.overflow = append(r.overflow, bs...)
	}*/
}

type ShmWriter struct {
	c chan []byte
}

func (w *ShmWriter) Write(p []byte) (n int, err error) {
	w.c <- p	
	return len(p), nil
}



/*var (
	ErrCloseUnfinishedConn = errors.New("transport/tcp: closing connection with unread data")
)

// SerialisationError is the kind of error where a value
// cannot be sent due to serialisation failure.
type SerialisationError struct {
	cause error
}

func (e SerialisationError) Error() string {
	return fmt.Sprintf("transport/tcp send: serialisation failed: %v", e.cause)
}

// DeserialisationError is the kind of error where a value
// cannot be received due to deserialisation failure.
type DeserialisationError struct {
	cause error
}

func (e DeserialisationError) Error() string {
	return fmt.Sprintf("transport/tcp recv: deserisalisation failed: %v", e.cause)
}

// ConnCfg is a connection configuration, contains
// the details required to establish a connection.
type ConnCfg struct {
	Host string
	Port string

	// DelimMeth specifies delimiter implementation.
	DelimMeth     DelimitMethod
	SerialiseMeth SerialiseMethod

	// retryWait specifies the time to wait before retrying connection.
	retryWait time.Duration
}

// NewConnection is a convenient wrapper for a TCP connection
// and can be used as either server-side or client-side.
func NewConnection(host, port string) ConnCfg {
	return ConnCfg{Host: host, Port: port}
}

func Listen(port string) ConnCfg {
	return NewConnection("__dummy", port)
}

func NewAcceptor(port string) ConnCfg {
	return NewConnection("__dummy", port)
}

func NewRequestor(host string, port string) ConnCfg {
	return NewConnection(host, port)
}

func NewConnectionWithRetry(host, port string, retryWait time.Duration) ConnCfg {
	return ConnCfg{Host: host, Port: port, retryWait: retryWait}
}

// Accept listens for and accepts connection from a TCP socket using
// details from cfg, and returns the TCP stream as a ReadWriteCloser.
//
// Accept blocks while waiting for connection to be accepted.
func (cfg ConnCfg) Accept() transport.Channel {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))  // FIXME: port should be opened on Listen, not Accept
	if err != nil {
		log.Fatalf("cannot listen at :%s: %v", cfg.Port, err)
	}
	conn, err := ln.Accept()
	if err != nil {
		log.Fatalf("cannot accept connection at :%s: %v", cfg.Port, err)
	}
	return cfg.newConn(conn)
}

// Connect establishes a connection with a TCP socket using details
// from cfg, and returns the TCP stream as a ReadWriteCloser.
func (cfg ConnCfg) Connect() transport.Channel {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		if cfg.retryWait > 0 {
			time.Sleep(cfg.retryWait)
			cfg.retryWait = 0
			return cfg.Connect()
		}
		log.Fatalf("cannot connect to %s: %v", addr, err)
	}
	return cfg.newConn(conn)
}

func (cfg ConnCfg) Request() transport.Channel {
	return cfg.Connect()
}

func (cfg ConnCfg) newConn(rwc net.Conn) *Conn {
	c := &Conn{
		rwc: rwc,
	}
	c.rdMu.Lock()
	c.bufr = newReader(c.rwc)
	c.dec = NewDeserialiser(NewDelimReader(c, cfg.DelimMeth), cfg.SerialiseMeth)
	c.rdMu.Unlock()

	c.wtMu.Lock()
	c.bufw = newWriter(c.rwc)
	c.enc = NewSerialiser(NewDelimWriter(c, cfg.DelimMeth), cfg.SerialiseMeth)
	c.wtMu.Unlock()
	return c
}

// Conn is a connected TCP stream/connection, and wraps a net.Conn created
// by either Accept or Connect.
//
// Conn implements ReadWriteCloser and can be used as is, more fine-grained
// message formatting control (such as delimited multi messages) should use
// NewSizedReader/SizedWriter (message with size prefix) or
// NewDelimReader/DelimWriter (delimited message).
type Conn struct {
	// rwc is the real TCP connection.
	rwc net.Conn

	// guards the read buffer and the decoder
	rdMu sync.Mutex

	bufr *bufio.Reader // bufr is a buffered stream to the TCP connection.
	dec  deserialiser  // dec is a serialisation decoder for messages from rwc.

	// guards the write buffer and the encoder
	wtMu sync.Mutex

	bufw *bufio.Writer // bufw is a buffered stream to the TCP connection.
	enc  serialiser    // enc is a serialisation encoder for messages to rwc.
}

// newReader returns a fresh buffered Reader.
func newReader(r io.Reader) *bufio.Reader {
	// TODO(nickng): use sync.Pool to reduce allocation per new connection.
	return bufio.NewReader(r)
}

// newWriter returns a fresh buffered Writer.
func newWriter(w io.Writer) *bufio.Writer {
	// TODO(nickng): use sync.Pool to reduce allocation per new connection.
	return bufio.NewWriter(w)
}

// Read reads data into p. It returns the number of bytes read into p. The
// bytes are taken from at most one Read on the underlying Reader, hence n
// may be less than len(p). At EOF, the count will be zero and err will be
// io.EOF.
//
// The underlying implementation is a *bufio.Reader.
func (c *Conn) Read(p []byte) (n int, err error) {
	return c.bufr.Read(p)
}

// Writer writes the content of p into the underlying stream. It returns
// the number of bytes written. If n < len(p), it also returns an error
// explaining why the write is short.
//
// The underlying implementation is a *bufio.Writer, and data will be
// flushed whenever Write is called.
func (c *Conn) Write(p []byte) (n int, err error) {
	n, err = c.bufw.Write(p)
	c.bufw.Flush()
	return n, err
}

func (c *Conn) ScribWrite(bs []byte) error {
	return nil
}

func (c *Conn) ScribRead(bs *[]byte) error {
	return nil	
}

// Close closes the underlying TCP connection.
func (c *Conn) Close() error {
	if c.bufw.Available() > 0 {
		c.bufw.Flush()
	}
	if c.bufr.Buffered() > 0 {
		c.rwc.Close()
		return ErrCloseUnfinishedConn
	}
	return c.rwc.Close()
}

// Send serialises values val then sends the serialised
// values to the underlying stream of connection c.
func (c *Conn) Send(val interface{}) error {
	c.wtMu.Lock()
	defer c.wtMu.Unlock()
	if err := c.enc.Encode(val); err != nil {
		return SerialisationError{cause: err}
	}
	return nil
}

// Recv receives values from the underlying stream then deserialises and
// writes the deserialised values to the pointer addresses specified by ptr.
func (c *Conn) Recv(ptr interface{}) error {
	c.rdMu.Lock()
	defer c.rdMu.Unlock()
	if err := c.dec.Decode(ptr); err != nil {
		return DeserialisationError{cause: err}
	}
	return nil
}
*/