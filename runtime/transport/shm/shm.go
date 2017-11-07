// Copyright 2017 The Scribble Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package shm provides a 'channel' transport implementation.

Typical usage of a connection:

	cfg := NewConnection() // starts channel
	s := cfg.Accept()      // does nothing except checking channel is OK
	defer func(s *Conn){
		if err := s.Close(); err != nil {
			// handle errors
		}(s)
	}
	...
	c := cfg.Connect() // does nothing except checking channel is OK
	defer func(c *Conn){
		if err := c.Close(); err != nil {
			// handle errors
		}(c)
	}

Client c and Server s can then be used as I/O streams implementing
io.Reader and io.Writer.

*/
package shm

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/nickng/scribble-go/runtime/transport"
)

// SerialisationError is the kind of error where a value
// cannot be sent due to serialisation failure.
type SerialisationError struct {
	cause error
}

func (e SerialisationError) Error() string {
	return fmt.Sprintf("transport/shm send: serialisation failed: %v", e.cause)
}

// DeserialisationError is the kind of error where a value
// cannot be received due to deserialisation failure.
type DeserialisationError struct {
	cause error
}

func (e DeserialisationError) Error() string {
	return fmt.Sprintf("transport/shm recv: deserisalisation failed: %v", e.cause)
}

// ConnCfg is a connection configuration, contains
// the details required to establish a connection.
type ConnCfg struct {
	chl chan unsafe.Pointer
	chr chan unsafe.Pointer
}

type Conn struct {
	chl chan unsafe.Pointer
	chr chan unsafe.Pointer
}

// NewConnection is a convenient wrapper for a TCP connection
// and can be used as either server-side or client-side.
func NewConnection() ConnCfg {
	return ConnCfg{make(chan unsafe.Pointer), make(chan unsafe.Pointer)}
}

// NewConnection is a convenient wrapper for a TCP connection
// and can be used as either server-side or client-side.
func NewBufferedConnection(n int) ConnCfg {
	return ConnCfg{make(chan unsafe.Pointer, n), make(chan unsafe.Pointer, n)}
}

// Connect establishes a connection with a TCP socket using details
// from cfg, and returns the TCP stream as a ReadWriteCloser.
func (cfg ConnCfg) Connect() transport.Channel {
	if cfg.chl == nil || cfg.chr == nil {
		log.Fatalf("transport/shm: invalid channel")
	}
	return &Conn{chl: cfg.chr, chr: cfg.chl}
}

// Accept listens for and accepts connection from a TCP socket using
// details from cfg, and returns the TCP stream as a ReadWriteCloser.
//
// Accept blocks while waiting for connection to be accepted.
func (cfg ConnCfg) Accept() transport.Channel {
	if cfg.chl == nil || cfg.chr == nil {
		log.Fatalf("transport/shm: invalid channel")
	}
	return &Conn{cfg.chl, cfg.chr}
}

func (c *Conn) Close() error {
	return nil
}

func (c *Conn) Send(val interface{}) error {
	if c.chl == nil {
		return SerialisationError{}
	}
	c.chl <- unsafe.Pointer(&val)
	return nil
}

func (c *Conn) Recv(ptr interface{}) error {
	if c.chr == nil {
		return DeserialisationError{}
	}
	uptr := <-c.chr

	var word uint

	switch ptr.(type) {
	case *bool:
		**(**bool)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**bool)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *float32:
		**(**float32)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**float32)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *float64:
		**(**float64)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**float64)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *int:
		**(**int)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**int)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *int8:
		**(**int8)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**int8)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *int16:
		**(**int16)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**int16)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *int32:
		**(**int32)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**int32)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *int64:
		**(**int64)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**int64)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *uint:
		**(**uint)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**uint)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *uint8:
		**(**uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**uint8)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *uint16:
		**(**uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**uint16)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *uint32:
		**(**uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**uint32)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *uint64:
		**(**uint64)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**uint64)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *uintptr:
		**(**uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**uintptr)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	case *string:
		**(**string)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**string)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	default:
		**(**unsafe.Pointer)(unsafe.Pointer(uintptr(unsafe.Pointer(&ptr)) + unsafe.Sizeof(word))) =
			**(**unsafe.Pointer)(unsafe.Pointer(uintptr(uptr) + unsafe.Sizeof(word)))
	}
	return nil
}
