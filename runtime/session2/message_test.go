package session2

import (
	"bytes"
	"encoding/gob"
	"io"
	"testing"
)

type fakeChan struct {
	buf *bytes.Buffer
	ptr chan interface{}
}

func newFakeChan(N int) *fakeChan {
	return &fakeChan{new(bytes.Buffer), make(chan interface{}, N)}
}
func (c *fakeChan) GetReader() io.Reader       { return c.buf }
func (c *fakeChan) GetWriter() io.Writer       { return c.buf }
func (c *fakeChan) Close() error               { return nil }
func (c *fakeChan) ReadPointer(m *interface{}) { *m = <-c.ptr }
func (c *fakeChan) WritePointer(m interface{}) { c.ptr <- m }

func mockMPChan(fmtr ScribMessageFormatter) *MPChan {
	mpc := NewMPChan(0, []string{""})
	mpc.Fmts[""][0] = fmtr
	return mpc
}

type transport struct {
	name       string
	sFmt, rFmt ScribMessageFormatter
}

func newFakeTransports(N int) []transport {
	transports := []transport{
		transport{"shm", new(PassByPointer), new(PassByPointer)},
		transport{"tcp", new(GobFormatter), new(GobFormatter)},
	}
	for _, transport := range transports {
		ch := newFakeChan(N)
		transport.sFmt.Wrap(ch)
		transport.rFmt.Wrap(ch)
	}
	return transports
}

// This covers the cases when the message is of format: Label(int)
func TestSerialisePrimitiveType(t *testing.T) {
	transports := newFakeTransports(3)

	toSend := []int{1, 2, 3}
	toRecv := make([]int, 3)

	for _, transport := range transports {
		t.Run(transport.name, func(t *testing.T) {
			mpc := mockMPChan(transport.sFmt)
			// Send
			for i := range toSend {
				if err := mpc.ISend("", 0, &toSend[i]); err != nil {
					t.Errorf("serialise failed: %v", err)
				}
			}
			// Receive
			for i := range toRecv {
				if err := mpc.IRecv("", 0, &toRecv[i]); err != nil {
					t.Errorf("deserialise failed: %v", err)
				}
			}
			if want, got := len(toSend), len(toRecv); want != got {
				t.Errorf("mismatch: sent %d items but received %d", want, got)
			}
			for i := range toSend {
				if want, got := toSend[i], toRecv[i]; want != got {
					t.Errorf("mismatch at %d: sent %#v but got %#v", i, want, got)
				}
			}
		})
	}
}

// This covers the cases when the message is of format: Label(StructType)
// where type StructType struct { Field int }
func TestSerialiseStructType(t *testing.T) {
	transports := newFakeTransports(3)

	type StructType struct {
		Field int
	}

	toSend := []StructType{StructType{1}, StructType{2}, StructType{3}}
	toRecv := make([]StructType, 3)

	for _, transport := range transports {
		t.Run(transport.name, func(t *testing.T) {
			mpc := mockMPChan(transport.sFmt)
			gob.Register(new(StructType)) // Register type
			// Send
			for i := range toSend {
				if err := mpc.ISend("", 0, &toSend[i]); err != nil {
					t.Errorf("serialise failed: %v", err)
				}
			}
			// Receive
			for i := range toRecv {
				if err := mpc.IRecv("", 0, &toRecv[i]); err != nil {
					t.Errorf("deserialise failed: %v", err)
				}
			}
			if want, got := len(toSend), len(toRecv); want != got {
				t.Errorf("mismatch: sent %d items but received %d", want, got)
			}
			for i := range toSend {
				if want, got := toSend[i], toRecv[i]; want != got {
					t.Errorf("mismatch at %d: sent %#v but got %#v", i, want, got)
				}
			}
		})
	}
}

// This covers the cases when the message is of format: Sig declared as sig
// where type Sig int
func TestSerialiseNamedSig(t *testing.T) {
	transports := newFakeTransports(3)

	type NamedSig int

	toSend := []NamedSig{NamedSig(1), NamedSig(2), NamedSig(3)}
	toRecv := make([]NamedSig, 3)

	for _, transport := range transports {
		t.Run(transport.name, func(t *testing.T) {
			mpc := mockMPChan(transport.sFmt)
			gob.Register(new(NamedSig)) // Register type
			// Send
			for i := range toSend {
				if err := mpc.ISend("", 0, &toSend[i]); err != nil {
					t.Errorf("serialise failed: %v", err)
				}
			}
			// Receive
			for i := range toRecv {
				if err := mpc.IRecv("", 0, &toRecv[i]); err != nil {
					t.Errorf("deserialise failed: %v", err)
				}
			}
			if want, got := len(toSend), len(toRecv); want != got {
				t.Errorf("mismatch: sent %d items but received %d", want, got)
			}
			for i := range toSend {
				if want, got := toSend[i], toRecv[i]; want != got {
					t.Errorf("mismatch at %d: sent %#v but got %#v", i, want, got)
				}
			}
		})
	}
}

// This covers the cases when the message is of format: StructSig declared as sig
// where type StructSig struct { Field int }
func TestSerialiseStructSig(t *testing.T) {
	transports := newFakeTransports(3)

	type StructSig struct {
		Field int
	}

	toSend := []StructSig{StructSig{1}, StructSig{2}, StructSig{3}}
	toRecv := make([]StructSig, 3)

	for _, transport := range transports {
		t.Run(transport.name, func(t *testing.T) {
			mpc := mockMPChan(transport.sFmt)
			gob.Register(new(StructSig)) // Register type
			// Send
			for i := range toSend {
				if err := mpc.ISend("", 0, &toSend[i]); err != nil {
					t.Errorf("serialise failed: %v", err)
				}
			}
			// Receive
			for i := range toRecv {
				if err := mpc.IRecv("", 0, &toRecv[i]); err != nil {
					t.Errorf("deserialise failed: %v", err)
				}
			}
			if want, got := len(toSend), len(toRecv); want != got {
				t.Errorf("mismatch: sent %d items but received %d", want, got)
			}
			for i := range toSend {
				if want, got := toSend[i].Field, toRecv[i].Field; want != got {
					t.Errorf("mismatch at %d: sent %#v but got %#v", i, want, got)
				}
			}
		})
	}
}

// This covers the cases when the message is of format: StructPtrFieldSig declared as sig
// where type StructPtrFieldSig struct { Field *int }
func TestSerialiseStructPtrFieldSig(t *testing.T) {
	transports := newFakeTransports(3)

	type StructPtrFieldSig struct {
		Field *int
	}

	i0, i1, i2 := 1, 2, 3
	toSend := []StructPtrFieldSig{
		StructPtrFieldSig{&i0},
		StructPtrFieldSig{&i1},
		StructPtrFieldSig{&i2},
	}
	toRecv := []StructPtrFieldSig{
		StructPtrFieldSig{new(int)},
		StructPtrFieldSig{new(int)},
		StructPtrFieldSig{new(int)},
	}

	for _, transport := range transports {
		t.Run(transport.name, func(t *testing.T) {
			mpc := mockMPChan(transport.sFmt)
			gob.Register(new(StructPtrFieldSig)) // Register type
			// Send
			for i := range toSend {
				if err := mpc.ISend("", 0, &toSend[i]); err != nil {
					t.Errorf("serialise failed: %v", err)
				}
			}
			// Receive
			for i := range toRecv {
				if err := mpc.IRecv("", 0, &toRecv[i]); err != nil {
					t.Errorf("deserialise failed: %v", err)
				}
			}
			if want, got := len(toSend), len(toRecv); want != got {
				t.Errorf("mismatch: sent %d items but received %d", want, got)
			}
			for i := range toSend {
				if want, got := *toSend[i].Field, *toRecv[i].Field; want != got {
					t.Errorf("mismatch at %d: sent %#v but got %#v", i, want, got)
				}
			}
		})
	}
}

// This covers the cases when the message is of format: PtrPrimitiveSig declared as sig
// where type PtrPrimitiveSig *int
func TestSerialisePtrPrimitiveSig(t *testing.T) {
	transports := newFakeTransports(3)[0:0] // Note: only test shm version.

	i0, i1, i2 := 1, 2, 3
	toSend := []*int{&i0, &i1, &i2}
	toRecv := make([]*int, 3)

	for _, transport := range transports {
		t.Run(transport.name, func(t *testing.T) {
			mpc := mockMPChan(transport.sFmt)
			// Send
			for i := range toSend {
				if err := mpc.ISend("", 0, &toSend[i]); err != nil {
					t.Errorf("serialise failed: %v", err)
				}
			}
			// Receive
			for i := range toRecv {
				if err := mpc.IRecv("", 0, &toRecv[i]); err != nil {
					t.Errorf("deserialise failed: %v", err)
				}
			}
			if want, got := len(toSend), len(toRecv); want != got {
				t.Errorf("mismatch: sent %d items but received %d", want, got)
			}
			for i := range toSend {
				if want, got := *toSend[i], *toRecv[i]; want != got {
					t.Errorf("mismatch at %d: sent %#v but got %#v", i, want, got)
				}
			}
		})
	}
}
