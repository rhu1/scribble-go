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
Package transport provides a common interface for binary transport for the
Scribble runtime.

The basic abstraction is binary (client-server) transport which establishes
connection by Accept and Connect:

	var t Transport
	c := t.Accept() // Server accepting connection from client as c.
	...
	s := t.Connect() // Client connecting to server as s.

*/
package transport

import "io"

// Transport is an interface that creates a binary communication channel.
type Transport interface {
	// Accept establishes a connection by listening and accepting a connection
	// from the opposite side of the Transport endpoint.
	//
	// The caller of Accept is typically the server-side of a binary transport.
	Accept() Channel

	// Connect establishes a connection by connecting to the opposite side of
	// the Transport endpoint.
	//
	// The caller of Connect is typically the client-side of a binary transport.
	Connect() Channel

	Request() Channel
}

// Channel is an abstract binary communication channel
// where messages are Send and Recv on the channel.
type Channel interface {
	io.Closer

	ScribWrite(bs []byte) error
	ScribRead(bs *[]byte) error

	// Send sends a value val to the channel.
	// The value is transformed and transmitted over the Transport.
	Send(val interface{}) error

	// Recv receives a value from the underlying Transport,
	// and writes the received value to ptr (pointer to the variable).
	Recv(ptr interface{}) error
}
