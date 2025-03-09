package channel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFullCommunicationFlow(t *testing.T) {
	// generate the server key pair
	serverKeyPair, err := GenerateKeyPair()
	require.NoError(t, err)

	// generate the client key pair
	clientKeyPair, err := GenerateKeyPair()
	require.NoError(t, err)

	// establish the secure channel (use the server private key and client public key)
	serverEncryptor, serverDecryptor, err := EstablishSecureChannel(
		serverKeyPair.PrivateKey,
		clientKeyPair.PublicKey,
	)
	require.NoError(t, err)

	// establish the secure channel (use the client private key and server public key)
	clientEncryptor, clientDecryptor, err := EstablishSecureChannel(
		clientKeyPair.PrivateKey,
		serverKeyPair.PublicKey,
	)
	require.NoError(t, err)

	// test the communication from client to server
	originalMessage := []byte("Hello Secure World!")

	// client encrypts
	encrypted, err := clientEncryptor.Encrypt(originalMessage)
	require.NoError(t, err)

	// server decrypts
	decrypted, err := serverDecryptor.Decrypt(encrypted)
	require.NoError(t, err)
	require.Equal(t, originalMessage, decrypted)

	// test the communication from server to client
	serverMessage := []byte("Hello from Server!")

	// server encrypts
	encryptedServer, err := serverEncryptor.Encrypt(serverMessage)
	require.NoError(t, err)

	// client decrypts
	decryptedClient, err := clientDecryptor.Decrypt(encryptedServer)
	require.NoError(t, err)
	require.Equal(t, serverMessage, decryptedClient)

	// test the tampering detection
	if len(encrypted) > 0 {
		encrypted[0] ^= 0xFF // modify the first byte
		_, err = serverDecryptor.Decrypt(encrypted)
		require.Error(t, err)
	}
}
