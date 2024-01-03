package credentials

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeyGeneration_assertKeyReturned(t *testing.T) {
	keyset := GenerateKey()

	assert.True(t, len(keyset) > 10)
}

func TestEncryption(t *testing.T) {
	keyset := GenerateKey()
	cleartext := make([]byte, 128)

	// Encrypt data
	ciphertext, err := Encrypt(keyset, cleartext)
	if err != nil {
		assert.FailNow(t, "unable to encrypt data: "+err.Error())
	}

	assert.NotNil(t, ciphertext)

	// Decrypt data
	decryptedData, err := Decrypt(keyset, ciphertext)
	if err != nil {
		assert.FailNow(t, "unable to decrypt data: "+err.Error())
	}

	assert.Equal(t, cleartext, decryptedData)
}
