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
Package shm provides a shared memory transport implementation.

The messages in this transport are delivered by native Go channels.

Typical usage of a connection:

	cfg := NewConnection()  // Initialises a shared memory connection
	s, c := cfg.Endpoints() // Equivalent to cfg.Accept(), cfg.Connect()
	defer func(s *IOChan){
		if err := s.Close(); err != nil {
			// handle errors
		}(s)
	}
	...
	defer func(c *IOChan){
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

	"github.com/rhu1/scribble-go-runtime/runtime/transport"
)

// ChannelNotReadyError is the kind of error where a channel
// is not ready to be used (e.g. uninitialised).
type ChannelNotReadyError struct {
	at string
}

func (e ChannelNotReadyError) Error() string {
	if e.at != "" {
		return fmt.Sprintf("transport/shm: channel not ready at %s", e.at)
	}
	return fmt.Sprintf("transport/shm: channel not ready\n\t(has it been initialised with NewConnection()?)")
}

type t interface{}

// ConnCfg is a connection configuration, contains
// the details required to establish a connection.
type ConnCfg struct {
	chl chan t // channel left
	chr chan t // channel right
}

// IOChan is a connected shared memory connection,
// and wraps a pair of read/write Go channels for communication.
//
// IOChan implements ReadWriteCloser and can be used as it. Messages
// are delimited by the data type boundary as they are passed as
// pointers. No serialisation are defined for IOChan.
type IOChan struct {
	chw chan<- t // channel to write to
	chr <-chan t // channel to read from
}

// NewConnection is a convenient wrapper for an in-memory connection
// and can be used as either server-side or client-side.
func NewConnection() ConnCfg {
	return ConnCfg{chl: make(chan t), chr: make(chan t)}
}

// NewBufferedConnection is a convenient wrapper for an in-memory connection
// and can be used as either server-side or client-side.
func NewBufferedConnection(n int) ConnCfg {
	return ConnCfg{chl: make(chan t, n), chr: make(chan t, n)}
}

// Connect establishes a connection with a TCP socket using details
// from cfg, and returns the TCP stream as a ReadWriteCloser.
func (cfg ConnCfg) Connect() transport.Channel {
	if cfg.chl == nil || cfg.chr == nil {
		log.Fatalf("cannot connect: %v", ChannelNotReadyError{})
	}
	return &IOChan{chw: cfg.chr, chr: cfg.chl}
}

// Accept listens for and accepts connection from a TCP socket using
// details from cfg, and returns the TCP stream as a ReadWriteCloser.
//
// Accept blocks while waiting for connection to be accepted.
func (cfg ConnCfg) Accept() transport.Channel {
	if cfg.chl == nil || cfg.chr == nil {
		log.Fatalf("cannot accept: %v", ChannelNotReadyError{})
	}
	return &IOChan{chw: cfg.chl, chr: cfg.chr}
}

// Endpoints is a convenient function that returns the dual endpoints
// created by Accept and Connect, such that messages sent to one endpoint
// can be received by the other and vice versa.
func (cfg ConnCfg) Endpoints() (s transport.Channel, c transport.Channel) {
	s, c = cfg.Accept(), cfg.Connect()
	return s, c
}

func (c *IOChan) Close() error {
	close(c.chw) // Close the sending side.
	return nil
}

func (c *IOChan) Send(val interface{}) error {
	if c.chw == nil {
		return ChannelNotReadyError{at: "Send()"}
	}

	switch val := val.(type) {
	case bool:
		c.chw <- &val
	case float32:
		c.chw <- &val
	case float64:
		c.chw <- &val
	case int:
		c.chw <- &val
	case int8:
		c.chw <- &val
	case int16:
		c.chw <- &val
	case int32:
		c.chw <- &val
	case int64:
		c.chw <- &val
	case uint:
		c.chw <- &val
	case uint8:
		c.chw <- &val
	case uint16:
		c.chw <- &val
	case uint32:
		c.chw <- &val
	case uint64:
		c.chw <- &val
	case uintptr:
		c.chw <- &val
	case string:
		c.chw <- &val
	default:
		// Handle pointer types.
		c.chw <- &val
	}
	return nil
}

func (c *IOChan) Recv(ptr interface{}) error {
	if c.chr == nil {
		return ChannelNotReadyError{at: "Recv()"}
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
