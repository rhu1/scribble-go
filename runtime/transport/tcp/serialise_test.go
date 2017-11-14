package tcp

import (
	"bytes"
	"testing"
)

func TestPassthruSerialise(t *testing.T) {
	var buf bytes.Buffer
	enc := passthruSerialiser{w: &buf}
	dec := passthruDeserialiser{r: &buf}
	b := []byte("hello world")
	t.Log("Origin:", b)
	enc.Encode(b)
	t.Log("Encoded:", buf.Bytes())
	if want, got := b, buf.Bytes(); bytes.Compare(want, got) != 0 {
		t.Errorf("PassthruSerialise: expected encoded message to be %v but got %v", want, got)
	}
	s := make([]byte, 100)
	dec.Decode(&s)
	t.Log("Decoded:", s)
	if want, got := b, s; bytes.Compare(want, got) != 0 {
		t.Errorf("PassthruSerialise: expected encoded message to be %v but got %v", want, got)
	}
}
