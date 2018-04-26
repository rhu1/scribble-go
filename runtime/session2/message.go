package session2

import (
	//"bytes"
	//"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/rhu1/scribble-go-runtime/runtime/transport2"
)

var _ = fmt.Print

type T interface {
	
}

// cf. org.scribble.runtime.net.ScribMessage
type ScribMessage interface {
	GetOp() string
}

// cf. org.scribble.runtime.net.ScribMessageFormatter
type ScribMessageFormatter interface {
	Wrap(transport2.BinChannel) 	
	/*EncodeInt(int) error
	DecodeInt() (int, error)
	EncodeString(string) error
	DecodeString() (string, error)*/
	EncodeBytes([]byte) error
	DecodeBytes() ([]byte, error)
	Serialize(ScribMessage) error
	Deserialize() (ScribMessage, error)
	
	GetEnc() *gob.Encoder
	GetDec() *gob.Decoder
}

type GobFormatter struct {
	c transport2.BinChannel
	enc *gob.Encoder
	dec *gob.Decoder
}

func (f *GobFormatter) GetEnc() *gob.Encoder {
	return f.enc
}

func (f *GobFormatter) GetDec() *gob.Decoder {
	return f.dec
}

func (f *GobFormatter) Wrap(c transport2.BinChannel) {
	f.c = c
	f.enc = gob.NewEncoder(c.GetConn())
	f.dec = gob.NewDecoder(c.GetConn())
}	

/*type wrapper struct {
	Msg *ScribMessage	
	X int
}*/

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
}*/

func (f *GobFormatter) EncodeBytes(m []byte) error {
	return f.enc.Encode(&m)
}

func (f *GobFormatter) DecodeBytes() ([]byte, error) {
	var m []byte
	err := f.dec.Decode(&m)
	return m, err
}

func (f *GobFormatter) Serialize(m ScribMessage) error {
	return f.enc.Encode(&m)  // Encode *ScribMessage
}

func (f *GobFormatter) Deserialize() (ScribMessage, error) {
	var m ScribMessage
  err := f.dec.Decode(&m)  // Decode *ScribMessage
	return m, err
}
