package objects

import (
	"testing"
)

func TestBlob_Hash(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "empty blob",
			data:     []byte{},
			expected: "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
		},
		{
			name:     "hello",
			data:     []byte("hello\n"),
			expected: "ce013625030ba8dba906f756967f9e9ca394464a",
		},
		{
			name:     "mutltiline text",
			data:     []byte("This is a long text of random characters.   2232  23222dfd sdfi I skdfjo j\nasdfdsfosdfdsfsdf\njdfsdjfkskd499494994 afsfd <<<<<<<<>>>> {{}\n{{\n|';[[[[\n"),
			expected: "1741932947564dd278b7ebfe17fd79a4d708a49e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBlob(tt.data)

			hash, err := b.Hash()
			if err != nil {
				t.Fatalf("Hash() error: %v", err)
			}

			if hash != tt.expected {
				t.Errorf("Hash() = %s, want %s", hash, tt.expected)
			}
		})
	}
}

func TestBlob_Hash_Caching(t *testing.T) {
	b := NewBlob([]byte("hello"))

	hash1, err := b.Hash()
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	b.Data = []byte("hello there")
	hash2, err := b.Hash()
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	if hash1 != hash2 {
		t.Error("Hash changed after modification- caching is not working")
	}
}
