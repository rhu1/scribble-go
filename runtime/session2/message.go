package session2

import (
	//"bytes"
	//"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
	"github.com/rhu1/scribble-go-runtime/runtime/transport2/shm"
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

type GobFormatter struct {
	c transport2.BinChannel
	enc *gob.Encoder
	dec *gob.Decoder
}

func (f *GobFormatter) Wrap(c transport2.BinChannel) {
	f.c = c
	f.enc = gob.NewEncoder(c.GetWriter())
	f.dec = gob.NewDecoder(c.GetReader())
}	

func (f *GobFormatter) Serialize(m ScribMessage) error {
	return f.enc.Encode(&m)  // Encode *ScribMessage
}

func (f *GobFormatter) Deserialize(m *ScribMessage) (error) {
  err := f.dec.Decode(m)  // Decode *ScribMessage
	return err
}

// FIXME: (rename?) and move to shm package
type PassByPointer struct {
	c *shm.ShmChannel
}

func (f *PassByPointer) Wrap(c transport2.BinChannel) {
	f.c = c.(*shm.ShmChannel)
}	

func (f *PassByPointer) Serialize(m ScribMessage) error {
	f.c.WritePointer(&m)
	return nil
}

func (f *PassByPointer) Deserialize(m *ScribMessage) error {
	var ptr interface{}	
	f.c.ReadPointer(&ptr)
	*m = *(ptr.(*ScribMessage))  // CHECKME: is this copying the actual message value, or just the "interface value"?
	return nil
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
