package credentials

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"os"

	"github.com/charmbracelet/log"
)

// credentialsHeader contains the file sizes for the cleartext certificate and Macaroon file
type credentialsHeader struct {
	CertLength     int
	MacaroonLength int
}

// newHeader creates a new FileHeader instance
func newHeader(certFileLen, macaroonFileLen int) *credentialsHeader {
	return &credentialsHeader{
		CertLength:     certFileLen,
		MacaroonLength: macaroonFileLen,
	}
}

// Size returns the size of the header in bytes
func (h *credentialsHeader) Size() int {
	return 8 // 2 integers
}

// serialize encodes the header into a byte slice
func (h *credentialsHeader) serialize() []byte {
	headerBytes := make([]byte, h.Size())
	binary.BigEndian.PutUint32(headerBytes[:4], uint32(h.CertLength))
	binary.BigEndian.PutUint32(headerBytes[4:], uint32(h.MacaroonLength))
	return headerBytes
}

// deserialize decodes the byte slice into a Header
func deserialize(data []byte) (*credentialsHeader, error) {
	if len(data) != 8 {
		return nil, errors.New("invalid header size")
	}

	certLength := int(binary.BigEndian.Uint32(data[:4]))
	macaroonLength := int(binary.BigEndian.Uint32(data[4:]))
	return &credentialsHeader{
		CertLength:     certLength,
		MacaroonLength: macaroonLength,
	}, nil
}

func EncryptCredentials(certificatePath string, macaroonPath string) string {
	// Read files
	certData, err := os.ReadFile(certificatePath)
	if err != nil {
		log.Fatal("Error reading TLS Certificate:", err)
	}

	macData, err := os.ReadFile(macaroonPath)
	if err != nil {
		log.Fatal("Error reading macaroon file:", err)
	}

	// Create a header with file lengths
	header := newHeader(len(certData), len(macData))
	headerBytes := header.serialize()
	log.Info("Header: " + hex.EncodeToString(headerBytes))

	concatenatedData := append(certData, macData...)
	generatedKey := GenerateKey()
	encryptedData, err := Encrypt(generatedKey, concatenatedData)
	if err != nil {
		log.Fatal("Error encrypting key material:", err)
	}

	encryptedDataWithHeader := append(headerBytes, encryptedData...)

	// Write encrypted data with header to a file
	err = os.WriteFile("auth.bin", encryptedDataWithHeader, 0644)
	if err != nil {
		log.Fatal("Error writing to file:", err)
	}

	return generatedKey

}
