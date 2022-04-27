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

// Deserialize will deserialize the passed []byte into the passed ptr interface{}
func Deserialize(payload []byte, ptr interface{}) (err error) {
	return gob.NewDecoder(bytes.NewBuffer(payload)).Decode(ptr)
}
