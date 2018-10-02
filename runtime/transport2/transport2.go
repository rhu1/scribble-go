package transport2

import (
	"io"
)

// ScribListener is a generic Listener
// implemented by all Scribble transports.
type ScribListener interface {
	Accept() (BinChannel, error)
	Close() error
}

// BinChannel is a generic binary channel
// implemented by all Scribble transports.
type BinChannel interface {
	//io.Closer

	// Conn implements the Reader and Writer interfaces -- can use with gob
	//GetConn() net.Conn
	GetReader() io.Reader	
	GetWriter() io.Writer	
	/*Write(bs []byte) error
	Read(bs []byte) error  // Read fully*/
	Close() error
}
