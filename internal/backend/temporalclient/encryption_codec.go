package temporalclient

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
)

const (
	metadataEncryptionEncoding = "binary/encrypted"
	metadataEncryptionKeyID    = "encryption-key-id"
)

type encryptionCodec struct {
	main    string
	ciphers map[string]cipher.AEAD
}

var _ converter.PayloadCodec = (*encryptionCodec)(nil)

func (c encryptionCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		plain, err := p.Marshal()
		if err != nil {
			return payloads, err
		}

		cipher := c.ciphers[c.main]
		if cipher == nil {
			return nil, fmt.Errorf("cipher %q not found", c.main)
		}

		nonce := make([]byte, cipher.NonceSize())
		if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}

		encrypted := cipher.Seal(nonce, nonce, plain, nil)

		result[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				converter.MetadataEncoding: []byte(metadataEncryptionEncoding),
				metadataEncryptionKeyID:    []byte(c.main),
			},
			Data: encrypted,
		}
	}

	return result, nil
}

func (c encryptionCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))
	for i, p := range payloads {
		// Only if it's encrypted.
		if string(p.Metadata[converter.MetadataEncoding]) != metadataEncryptionEncoding {
			result[i] = p
			continue
		}

		keyName, ok := p.Metadata[metadataEncryptionKeyID]
		if !ok {
			return payloads, errors.New("no encryption key id")
		}

		cipher := c.ciphers[string(keyName)]
		if cipher == nil {
			return nil, fmt.Errorf("cipher %q not found", keyName)
		}

		nonceSize := cipher.NonceSize()
		if len(p.Data) < nonceSize {
			return nil, fmt.Errorf("ciphertext too short, %d < %d bytes", len(p.Data), nonceSize)
		}

		nonce, encrypted := p.Data[:nonceSize], p.Data[nonceSize:]
		plain, err := cipher.Open(nil, nonce, encrypted, nil)
		if err != nil {
			return nil, fmt.Errorf("cipher open: %w", err)
		}

		result[i] = &commonpb.Payload{}
		if err = result[i].Unmarshal(plain); err != nil {
			return payloads, fmt.Errorf("unmarshal: %w", err)
		}
	}

	return result, nil
}
