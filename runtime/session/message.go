package session

import (
	/*"encoding/base64"
	"encoding/gob"
	"bytes"*/
)

type T interface {
	
}

// cf. org.scribble.runtime.net.ScribMessage
type ScribMessage interface {
	GetOp() string
}

// cf. org.scribble.runtime.net.ScribMessageFormatter
type ScribMessageFormatter interface {
	ToBytes(*ScribMessage) ([]byte, error)
	FromBytes([]byte) (*ScribMessage, error)  // Or bytes.Buffer?
}

type ScribDefaultFormatter struct {
	//enc gob.Encoder
}

/*func NewScribDefaultFormatter() *ScribDefaultFormatter {
	return ScribDefaultFormatter{enc: gob.NewEncoder(&b)}
}*/

func (dfmt *ScribDefaultFormatter) ToBytes(m *ScribMessage) ([]byte, error) {
	return []byte{}, nil	

	/*b := bytes.Buffer{}
	err := dfmt.enc.Encode(m)
	//if err != nil { fmt.Println(`failed gob Encode`, err) }
	return b.Bytes(), err*/
}

func (dfmt *ScribDefaultFormatter) FromBytes(bs []byte) (*ScribMessage, error) {
	return nil, nil
	
	/*...var m ScribMessage
	b := bytes.Buffer{}
	b.Write(bs)
	d := gob.NewDecoder(&b)  // Have to create each time?
	err = d.Decode(&m)
	return m, err*/
}

/*// go binary encoder
func ToGOB64(m SX) string {
    b := bytes.Buffer{}
    e := gob.NewEncoder(&b)
    err := e.Encode(m)
    if err != nil { fmt.Println(`failed gob Encode`, err) }
    return base64.StdEncoding.EncodeToString(b.Bytes())
}

// go binary decoder
func FromGOB64(str string) SX {
    m := SX{}
    by, err := base64.StdEncoding.DecodeString(str)
    if err != nil { fmt.Println(`failed base64 Decode`, err); }
    b := bytes.Buffer{}
    b.Write(by)
    d := gob.NewDecoder(&b)
    err = d.Decode(&m)
    if err != nil { fmt.Println(`failed gob Decode`, err); }
    return m
}

func init() {
    gob.Register(SX{})
    gob.Register(Session{}) 
}
*/
