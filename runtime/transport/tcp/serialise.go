package tcp

import (
	"encoding/gob"
	"fmt"
	"io"
)

// SerialiseMethod defines the serialisation method for the messages passed
// through the TCP transport.
type SerialiseMethod int

const (
	// SerialiseWithGob is the default serialisation method, where the
	// data stream are encoded with the Go-specific encoding/gob format.
	SerialiseWithGob SerialiseMethod = iota

	// SerialiseWithPassthru is the serialisation method, where the
	// data stream are passed through the transport unmodified.
	// This should be used for interacting with existing text-based
	// protocol where no encoding are needed and data are treated
	// directly as []byte or string.
	SerialiseWithPassthru
)

type serialiser interface {
	Encode(interface{}) error
}

type deserialiser interface {
	Decode(interface{}) error
}

// NewSerialiser returns a new serialisation encoder for the output writer w.
func NewSerialiser(w io.Writer, m SerialiseMethod) serialiser {
	switch m {
	case SerialiseWithGob:
		return gob.NewEncoder(w)
	case SerialiseWithPassthru:
		return &passthruSerialiser{w: w}
	}
	return nil
}

// NewDeserialiser returns a new deserialisation decoder for the input reader r.
func NewDeserialiser(r io.Reader, m SerialiseMethod) deserialiser {
	switch m {
	case SerialiseWithGob:
		return gob.NewDecoder(r)
	case SerialiseWithPassthru:
		return &passthruDeserialiser{r: r}
	}
	return nil
}

// passthruSerialiser is a serialiser the does not modify the outgoing.
type passthruSerialiser struct {
	w io.Writer
}

// Encode encodes and writes input v (of either []byte or string type) to the
// underlying output writer.
func (s *passthruSerialiser) Encode(v interface{}) error {
	switch v := v.(type) {
	case []byte:
		_, err := s.w.Write(v)
		if err != nil {
			return err
		}
		return nil
	case string:
		_, err := s.w.Write([]byte(v))
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("passthru serialiser Encode() expects a []byte or string parameter, got %v (type: %T)", v, v)
}

// passthruSerialiser is a serialiser that does not modify the incoming data.
type passthruDeserialiser struct {
	r io.Reader
}

// Decode reads from the underlying input reader and decodes
// to v as a byte slice, v must be of type *[]byte.
func (s *passthruDeserialiser) Decode(v interface{}) error {
	ptr, ok := v.(*[]byte)
	if !ok {
		return fmt.Errorf("passthru deserialiser Decode() expects a *[]byte parameter, got %v (type: %T)", v, v)
	}
	n, err := s.r.Read(*v.(*[]byte))
	if err != nil {
		return err
	}
	*ptr = (*v.(*[]byte))[:n]
	return nil
}
