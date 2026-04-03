package objects

import (
	"bytes"
	"testing"
)

func TestBlob_Hash(t *testing.T) {
	t.Run("should produce valid hash", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte("hello\n"))
		blob := NewBlob([]byte(buf.Bytes()))

		expected := "ce013625030ba8dba906f756967f9e9ca394464a"
		hash, err := blob.Hash()
		if err != nil {
			t.Errorf("error getting blob hash %s", err.Error())
		}

		if blob.hash != expected {
			t.Errorf("Got %s, expected %s", hash, expected)
		}
	})

	t.Run("should produce valid hash long text", func(t *testing.T) {
		data := "This is a long text of random characters.   2232  23222dfd sdfi I skdfjo j\nasdfdsfosdfdsfsdf\njdfsdjfkskd499494994 afsfd <<<<<<<<>>>> {{}\n{{\n|';[[[[\n"

		blob := NewBlob([]byte(data))
		hash, err := blob.Hash()
		if err != nil {
			t.Errorf("error getting blob hash %s", err.Error())
		}
		expected := "1741932947564dd278b7ebfe17fd79a4d708a49e"

		if blob.hash != expected {
			t.Errorf("Got %s, expected %s", hash, expected)
		}
	})
}
