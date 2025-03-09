package aes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vulcan-frame/vulcan-pkg-tool/rand"
)

var (
	org     = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	empty   = []byte("")
	utf8    = []byte("测试中文加密解密")
	special = []byte("!@#$%^&*()_+-=[]{}|;:,.<>?")
)

func TestAESCBCCodec(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "normal ascii text",
			input:   org,
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   empty,
			wantErr: true,
		},
		{
			name:    "chinese characters",
			input:   utf8,
			wantErr: false,
		},
		{
			name:    "special characters",
			input:   special,
			wantErr: false,
		},
	}

	data, err := rand.RandAlphaNumString(32)
	fmt.Println(data)
	assert.Nil(t, err)
	aesKey := []byte(data)
	aesBlock, err := NewBlock(aesKey)
	assert.Nil(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := Encrypt(aesKey, aesBlock, tt.input)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)

			// Decrypt
			decrypted, err := Decrypt(aesKey, aesBlock, encrypted)
			assert.Nil(t, err)
			assert.Equal(t, tt.input, decrypted)
		})
	}
}

func TestInvalidInputs(t *testing.T) {
	// Test with invalid key length
	invalidKey := []byte("too short")
	_, err := NewBlock(invalidKey)
	assert.NotNil(t, err)

	// Test with nil inputs
	validKey, _ := rand.RandAlphaNumString(32)
	block, _ := NewBlock([]byte(validKey))

	_, err = Encrypt(nil, block, org)
	assert.NotNil(t, err)

	_, err = Encrypt([]byte(validKey), nil, org)
	assert.NotNil(t, err)

	_, err = Encrypt([]byte(validKey), block, nil)
	assert.NotNil(t, err)
}

func BenchmarkAESCBCEncrypt(b *testing.B) {
	data, _ := rand.RandAlphaNumString(32)
	key := []byte(data)
	block, _ := NewBlock(key)

	for i := 0; i < b.N; i++ {
		if _, err := Encrypt(key, block, org); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAESCBCDecrypt(b *testing.B) {
	data, _ := rand.RandAlphaNumString(32)
	key := []byte(data)
	block, _ := NewBlock(key)
	ser, _ := Encrypt(key, block, org)

	for i := 0; i < b.N; i++ {
		if _, err := Decrypt(key, block, ser); err != nil {
			b.Fatal(err)
		}
	}
}
