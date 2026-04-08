package objects

import (
	"bytes"
	"encoding/hex"
	"testing"
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
