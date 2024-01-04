package credentials

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentialsEncryption(t *testing.T) {
	// Generate cleartext
	certData := make([]byte, 128)
	macData := make([]byte, 540)
	_, err := rand.Read(certData)
	if err != nil {
		assert.FailNow(t, "error generating random cert data")
	}
	_, err = rand.Read(macData)
	if err != nil {
		assert.FailNow(t, "error generating random macaroon data")
	}

	key := GenerateKey()

	encryptedData := encryptData(certData, macData, key)

	headerBytes := encryptedData[:8]
	header, _ := deserializeHeader(headerBytes)

	assert.Equal(t, 128, header.CertLength)
	assert.Equal(t, 540, header.MacaroonLength)

	assert.NotEqual(t, encryptedData[8:header.CertLength], certData)
	assert.NotEqual(t, encryptedData[8+header.CertLength:], macData)

	decryptedCertData, decryptedMacData := decryptData(encryptedData, key)

	assert.Equal(t, certData, decryptedCertData)
	assert.Equal(t, macData, decryptedMacData)
}
