package tcp

import (
	"bufio"
	"bytes"
	"io"
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

func TestCrlfDelim(t *testing.T) {
	var buf bytes.Buffer
	c := &Conn{ // Mock conn to hold the buffers.
		bufr: bufio.NewReader(&buf),
		bufw: bufio.NewWriter(&buf),
	}
	r := crlfDelimReader{conn: c}
	w := crlfDelimWriter{conn: c}
	b := []byte("message\nwith\nnewline\n\r")
	t.Logf("Origin: %v", b)
	w.Write(b)
	t.Logf("Packed: %v", buf.Bytes())
	result := make([]byte, len(b)+3)
	n, err := r.Read(result)
	if err != nil && err != io.EOF {
		t.Logf("Read error: %v", err)
	}
	t.Logf("Unpacked: %v", result[:n])
	if want, got := b, result[:n]; bytes.Compare(want, got) != 0 {
		t.Errorf("crlfDelim: expected message to be %v but got %v", want, got)
	}
}

// This test if \r\n is in beginning of line
func TestCrlfDelim2(t *testing.T) {
	var buf bytes.Buffer
	c := &Conn{ // Mock conn to hold the buffers.
		bufr: bufio.NewReader(&buf),
		bufw: bufio.NewWriter(&buf),
	}
	r := crlfDelimReader{conn: c}
	w := crlfDelimWriter{conn: c}
	b := []byte("\r\nmessage\nwith\nnewline\n\r")
	t.Logf("Origin: %v", b)
	w.Write(b)
	t.Logf("Packed: %v", buf.Bytes())
	result := make([]byte, len(b)+3)
	n, err := r.Read(result)
	if err != nil && err != io.EOF {
		t.Logf("Read error: %v", err)
	}
	t.Logf("Unpacked: %v", result[:n])
	if want, got := b, result[:n]; bytes.Compare(want, got) != 0 {
		t.Errorf("crlfDelim: expected message to be %v but got %v", want, got)
	}
}
