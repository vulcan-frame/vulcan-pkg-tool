package id

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodecId(t *testing.T) {
	tests := []struct {
		id int64
	}{
		{id: int64(0)},
		{id: int64(1)},
		{id: int64(2)},
		{id: int64(3)},
		{id: int64(65534)},
		{id: int64(65535)},
		{id: int64(65536)},
		{id: math.MaxInt64},
		{id: math.MaxInt64 - 1},
		{id: int64(-1)},
		{id: -math.MaxInt64},
		{id: -(math.MaxInt64 - 1)},
	}

	for _, tt := range tests {
		str, _ := EncodeId(tt.id)
		id2, _ := DecodeId(str)
		assert.Equal(t, tt.id, id2)
	}
}

// check 5 millions users' id encode str is unique
// func TestIdUnique(t *testing.T) {
// 	v := make(map[string]struct{}, math.MaxInt64)
// 	for id := int64(0); id < 5_000_000; id++ {
// 		str, _ := EncodeId(id)
// 		_, ok := v[str]
// 		assert.False(t, ok)
// 		id2, _ := DecodeId(str)
// 		assert.Equal(t, id, id2)
// 	}
// }

func BenchmarkEncodeId(b *testing.B) {
	id := rand.Int63n(65535)
	for i := 0; i < b.N; i++ {
		EncodeId(id)
	}
}

func BenchmarkDecodeId(b *testing.B) {
	id := rand.Int63n(65535)
	for i := 0; i < b.N; i++ {
		str, _ := EncodeId(id)
		DecodeId(str)
	}
}
