package persist

import (
	"bytes"
	"encoding/gob"
)

// Serialize returns a []byte representing the passed value
func Serialize(value interface{}) ([]byte, error) {
	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Deserialize deserialices the passed []byte into a the passed ptr interface{}
func Deserialize(byt []byte, ptr interface{}) (err error) {
	b := bytes.NewBuffer(byt)
	decoder := gob.NewDecoder(b)
	if err = decoder.Decode(ptr); err != nil {
		return err
	}
	return nil
}
