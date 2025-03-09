package curve25519

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyExchange(t *testing.T) {
	serverPrivate, serverPublic, err := GenerateKeyPair()
	assert.NoError(t, err)

	clientPrivate, clientPublic, err := GenerateKeyPair()
	assert.NoError(t, err)

	serverSecret, err := ComputeSharedSecret(serverPrivate, clientPublic)
	assert.NoError(t, err)

	clientSecret, err := ComputeSharedSecret(clientPrivate, serverPublic)
	assert.NoError(t, err)

	assert.Equal(t, serverSecret, clientSecret)
}

func TestInvalidPublicKey(t *testing.T) {
	invalidKey := make([]byte, 31)
	_, err := rand.Read(invalidKey)
	assert.NoError(t, err)

	_, err = ParsePublicKey(invalidKey)
	assert.Error(t, err)
}

func BenchmarkKeyGeneration(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, err := GenerateKeyPair()
		if err != nil {
			b.Fatalf("key generation failed: %v", err)
		}
	}
}

func BenchmarkSharedSecretComputation(b *testing.B) {
	// pre-generate fixed key pair
	serverPriv, serverPub, _ := GenerateKeyPair()
	clientPriv, clientPub, _ := GenerateKeyPair()

	// reset timer
	b.ResetTimer()
	b.ReportAllocs()

	// parallel test
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// test two direction key computation
			if _, err := ComputeSharedSecret(serverPriv, clientPub); err != nil {
				b.Fatal(err)
			}
			if _, err := ComputeSharedSecret(clientPriv, serverPub); err != nil {
				b.Fatal(err)
			}
		}
	})
}

var sink interface{} // prevent compiler optimize

func BenchmarkKeyGenerationAndExchange(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// generate server key
		serverPriv, _, err := GenerateKeyPair()
		if err != nil {
			b.Fatal(err)
		}

		// generate client key
		_, clientPub, err := GenerateKeyPair()
		if err != nil {
			b.Fatal(err)
		}

		// compute shared secret
		secret, err := ComputeSharedSecret(serverPriv, clientPub)
		if err != nil {
			b.Fatal(err)
		}

		// prevent compiler optimize
		sink = secret
	}
}
