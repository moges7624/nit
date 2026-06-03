package diff

import (
	"fmt"
	"strings"

	"github.com/moges7624/nit/internal/objects"
)

func PrintDiffContent(a, b *objects.Blob, fileName string, mode uint32) {
	aHash, _ := a.Hash()
	bHash, _ := b.Hash()

	if aHash == bHash {
		return
	}

	// var buf bytes.Buffer

	fmt.Printf("diff --nit a/%s b/%[1]s\n", fileName)
	fmt.Printf("index %s..%s %o\n", aHash[:7], bHash[:7], mode)
	fmt.Printf("--- a/%s\n", fileName)
	fmt.Printf("+++ b/%s\n", fileName)

	aArr := strings.Split(strings.TrimSpace(string(a.Data)), "\n")
	bArr := strings.Split(strings.TrimSpace(string(b.Data)), "\n")
	ses := myersDiff(aArr, bArr)

	const (
		Reset = "\033[0m"
		Red   = "\033[31m"
		Green = "\033[32m"
	)

	for _, op := range ses {
		switch op.Type {
		case Equal:
			fmt.Printf("  %s\n", op.Line)
		case Delete:
			fmt.Printf("%s- %s%s\n", Red, op.Line, Reset)
		case Insert:
			fmt.Printf("%s+ %s%s\n", Green, op.Line, Reset)
		}
	}
}

type DiffStat struct {
	Equal     int
	Deletion  int
	Insertion int
}

func Stat(a, b []string) DiffStat {
	var eq int
	var del int
	var ins int

	if len(a) == 0 {
		return DiffStat{
			Equal:     0,
			Deletion:  0,
			Insertion: len(b),
		}
	}

	script := myersDiff(a, b)
	for _, k := range script {
		switch k.Type {
		case Equal:
			eq++
		case Delete:
			del++
		case Insert:
			ins++
		}
	}

	return DiffStat{
		Equal:     eq,
		Deletion:  del,
		Insertion: ins,
	}
}
