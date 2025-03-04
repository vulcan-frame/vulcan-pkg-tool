package rand

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandAlphaNumString(t *testing.T) {
	t.Run("normal cases", func(t *testing.T) {
		// Test generating random strings of different lengths
		lengths := []int{4, 8, 16, 32, 64}
		for _, length := range lengths {
			t.Run("length_"+strconv.Itoa(length), func(t *testing.T) {
				// Test generating random strings of different lengths
				ck := make(map[string]struct{}, 1000)
				for range 1000 {
					s, err := RandAlphaNumString(length)
					assert.Nil(t, err)
					assert.Equal(t, length, len(s))

					// Verify uniqueness for longer strings
					if length >= 16 {
						_, ok := ck[s]
						assert.False(t, ok, "duplicate string generated: %s", s)
						ck[s] = struct{}{}
					}
				}
			})
		}
	})

	t.Run("boundary cases", func(t *testing.T) {
		// Test boundary values
		testCases := []struct {
			name        string
			length      int
			shouldError bool
		}{
			{"zero length", 0, true},
			{"negative length", -1, true},
			{"very large length", 1 << 20, false}, // 1MB length
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				s, err := RandAlphaNumString(tc.length)
				if tc.shouldError {
					assert.Error(t, err)
				} else {
					assert.Nil(t, err)
					assert.Equal(t, tc.length, len(s))
				}
			})
		}
	})
}

func BenchmarkRandAlphaNumString(b *testing.B) {
	benchCases := []struct {
		name   string
		length int
	}{
		{"tiny", 4},
		{"small", 8},
		{"medium", 16},
		{"large", 32},
		{"huge", 64},
		{"massive", 128},
		{"massive", 256},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := RandAlphaNumString(bc.length)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestRandomBytes(t *testing.T) {
	testCases := []struct {
		name   string
		length int
		valid  bool
	}{
		{"normal length", 16, true},
		{"zero length", 0, false},
		{"negative length", -1, false},
		{"large length", 1 << 20, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := RandomBytes(tc.length)

			if tc.valid {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(b) != tc.length {
					t.Errorf("Length mismatch. Expected %d, got %d", tc.length, len(b))
				}
			} else {
				if err == nil {
					t.Error("Expected error but got none")
				}
			}
		})
	}
}

func BenchmarkRandomBytes(b *testing.B) {
	lengths := []int{16, 64, 256}
	for _, length := range lengths {
		b.Run(fmt.Sprintf("%d bytes", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = RandomBytes(length)
			}
		})
	}
}
