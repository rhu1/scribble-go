package session2

import (
	"encoding/gob"
	"unsafe"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

// ScribMessage represents a basic message.
//
// Corresponds to Scribble Java runtime's
// org.scribble.runtime.net.ScribMessage class
type ScribMessage interface {
	GetOp() string
}

// ScribMessageFormatter represents a message formatter (or serialiser).
//
// Corresponds to Scribble Java runtime's
// org.scribble.runtime.net.ScribMessageFormatter class
type ScribMessageFormatter interface {
	// Wrap wraps a given binary channel with the formatter.
	// The channel is the target input and output stream for
	// serialisation and deserialisation using the formatter.
	Wrap(c transport2.BinChannel)

	// Serialize serialises the given message m the writes to
	// the underlying output stream.
	Serialize(m ScribMessage) error

	// Deserialize read from the underlying input stream and
	// deserialises the message into the message container m.
	Deserialize(m *ScribMessage) error
}

// GobFormatter is an implementation of a ScribMessageFormatter using
// Go's "encoding/gob" package.
//
// User must manually run gob.Register on the message type pointer
// to register the message type for encoding.
//
//     // T is type being sent/received
//     gob.Register(new(T))
//
// The snippet above registers type T for sending and receiving.
//
type GobFormatter struct {
	c   transport2.BinChannel
	enc *gob.Encoder
	dec *gob.Decoder
}

// Wrap wraps a binary channel for gob encoding.
func (f *GobFormatter) Wrap(c transport2.BinChannel) {
	f.c = c
	f.enc = gob.NewEncoder(c)
	f.dec = gob.NewDecoder(c)
}

// Serialize encodes a ScribMessage m using gob.
//
// The message type implementing ScribMessage should be
// registered before calling this method.
func (f *GobFormatter) Serialize(m ScribMessage) error {
	// If the message is recognised as a special message type,
	// use special encoding strategy.
	switch smsg := m.(type) {
	case wrapper:
		return f.enc.Encode(smsg.Msg)
	}
	// Encode ScribMessage as-is.
	return f.enc.Encode(&m)
}

// Deserialize decode a ScribMessage m using gob.
//
// The message type implementing ScribMessage should be
// registered before calling this method.
func (f *GobFormatter) Deserialize(m *ScribMessage) error {
	// If the message container is recognised as a special message type,
	// use special decoding strategy.
	switch smeg := (*m).(type) {
	case wrapper:
		msg := smeg.Msg
		return f.dec.Decode(msg)
	}
	// Decode ScribMessage as-is.
	return f.dec.Decode(m)
}

// PointerWriter is an interface for writing a pointer ptr to a channel.
//
// Transports supporting transparent movement
// of memory should implement this interface.
type PointerWriter interface {
	WritePointer(ptr interface{})
}

// PointerReader is an interface for reading a pointer from a channel
// and write the received content to ptr.
//
// Transports supporting transparent movement
// of memory should implement this interface.
type PointerReader interface {
	ReadPointer(ptr *interface{})
}

// PointerReadWriter is an interface for reading and writing
// a pointer over a channel.
//
// Transports supporting transparent movement
// of memory should implement this interface.
type PointerReadWriter interface {
	PointerWriter
	PointerReader
}

// PassByPointer is an implementation of ScribMessageFormatter
// using pointer passing.
//
// This is formatter is only available for transports
// implementing PointerReadWriter.
type PassByPointer struct {
	c PointerReadWriter
}

// Wrap wraps a binary channel for pointer encoding.
func (f *PassByPointer) Wrap(c transport2.BinChannel) {
	f.c = c.(PointerReadWriter)
}

// Serialize encodes a ScribMessage m as pointer.
func (f *PassByPointer) Serialize(m ScribMessage) error {
	switch m := m.(type) {
	case wrapper:
		msg := m.Msg
		f.c.WritePointer(msg)
	default:
		f.c.WritePointer(&m)
	}
	return nil
}

// Deserialize decodes a ScribMessage m as pointer.
func (f *PassByPointer) Deserialize(m *ScribMessage) error {
	switch smsg := (*m).(type) {
	case wrapper:
		// rcvd is a temporary container
		// to hold received data from ReadPointer.
		var rcvd interface{}
		f.c.ReadPointer(&rcvd) // at runtime rcvd is of type *T

		// In current implementation, m is always wrapped
		// and the real *T is in (wrapper).Msg
		// Note: msg is an interface{}, but at runtime it is the
		// address to where the received data should go (i.e. *T)
		msg := smsg.Msg
		ptrToMsg := derefIface(msg)
		ptrToRcvd := derefIface(rcvd)

		if src, ok := rcvd.(*string); ok {
			// The pointer assignment below does not work for string variables
			// as string is represented as a 2-word data structure, the length
			// (2nd word) also needs to be updated for a successful assignment
			//
			//    *s = [ ptr | len ]
			//
			// this workround treats *string as *string (not just a pointer)
			// and handles the 2-word write correctly.
			*smsg.Msg.(*string) = *src
		} else {
			// This assignment is equivalent to
			// *msg = *rcvd
			// except msg and *rcvd are both hidden under interface{}
			*(**unsafe.Pointer)(unsafe.Pointer(uintptr(ptrToMsg))) =
				*(**unsafe.Pointer)(unsafe.Pointer(uintptr(ptrToRcvd)))
		}
	default:
		var ptr interface{}
		f.c.ReadPointer(&ptr)
		*m = *(ptr.(*ScribMessage))
	}
	return nil
}

// derefIface takes an interface and returns a pointer
// to its underlying value.
func derefIface(iface interface{}) unsafe.Pointer {
	var word uint
	// An interface{} is a 2-word wide data structure where the latter word
	// contains pointer to the underlying value in the interface{} variable
	return *(*unsafe.Pointer)(unsafe.Pointer(uintptr(unsafe.Pointer(&iface)) + unsafe.Sizeof(word)))
}
