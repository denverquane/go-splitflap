package utils

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func Test_Crc32(t *testing.T) {
	encoded := CreatePayloadWithCRC32Checksum([]byte{1, 2, 3})

	decoded, ok := ParseCRC32EncodedPayload(encoded)

	assert.Equal(t, true, ok)
	assert.Equal(t, []byte{1, 2, 3}, decoded)
}
