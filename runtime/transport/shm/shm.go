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

type t = interface{}

// ConnCfg is a connection configuration, contains
// the details required to establish a connection.
type ConnCfg struct {
	chl chan t
	chr chan t
}

type Conn struct {
	chl chan t
	chr chan t
}

// NewConnection is a convenient wrapper for a TCP connection
// and can be used as either server-side or client-side.
func NewConnection() ConnCfg {
	return ConnCfg{make(chan t), make(chan t)}
}

// NewConnection is a convenient wrapper for a TCP connection
// and can be used as either server-side or client-side.
func NewBufferedConnection(n int) ConnCfg {
	return ConnCfg{make(chan t, n), make(chan t, n)}
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

	switch val := val.(type) {
	case bool:
		c.chl <- &val
	case float32:
		c.chl <- &val
	case float64:
		c.chl <- &val
	case int:
		c.chl <- &val
	case int8:
		c.chl <- &val
	case int16:
		c.chl <- &val
	case int32:
		c.chl <- &val
	case int64:
		c.chl <- &val
	case uint:
		c.chl <- &val
	case uint8:
		c.chl <- &val
	case uint16:
		c.chl <- &val
	case uint32:
		c.chl <- &val
	case uint64:
		c.chl <- &val
	case uintptr:
		c.chl <- &val
	case string:
		c.chl <- &val
	default:
		// Handle pointer types.
		c.chl <- &val
	}
	return nil
}

func (c *Conn) Recv(ptr interface{}) error {
	if c.chr == nil {
		return DeserialisationError{}
	}
	ifacePtr := <-c.chr

	switch ptr := ptr.(type) {
	case *bool:
		*ptr = *(ifacePtr.(*bool))
	case *float32:
		*ptr = *(ifacePtr.(*float32))
	case *float64:
		*ptr = *(ifacePtr.(*float64))
	case *int:
		*ptr = *(ifacePtr.(*int))
	case *int8:
		*ptr = *(ifacePtr.(*int8))
	case *int16:
		*ptr = *(ifacePtr.(*int16))
	case *int32:
		*ptr = *(ifacePtr.(*int32))
	case *int64:
		*ptr = *(ifacePtr.(*int64))
	case *uint:
		*ptr = *(ifacePtr.(*uint))
	case *uint8:
		*ptr = *(ifacePtr.(*uint8))
	case *uint16:
		*ptr = *(ifacePtr.(*uint16))
	case *uint32:
		*ptr = *(ifacePtr.(*uint32))
	case *uint64:
		*ptr = *(ifacePtr.(*uint64))
	case *uintptr:
		*ptr = *(ifacePtr.(*uintptr))
	case *string:
		*ptr = *(ifacePtr.(*string))
	default:
		// Handle pointer types.
		dstPtr := *(**unsafe.Pointer)(ifaceToConcrete(ptr))
		srcPtr := (*unsafe.Pointer)(ifaceToConcrete(**(**interface{})(ifaceToConcrete(ifacePtr))))
		*dstPtr = *srcPtr
	}
	return nil
}

// ifaceToConcrete converts an interface to its concrete (pointer) value
// based on the internals of the Go interface implementation.
// This conversion skips the first 1-word of iftable and
// returns an unsafe.Pointer to the concrete type.
func ifaceToConcrete(ptr interface{}) unsafe.Pointer {
	var word uint
	return unsafe.Pointer(uintptr(unsafe.Pointer((*interface{})(unsafe.Pointer(&ptr)))) + unsafe.Sizeof(word))
}
