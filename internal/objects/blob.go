package objects

import (
	"crypto/sha1"
	"fmt"
)

type Blob struct {
	Data []byte
	hash string
}

func NewBlob(data []byte) *Blob {
	return &Blob{
		Data: data,
	}
}

func (b *Blob) Type() string {
	return "blob"
}

func (b *Blob) Serialize() ([]byte, error) {
	return b.Data, nil
}

func (b *Blob) Hash() (string, error) {
	if b.hash != "" {
		return b.hash, nil
	}

	data, _ := b.Serialize()
	header := fmt.Sprintf("%s %d\x00", b.Type(), len(data))
	fullContent := append([]byte(header), data...)

	hash := sha1.Sum(fullContent)
	hashHex := fmt.Sprintf("%x", hash)

	b.hash = hashHex
	return hashHex, nil
}
