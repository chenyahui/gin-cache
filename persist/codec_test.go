package persist

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type testStruct struct {
	A int
	B string
	C *int
}

func TestCodec(t *testing.T) {
	src := &testStruct{
		A: 1,
		B: "2",
	}

	payload, err := Serialize(src)
	require.Nil(t, err)
	require.True(t, len(payload) > 0)

	var dest testStruct
	err = Deserialize(payload, &dest)
	require.Nil(t, err)

	assert.Equal(t, src.A, dest.A)
	assert.Equal(t, src.B, dest.B)
	assert.Equal(t, src.C, dest.C)
}
