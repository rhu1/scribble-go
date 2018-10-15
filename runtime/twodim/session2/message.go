package session2

import (
	"unsafe"

	"github.com/rhu1/scribble-go-runtime/runtime/session2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

type ScribMessage = session2.ScribMessage
type ScribMessageFormatter = session2.ScribMessageFormatter
type GobFormatter = session2.GobFormatter
type PointerReadWriter = session2.PointerReadWriter

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
