package session2

import (
	//"bytes"
	//"encoding/base64"
	"encoding/gob"
	"fmt"
	"io"
	"unsafe"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

var _ = fmt.Print


// cf. org.scribble.runtime.net.ScribMessage
type ScribMessage interface {
	GetOp() string
}

// cf. org.scribble.runtime.net.ScribMessageFormatter
type ScribMessageFormatter interface {
	Wrap(transport2.BinChannel) 	
	Serialize(ScribMessage) error
	Deserialize(*ScribMessage) error

	/*EncodeInt(int) error
	DecodeInt() (int, error)
	EncodeString(string) error
	DecodeString() (string, error)
	EncodeBytes([]byte) error
	DecodeBytes() ([]byte, error)*/
	
	/*GetEnc() *gob.Encoder
	GetDec() *gob.Decoder*/
}


/**
 * N.B. must do gob.Register on _pointer_ to message types (cf. sigs02, shm03) -- because MPChan MSend/Receive communicate by pointer (for efficient transparency with shm)
 */
type GobFormatter struct {
	c transport2.BinChannel
	enc *gob.Encoder
	dec *gob.Decoder
	rdr io.Reader
}

func (f *GobFormatter) Wrap(c transport2.BinChannel) {
	f.c = c
	f.enc = gob.NewEncoder(c.GetWriter())
	f.rdr = c.GetReader()
	f.dec = gob.NewDecoder(f.rdr)
}

func (f *GobFormatter) Serialize(m ScribMessage) error {
	//fmt.Printf("Serialize: %v %T\n", m, m)
	switch smsg := m.(type) {
	case wrapper:
		return f.enc.Encode(smsg.Msg)
	default:
		return f.enc.Encode(&m) // Encode *ScribMessage  // CHECKME just m? not &m
	}
	// "val" should be m
}

func (f *GobFormatter) Deserialize(m *ScribMessage) error {
	//fmt.Printf("Deserialize1: %v %T\n", *m, *m)
	//b := make([]byte, 100)
	//f.rdr.Read(b)
	//fmt.Printf("To deserialise\n", b)

	switch smeg := (*m).(type) {
	case wrapper:
		msg := smeg.Msg
		return f.dec.Decode(msg)
	default:
		return f.dec.Decode(m) // Decode *ScribMessage
	}
	// pointer, "m" is *

	//fmt.Printf("Deserialize2: %v %T\n", *m, *m)
}

// PointerWriter is an interface for writing a pointer ptr to a channel.
type PointerWriter interface {
	WritePointer(ptr interface{})
}

// PointerReader is an interface for reading a pointer from a channel
// and write the received content to ptr.
type PointerReader interface {
	ReadPointer(ptr *interface{})
}

// PointerReadWriter is an interface for reading and writing
// a pointer over a channel.
type PointerReadWriter interface {
	PointerWriter
	PointerReader
}

// FIXME: (rename?) and move to shm package
type PassByPointer struct {
	c PointerReadWriter
}

func (f *PassByPointer) Wrap(c transport2.BinChannel) {
	f.c = c.(PointerReadWriter)
}

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

		// This assignment is equivalent to
		// *msg = *rcvd
		// except msg and *rcvd are both hidden under interface{}
		*(**unsafe.Pointer)(unsafe.Pointer(uintptr(ptrToMsg))) =
			*(**unsafe.Pointer)(unsafe.Pointer(uintptr(ptrToRcvd)))
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

/*func (f *GobFormatter) EncodeInt(m int) error {
	return f.enc.Encode(&m)
}

func (f *GobFormatter) DecodeInt() (int, error) {
	var m int
	err := f.dec.Decode(&m)
	return m, err
}

func (f *GobFormatter) EncodeString(m string) error {
	return f.enc.Encode(&m)
}

func (f *GobFormatter) DecodeString() (string, error) {
	var m string
	err := f.dec.Decode(&m)
	return m, err
}

func (f *GobFormatter) EncodeBytes(m []byte) error {
	return f.enc.Encode(&m)
}

func (f *GobFormatter) DecodeBytes() ([]byte, error) {
	var m []byte
	err := f.dec.Decode(&m)
	return m, err
}*/

/*func (f *GobFormatter) GetEnc() *gob.Encoder {
	return f.enc
}

func (f *GobFormatter) GetDec() *gob.Decoder {
	return f.dec
}*/
