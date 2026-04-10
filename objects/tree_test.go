package objects

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"

	"github.com/moges7624/nit/index"
	"github.com/moges7624/nit/repo"
)

func TestTree_Serialize_Sorting(t *testing.T) {
	tree := Tree{
		Entries: []Entry{
			{
				Mode: "100644",
				Name: "zfile.txt",
				Hash: "0000000000000000000000000000000000000000",
			},
			{
				Mode: "100644",
				Name: "afile.txt",
				Hash: "1111111111111111111111111111111111111111",
			},
			{
				Mode: "040000",
				Name: "subdir",
				Hash: "2222222222222222222222222222222222222222",
			},
		},
	}

	got, err := tree.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error: %v", err)
	}

	expectedOrder := []string{"afile.txt", "subdir", "zfile.txt"}

	for i, name := range expectedOrder {
		if !bytes.Contains(got, []byte(name)) {
			t.Errorf("Serialize() missing name %q", name)
		}

		pos := bytes.Index(got, []byte(name))
		if i > 0 && pos < bytes.Index(got, []byte(expectedOrder[i-1])) {
			t.Errorf("Entries not sorted alphabetically by name")
		}
	}
}

func TestTree_Serialize_Format(t *testing.T) {
	blobHashHex := "ce013625030ba8dba906f756967f9e9ca394464a"
	blobHashBin, _ := hex.DecodeString(blobHashHex)

	tree := Tree{
		Entries: []Entry{
			{
				Name: "readme.md",
				Hash: blobHashHex,
				Mode: "100644",
			},
		},
	}

	serialized, err := tree.Serialize()
	if err != nil {
		t.Fatalf("Serliaze() error: %v", err)
	}

	expectedPrefix := []byte("100644 readme.md\x00")
	if !bytes.HasPrefix(serialized, expectedPrefix) {
		t.Errorf("Serliaze() prefix wrong. \nGot: %s\nWant: %s",
			serialized[:len(expectedPrefix)], expectedPrefix)
	}

	hashPart := serialized[len(expectedPrefix):]
	if len(hashPart) != 20 || !bytes.Equal(hashPart, blobHashBin) {
		t.Errorf("Serialized hash bytes wrong.\nGot: %x\nWant: %x",
			hashPart, blobHashBin)
	}
}

func TestTree_Hash(t *testing.T) {
	t.Run("should produce valid tree for single blob", func(t *testing.T) {
		tree := &Tree{
			Entries: []Entry{
				{
					Name: "readme.md",
					Hash: "ce013625030ba8dba906f756967f9e9ca394464a",
					Mode: Regular,
				},
			},
		}

		hash, err := tree.Hash()
		if err != nil {
			t.Fatal("error getting tree hash: ", err.Error())
		}

		expected := "f071d59456a38f47c1bb65edbc1c11b405aed491"
		if hash != expected {
			t.Errorf("Got %s, expected %s", hash, expected)
		}
	})

	t.Run("should produce valid tree for mutliple blobs", func(t *testing.T) {
		tree := &Tree{
			Entries: []Entry{
				{
					Name: "readme.md",
					Hash: "ce013625030ba8dba906f756967f9e9ca394464a",
					Mode: Regular,
				},
				{
					Name: "longtxt.txt",
					Hash: "1741932947564dd278b7ebfe17fd79a4d708a49e",
					Mode: Regular,
				},
			},
		}

		hash, err := tree.Hash()
		if err != nil {
			t.Fatal("error getting tree hash: ", err.Error())
		}

		expected := "cba2cb2c72c46adca56088406df771186d9043ca"
		if hash != expected {
			t.Errorf("Got %s, expected %s", hash, expected)
		}
	})

	t.Run("should produce valid tree for nested dirs", func(t *testing.T) {
		tree := &Tree{
			Entries: []Entry{
				{
					Name: "readme.md",
					Hash: "ce013625030ba8dba906f756967f9e9ca394464a",
					Mode: Regular,
				},
				{
					Name: "longtxt.txt",
					Hash: "1741932947564dd278b7ebfe17fd79a4d708a49e",
					Mode: Regular,
				},
				{
					Name: "internal",
					Hash: "c652b34269663e01105cc210b7ed9a5bc196322f",
					Mode: Directory,
				},
			},
		}

		hash, err := tree.Hash()
		if err != nil {
			t.Fatal("error getting tree hash: ", err.Error())
		}

		expected := "bf8dda0211bafd8e853c532eb1b2d303641f4cc0"
		if hash != expected {
			t.Errorf("Got %s, expected %s", hash, expected)
		}
	})
}

func TestTree_BuildFromIndex(t *testing.T) {
	tests := []struct {
		name     string
		idx      index.Index
		expected string
	}{
		{
			name: "single entry in the index",
			idx: index.Index{
				Entries: map[string]index.Entry{
					"file.txt": { // has conent => "new content"
						Name:    "file.txt",
						ObjHash: "b66ba06d315d46280bb09d54614cc52d1677809f",
						Mode:    uint32(0o100644),
					},
				},
			},
			expected: "49215c458fc9c0cab17387fd7bf069464d34a861",
		},
		{
			name: "nested directories",
			idx: index.Index{
				Entries: map[string]index.Entry{
					"file.txt": { // has conent => "new content". without the quatations
						Name:    "file.txt",
						ObjHash: "b66ba06d315d46280bb09d54614cc52d1677809f",
						Mode:    uint32(0o100644),
					},
					"internal/one.js": { // has empty content
						Name:    "internal/one.js",
						ObjHash: "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
						Mode:    uint32(0o100644),
					},
				},
			},
			expected: "3945d7ecd1b0534cb8760075cae3ea9a892a04c2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			originalWd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			if err = os.Chdir(tempDir); err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				_ = os.Chdir(originalWd)
			})

			repo := repo.NewRepository(tempDir)
			hash, err := BuildFromIndex(repo, tt.idx)
			if err != nil {
				t.Fatal(err)
			}

			if hash != tt.expected {
				t.Errorf("Incorrect tree generated\nExpected: %s\nGot: %s\n", tt.expected, hash)
			}
		})
	}
}
