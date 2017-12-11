// Copyright 2017 The Scribble Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tcp

// Implementation of delimit methods.

import (
	"encoding/binary"
	"io"

	//"log"

)

// DelimitMethod defines the delimitation method for the messages passed
// through the TCP transport.
type DelimitMethod int

const (
	// DelimitBySize is the default delimitation method, where the
	// size of the message is prepended to each message. This is suitable
	// for larger messages where initial buffer allocations are often not
	// large enough to receive the complete message and requires resizing.
	// Using this method the allocation can be reduced to minimum since
	// the size of the message is known from the beginning.
	DelimitBySize DelimitMethod = iota

	// DelimitByNewline is an alternative delimitation method, where a
	// delimiter character '\n' is added to the end of each message.
	// This delimitation method is minimal but could cause error if
	// the message is not encoded (e.g. contains '\n' in message body).
	DelimitByNewline

	// DelmitByCRLF is a delimitation method where the delimiter is the
	// byte sequence '\r\n' (CRLF).
	// This delimitation method is used by common protocols such as HTTP.
	DelimitByCRLF
)

// NewDelimReader returns a new delimiter Reader for the connection c.
func NewDelimReader(c *Conn, m DelimitMethod) io.Reader {
	switch m {
	case DelimitByNewline:
		return &newlineDelimReader{conn: c}
	case DelimitBySize:
		return &sizeDelimReader{conn: c}
	case DelimitByCRLF:
		return &crlfDelimReader{conn: c}
	}
	return nil
}

// NewDelimWriter returns a new delimiter Writer for the connection c.
func NewDelimWriter(c *Conn, m DelimitMethod) io.Writer {
	switch m {
	case DelimitByNewline:
		return &newlineDelimWriter{conn: c}
	case DelimitBySize:
		return &sizeDelimWriter{conn: c}
	case DelimitByCRLF:
		return &crlfDelimWriter{conn: c}
	}
	return nil
}

// sizeDelimReader is a Reader that decode size-prefixed data stream.
type sizeDelimReader struct {
	conn *Conn
}

// Read reads from and decodes the underlying size-prefixed data stream
// and copies the first decoded data into p.
func (sdr *sizeDelimReader) Read(p []byte) (n int, err error) {
	sizeBytes, err := sdr.conn.bufr.Peek(8)
	if err != nil {
		return 0, err
	}
	size := int64(binary.LittleEndian.Uint64(sizeBytes))
	sdr.conn.bufr.Discard(8) // Skip size bytes.
	b := make([]byte, size)
	n, err = sdr.conn.bufr.Read(b)
	copy(p, b)

	//log.Println("<<", p)

	return n, err
}

// sizeDelimWriter is a Writer that encodes data stream into a
// size-prefixed data stream.
type sizeDelimWriter struct {
	conn *Conn
}

// Write encodes p into a size-prefixed data stream and writes
// the encoded data to the underlying stream.
func (sdw *sizeDelimWriter) Write(p []byte) (n int, err error) {

	//log.Println(">>", p)

	n, err = sdw.conn.bufw.Write(packSize(p))
	sdw.conn.bufw.Flush()
	return
}

// packSize prepends the size of data to return a prepended slice
// of bytes. The size is encoded in the first 8 bytes of prepended.
func packSize(data []byte) (prepended []byte) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(len(data)))
	return append(b, data...)
}

// unpackSize splits the prepended slice size from data, and returns
// the size and the size-truncated data.
func unpackSize(prepended []byte) (size int64, truncated []byte) {
	size = int64(binary.LittleEndian.Uint64(prepended[:8]))
	return size, prepended[8 : 8+size]
}

const (
	// delim is newline as per convention.
	delim = byte('\n')
	// delimLen is the length of the delimiter, always 1 (byte).
	delimLen = 1
)

// newlineDelimReader is a Reader that decodes
// a delimited data stream encoded by newlineDelimWriter.
type newlineDelimReader struct {
	conn *Conn
}

// Read reads from and decodes the underlying delimited data stream
// and copies the first decoded item into p.
func (ndr *newlineDelimReader) Read(p []byte) (n int, err error) {
	b, err := ndr.conn.bufr.ReadBytes(delim)
	if err != nil {
		copy(p, b)
		return len(b), err
	}
	copy(p, b[:len(b)-delimLen])
	return len(b) - delimLen, err
}

// newlineDelimWriter is a Writer that encodes
// data stream into a delimited data stream.
type newlineDelimWriter struct {
	conn *Conn
}

// Write encodes p into a delimited data stream and writes
// the encoded data to the underlying stream.
func (ndw *newlineDelimWriter) Write(p []byte) (n int, err error) {
	n, err = ndw.conn.bufw.Write(p)
	err = ndw.conn.bufw.WriteByte(delim)
	if err != nil {
		ndw.conn.rwc.Close()
	}
	ndw.conn.bufw.Flush()
	return n, err
}

var (
	crlf       = []byte{'\r', '\n'}
	doubleCrlf = []byte{'\r', '\n', '\r', '\n'}
)

// crlfDelimReader is a Reader that decodes
// a \r\n delimited data stream encoded by crlfDelimWriter.
type crlfDelimReader struct {
	conn  *Conn
	state int
}

// Read reads from and decodes the underlying delimited data stream
// and copies the first decoded item into p.
func (cdr *crlfDelimReader) Read(p []byte) (n int, err error) {
	const (
		stateBegin = iota
		stateData
		stateCR
		stateCRLFBegin
		stateCRLFCR // End of line: \r\n
		stateEnd    // End of segment: \r\n \r\n
	)
	for n < len(p) && cdr.state != stateEnd {
		var b byte
		b, err := cdr.conn.bufr.ReadByte()
		if err != nil {
			if err == io.EOF {
				cdr.state = stateBegin // reset state
				return n, io.EOF       // end of all data
			}
			break
		}
		switch cdr.state {
		case stateBegin:
			if b == '\r' {
				cdr.state = stateCR
				continue
			}
			cdr.state = stateData

		case stateData:
			if b == '\r' {
				cdr.state = stateCR
				continue
			}

		case stateCR:
			if b == '\n' {
				cdr.state = stateCRLFBegin
				continue
			}
			// Reset state: last read is normal data.
			cdr.conn.bufr.UnreadByte()
			b = '\r'
			cdr.state = stateData

		case stateCRLFBegin:
			if b == '\r' {
				cdr.state = stateCRLFCR
				continue
			}
			if n+2 < len(p) {
				p[n] = '\r'
				n++
				p[n] = '\n'
				n++
			}
			cdr.state = stateBegin

		case stateCRLFCR:
			if b == '\n' {
				cdr.state = stateEnd
				continue
			}
			// Reset state: last read is normal data.
			cdr.conn.bufr.UnreadByte()
			b = '\r'
			cdr.state = stateData
		}
		p[n] = b
		n++
	}
	if err == nil && cdr.state == stateEnd { // Normal EOF
		// end of delimited data (more to come)
		cdr.state = stateBegin
	}
	return
}

// newlineDelimWriter is a Writer that adds \r\n suffix to delimit data.
type crlfDelimWriter struct {
	conn *Conn
}

// Write encode p into delimited data stream and writes
// the encoded data to the underlying stream.
func (cdw *crlfDelimWriter) Write(p []byte) (n int, err error) {
	n, err = cdw.conn.bufw.Write(p)
	_, err = cdw.conn.bufw.Write(doubleCrlf)
	if err != nil {
		cdw.conn.rwc.Close()
	}
	cdw.conn.bufw.Flush()
	return n, err
}
