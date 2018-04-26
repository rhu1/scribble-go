package transport2

import (
	//"io"
	"net"
)

/*type Transport interface {
	Listen(int)	(ScribListener, error)
	Dial(string, int) (BinChannel, error)
}*/

type ScribListener interface {
	Accept() (BinChannel, error)
}

type BinChannel interface {
	//io.Closer

	// Conn implements the Reader and Writer interfaces -- can use with gob
	GetConn() net.Conn
	Write(bs []byte) error
	Read(bs []byte) error  // Read fully
	Close() error
}
