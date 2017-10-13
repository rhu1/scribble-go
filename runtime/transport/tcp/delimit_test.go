package tcp

import (
	"bufio"
	"bytes"
	"testing"
)

func TestSizeDelim(t *testing.T) {
	var buf bytes.Buffer
	c := &Conn{ // Mock conn to hold the buffers.
		bufr: bufio.NewReader(&buf),
		bufw: bufio.NewWriter(&buf),
	}
	r := sizeDelimReader{conn: c}
	w := sizeDelimWriter{conn: c}
	b := []byte("message")
	t.Logf("Origin: %v", b)
	w.Write(b)
	t.Logf("Packed: %v", buf.Bytes())
	result := make([]byte, 7)
	r.Read(result)
	t.Logf("Unpacked: %v", result)
	if want, got := b, result; bytes.Compare(want, got) != 0 {
		t.Errorf("SizeDelim: expected message to be %v but got %v", want, got)
	}
}

func TestNewlineDelim(t *testing.T) {
	var buf bytes.Buffer
	c := &Conn{ // Mock conn to hold the buffers.
		bufr: bufio.NewReader(&buf),
		bufw: bufio.NewWriter(&buf),
	}
	r := newlineDelimReader{conn: c}
	w := newlineDelimWriter{conn: c}
	b := []byte("message")
	t.Logf("Origin: %v", b)
	w.Write(b)
	t.Logf("Packed: %v", buf.Bytes())
	result := make([]byte, 7)
	r.Read(result)
	t.Logf("Unpacked: %v", result)
	if want, got := b, result; bytes.Compare(want, got) != 0 {
		t.Errorf("NewlineDelim: expected message to be %v but got %v", want, got)
	}
}
