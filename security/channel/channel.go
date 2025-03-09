package channel

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"github.com/pkg/errors"
	"github.com/vulcan-frame/vulcan-pkg-tool/security/curve25519"
	"golang.org/x/crypto/hkdf"
)

// ECDHKeyPair includes Curve25519 KeyPair
type ECDHKeyPair struct {
	PrivateKey [32]byte
	PublicKey  [32]byte
}

// GenerateKeyPair generates a new Curve25519 key pair
func GenerateKeyPair() (*ECDHKeyPair, error) {
	private, public, err := curve25519.GenerateKeyPair()
	if err != nil {
		return nil, errors.Wrap(err, "generate key pair failed")
	}
	return &ECDHKeyPair{
		PrivateKey: private,
		PublicKey:  public,
	}, nil
}

// DeriveSharedKey derives the encryption key and nonce seed using HKDF
func DeriveSharedKey(sharedSecret []byte) (aesKey []byte, nonceSeed []byte, err error) {
	hkdf := hkdf.New(sha256.New, sharedSecret, nil, []byte("REC_GATESECURE_CHANNEL_V1"))

	combined := make([]byte, 44) // 32字节AES-256密钥 + 12字节Nonce种子
	if _, err := io.ReadFull(hkdf, combined); err != nil {
		return nil, nil, err
	}

	return combined[:32], combined[32:], nil
}

// Encryptor is the encryption structure
type Encryptor struct {
	aesgcm    cipher.AEAD
	nonceSeed []byte
}

// NewEncryptor creates a new encryptor
func NewEncryptor(aesKey []byte, nonceSeed []byte) (*Encryptor, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, errors.Wrap(err, "create aes cipher failed")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "create aes gcm failed")
	}

	return &Encryptor{
		aesgcm:    aesgcm,
		nonceSeed: nonceSeed,
	}, nil
}

// Encrypt encrypts data
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, e.aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, errors.Wrap(err, "generate nonce failed")
	}

	// use random nonce mode (generate new random nonce for each encryption)
	return e.aesgcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decryptor is the decryption structure
type Decryptor struct {
	aesgcm cipher.AEAD
}

// NewDecryptor creates a new decryptor
func NewDecryptor(aesKey []byte) (*Decryptor, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, errors.Wrap(err, "create aes cipher failed")
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "create aes gcm failed")
	}

	return &Decryptor{
		aesgcm: aesgcm,
	}, nil
}

// Decrypt decrypts data
func (d *Decryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < d.aesgcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:d.aesgcm.NonceSize()]
	ciphertext = ciphertext[d.aesgcm.NonceSize():]
	return d.aesgcm.Open(nil, nonce, ciphertext, nil)
}

// EstablishSecureChannel establishes the complete process of establishing a secure channel
func EstablishSecureChannel(localPrivateKey [32]byte, remotePublicKey [32]byte) (*Encryptor, *Decryptor, error) {
	// calculate the shared secret
	sharedSecret, err := curve25519.ComputeSharedSecret(localPrivateKey, remotePublicKey)
	if err != nil {
		return nil, nil, err
	}

	// derive the encryption key and nonce seed
	aesKey, nonceSeed, err := DeriveSharedKey(sharedSecret)
	if err != nil {
		return nil, nil, err
	}

	// create the encryptor and decryptor
	encryptor, err := NewEncryptor(aesKey, nonceSeed)
	if err != nil {
		return nil, nil, err
	}

	decryptor, err := NewDecryptor(aesKey)
	if err != nil {
		return nil, nil, err
	}

	return encryptor, decryptor, nil
}
