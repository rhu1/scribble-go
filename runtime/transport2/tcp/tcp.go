// Package tcp provides a TCP transport implementation.
package tcp

import (
	"log"

	"net"
	"strconv"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

// TcpListener is a server-side TCP listener
// which implements transport.ScribListener.
// It wraps the standard library "net".Listener.
type TcpListener struct {
	ln net.Listener
}

// Accept waits for and accepts incoming connection.
func (ss *TcpListener) Accept() (transport2.BinChannel, error) {
	conn, err := ss.ln.Accept()
	if err != nil {
		log.Fatalf("cannot accept %s: %v", ss.Addr().String(), err)
	}
	c := TcpChannel{conn: conn}
	return &c, err
}

// Addr returns the listener's network address.
func (ss *TcpListener) Addr() net.Addr {
	return ss.ln.Addr()
}

// Close terminates a listening TCP listener.
func (ss *TcpListener) Close() error {
	return ss.ln.Close()
}

// Listen creates a new TCP listener at port port.
// The running user needs to be a privileged user for port <= 1024.
func Listen(port int) (*TcpListener, error) {
	ln, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("cannot listen at :%d: %v", port, err)
	}
	ss := TcpListener{ln: ln}
	return &ss, err
}

// BListen creates a new TCP listener at port.
//
// FIXME HACK -- simply replace existing Listen signature with this one?
func BListen(port int) (transport2.ScribListener, error) {
	return Listen(port)
}

// Dial uses the given port to establish a TCP connection.
func Dial(host string, port int) (transport2.BinChannel, error) {
	conn, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("cannot connect to %s:%d: %v", host, port, err)
	}
	c := TcpChannel{conn: conn}
	return &c, err
}

// TcpChannel is a binary channel over TCP.
// It is implemented as a wrapper to standard library "net".Conn.
type TcpChannel struct {
	conn net.Conn
}

/*func (c *TcpChannel) GetConn() net.Conn {
	return c.conn
}*/
func (c *TcpChannel) GetReader() io.Reader {
	return c.conn
}

func (c *TcpChannel) GetWriter() io.Writer {
	return c.conn
}

// Close terminates a TCP channel.
func (c *TcpChannel) Close() error {
	return c.conn.Close()
}
