package rand

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/pkg/errors"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var charsetLen = big.NewInt(int64(len(charset)))

func RandAlphaNumString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than 0")
	}

	var buf bytes.Buffer
	buf.Grow(length)

	randomBytes := make([]byte, length)
	for range randomBytes {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", errors.Wrap(err, "rand int failed")
		}
		buf.WriteByte(charset[idx.Int64()])
	}

	return buf.String(), nil
}

func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return b, nil
}
