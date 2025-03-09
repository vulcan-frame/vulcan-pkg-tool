package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	vrand "github.com/vulcan-frame/vulcan-pkg-tool/rand"
	"golang.org/x/crypto/curve25519"
)

func TestGenCurve25519Key(t *testing.T) {
	aKey, bKey := genCurve25519Key()
	assert.Equal(t, aKey, bKey)
}

func TestRSAEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		keyBits   int
		plaintext []byte
	}{
		{
			name:      "short text with 2048 bits key",
			keyBits:   2048,
			plaintext: []byte("Hello, World!"),
		},
		{
			name:      "longer text with 4096 bits key",
			keyBits:   4096,
			plaintext: []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, pri, pubBytes, _, err := generateTestKeyPair(tt.keyBits)
			assert.NoError(t, err)

			pub, err := ParsePublicKey(pubBytes)
			assert.NoError(t, err)

			// Test encryption
			encrypted, err := Encrypt(pub, tt.plaintext)
			assert.NoError(t, err)
			assert.NotEmpty(t, encrypted)

			// Test decryption
			decrypted, err := Decrypt(pri, encrypted)
			assert.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestRSAKeyMarshaling(t *testing.T) {
	// Generate test key pair
	_, pri, pubBytes, priBytes, err := generateTestKeyPair(4096)
	assert.NoError(t, err)

	// Test public key unmarshaling
	pub, err := x509.ParsePKIXPublicKey(pubBytes)
	assert.NoError(t, err)

	// Test private key unmarshaling
	pri2, err := x509.ParsePKCS8PrivateKey(priBytes)
	assert.NoError(t, err)

	// Test marshaling back
	pubBytes2, err := x509.MarshalPKIXPublicKey(pub)
	assert.NoError(t, err)
	priBytes2, err := x509.MarshalPKCS8PrivateKey(pri2)
	assert.NoError(t, err)

	// Verify marshaled bytes are identical
	assert.Equal(t, pubBytes, pubBytes2)
	assert.Equal(t, priBytes, priBytes2)

	// Verify keys functionality
	testData := []byte("test encryption after marshaling")
	encrypted, err := Encrypt(pub.(*rsa.PublicKey), testData)
	assert.NoError(t, err)

	decrypted, err := Decrypt(pri, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, testData, decrypted)
}

func TestRSASignVerify(t *testing.T) {
	_, pri, _, _, err := generateTestKeyPair(2048)
	assert.NoError(t, err)

	pub := &pri.PublicKey

	testData := []byte("data to sign")
	hashed := sha256.Sum256(testData)

	// Test signing
	signature, err := rsa.SignPKCS1v15(rand.Reader, pri, crypto.SHA256, hashed[:])
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	// Test verification
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], signature)
	assert.NoError(t, err)

	// Test verification with modified data
	hashedModified := sha256.Sum256([]byte("modified data"))
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashedModified[:], signature)
	assert.Error(t, err)
}

func TestCurve25519KeyExchange(t *testing.T) {
	aKey, bKey := genCurve25519Key()
	assert.NotEmpty(t, aKey)
	assert.NotEmpty(t, bKey)
	assert.Equal(t, aKey, bKey, "Key exchange failed: keys don't match")

	// Test multiple key exchanges
	for i := 0; i < 5; i++ {
		aKey2, bKey2 := genCurve25519Key()
		assert.NotEmpty(t, aKey2)
		assert.NotEmpty(t, bKey2)
		assert.Equal(t, aKey2, bKey2)
		// Verify different key exchanges produce different keys
		assert.NotEqual(t, aKey, aKey2)
	}
}

func BenchmarkRSAEncrypt(b *testing.B) {
	pub, _, _, _, err := generateTestKeyPair(4096)
	assert.NoError(b, err)

	data, _ := vrand.RandAlphaNumString(256)
	org := []byte(data)

	for i := 0; i < b.N; i++ {
		_, _ = Encrypt(pub, org)
	}
}

func BenchmarkRSADecrypt(b *testing.B) {
	pub, pri, _, _, err := generateTestKeyPair(4096)
	assert.NoError(b, err)

	data, _ := vrand.RandAlphaNumString(30)
	org := []byte(data)

	dst, _ := rsa.EncryptPKCS1v15(rand.Reader, pub, org)
	for i := 0; i < b.N; i++ {
		_, _ = Decrypt(pri, dst)
	}
}

func BenchmarkGenRsaKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _, _, err := generateTestKeyPair(4096)
		assert.NoError(b, err)
	}
}

func BenchmarkGenCurve25519Key(b *testing.B) {
	for i := 0; i < b.N; i++ {
		genCurve25519Key()
	}
}

func genCurve25519Key() (aKey, bKey []byte) {
	var aPri, aPub [32]byte
	_, _ = io.ReadFull(rand.Reader, aPri[:])
	curve25519.ScalarBaseMult(&aPub, &aPri)

	var bPri, bPub [32]byte
	_, _ = io.ReadFull(rand.Reader, bPri[:])
	curve25519.ScalarBaseMult(&bPub, &bPri)

	aKey, _ = curve25519.X25519(aPri[:], bPub[:])
	bKey, _ = curve25519.X25519(bPri[:], aPub[:])

	return
}

func generateTestKeyPair(bits int) (*rsa.PublicKey, *rsa.PrivateKey, []byte, []byte, error) {
	pri, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	pub := &pri.PublicKey
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	priBytes, err := x509.MarshalPKCS8PrivateKey(pri)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return pub, pri, pubBytes, priBytes, nil
}
