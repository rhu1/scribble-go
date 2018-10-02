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
	// ReadWriteCloser is the standard binary channel interface in Go.
	// It contains Read/Write/Close method.
	io.ReadWriteCloser
}
