package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"github.com/pkg/errors"
)

// Encrypt use rsa public key to encrypt plaintext
// pubKey: parsed rsa public key
// plaintext: plaintext to encrypt
// return encrypted ciphertext or error
func Encrypt(pubKey *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	if pubKey == nil {
		return nil, errors.New("public key cannot be nil")
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, plaintext)
	if err != nil {
		return nil, errors.Wrap(err, "RSA encryption failed")
	}
	return ciphertext, nil
}

// Decrypt use rsa private key to decrypt ciphertext
// privKey: parsed rsa private key
// ciphertext: ciphertext to decrypt
// return decrypted plaintext or error
func Decrypt(privKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	if privKey == nil {
		return nil, errors.New("private key cannot be nil")
	}

	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, ciphertext)
	if err != nil {
		return nil, errors.Wrap(err, "RSA decryption failed")
	}
	return plaintext, nil
}

// ParsePublicKey parse der encoded pkix public key
// derBytes: der encoded public key
// return parsed rsa public key or error
func ParsePublicKey(derBytes []byte) (*rsa.PublicKey, error) {
	pub, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse public key")
	}

	pubKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.Errorf("expected *rsa.PublicKey, got %T", pub)
	}

	return pubKey, nil
}

// ParsePrivateKey parse pkcs#1 or pkcs#8 format private key
// derBytes: der encoded private key
// return parsed rsa private key or error
func ParsePrivateKey(derBytes []byte) (*rsa.PrivateKey, error) {
	priv, err := x509.ParsePKCS1PrivateKey(derBytes)
	if err == nil {
		return priv, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(derBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse private key")
	}

	privKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.Errorf("expected *rsa.PrivateKey, got %T", key)
	}

	return privKey, nil
}
