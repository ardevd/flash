package credentials

import (
	"bytes"
	"encoding/hex"

	"github.com/charmbracelet/log"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/insecurecleartextkeyset"
	"github.com/google/tink/go/keyset"
)

// Generate a AES256GCM keyset handle.
func GenerateKey() string {
	kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	if err != nil {
		log.Fatal(err)
	}


	// Create a writer to store the keyset
	buf := new(bytes.Buffer)
	writer := keyset.NewBinaryWriter(buf)

	err = insecurecleartextkeyset.Write(kh, writer)
	if err != nil {
		log.Fatal(err)
	}

	// Encode the serialized keyset from the buffer to a base64 string
	base64EncodedKey := hex.EncodeToString(buf.Bytes())

	log.Info("Key succesfully generated")
	return base64EncodedKey
}

func parseEncodedKeyset(encodedKeyset string) *keyset.Handle{
	encodedKeysetBytes, err := hex.DecodeString(encodedKeyset)
	if err != nil {
		log.Fatal(err)
	}

	reader := keyset.NewBinaryReader(bytes.NewReader(encodedKeysetBytes))
	kh, err := insecurecleartextkeyset.Read(reader)
	if err != nil {
		log.Fatal(err)
	}

	return kh
}

// Encrypt data
func Encrypt(encodedKeyset string, data []byte) ([]byte, error) {
	kh := parseEncodedKeyset(encodedKeyset)
	a, err := aead.New(kh)
	if err != nil {
		log.Fatal(err)
	}
	return a.Encrypt(data, nil)
}

// Decrypt a ciphertext with a provided keyset.
func Decrypt(encodedKeyset string, ciphertext []byte) ([]byte, error) {
	kh := parseEncodedKeyset(encodedKeyset)
	a, err := aead.New(kh)
	if err != nil {
		log.Fatal(err)
	}
	return a.Decrypt(ciphertext, nil)
}
