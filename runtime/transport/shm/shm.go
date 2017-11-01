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
	"reflect"

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
	ch chan interface{}
}

type Conn struct {
	ch chan interface{}
}

// NewConnection is a convenient wrapper for a TCP connection
// and can be used as either server-side or client-side.
func NewConnection() ConnCfg {
	return ConnCfg{make(chan interface{})}
}

// NewConnection is a convenient wrapper for a TCP connection
// and can be used as either server-side or client-side.
func NewBufferedConnection(n int) ConnCfg {
	return ConnCfg{make(chan interface{}, n)}
}

// Connect establishes a connection with a TCP socket using details
// from cfg, and returns the TCP stream as a ReadWriteCloser.
func (cfg ConnCfg) Connect() transport.Channel {
	if cfg.ch == nil {
		log.Fatalf("transport/shm: invalid channel")
	}
	return &Conn{cfg.ch}
}

// Accept listens for and accepts connection from a TCP socket using
// details from cfg, and returns the TCP stream as a ReadWriteCloser.
//
// Accept blocks while waiting for connection to be accepted.
func (cfg ConnCfg) Accept() transport.Channel {
	if cfg.ch == nil {
		log.Fatalf("transport/shm: invalid channel")
	}
	return &Conn{cfg.ch}
}

func (c *Conn) Close() error {
	return nil
}

func (c *Conn) Send(val interface{}) error {
	if c.ch == nil {
		return SerialisationError{}
	}
	c.ch <- val
	return nil
}

func (c *Conn) Recv(ptr interface{}) error {
	if c.ch == nil {
		return DeserialisationError{}
	}
	v := <-c.ch

	ptrValue := reflect.ValueOf(ptr)
	val := reflect.Indirect(ptrValue)
	val.Set(reflect.ValueOf(v))
	return nil
}
