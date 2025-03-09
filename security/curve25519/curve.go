package curve25519

import (
	"crypto/rand"

	"github.com/pkg/errors"
	"golang.org/x/crypto/curve25519"
)

// GenerateKeyPair generate curve25519 key pair
// return private key and public key 32 bytes array or error
func GenerateKeyPair() (privateKey [32]byte, publicKey [32]byte, err error) {
	_, err = rand.Read(privateKey[:])
	if err != nil {
		return [32]byte{}, [32]byte{}, errors.Wrap(err, "failed to generate random private key")
	}
	curve25519.ScalarBaseMult(&publicKey, &privateKey)
	return privateKey, publicKey, nil
}

// PublicKeyToBytes convert public key to bytes slice
func PublicKeyToBytes(publicKey *[32]byte) []byte {
	return publicKey[:]
}

// ParsePublicKey parse bytes slice to curve25519 public key
func ParsePublicKey(b []byte) ([32]byte, error) {
	if len(b) != 32 {
		return [32]byte{}, errors.New("invalid public key length")
	}
	var publicKey [32]byte
	copy(publicKey[:], b)
	return publicKey, nil
}

// ComputeSharedSecret compute shared secret
func ComputeSharedSecret(privateKey [32]byte, publicKey [32]byte) ([]byte, error) {
	secret, err := curve25519.X25519(privateKey[:], publicKey[:])
	if err != nil {
		return nil, errors.Wrap(err, "failed to compute shared secret")
	}
	return secret, nil
}
