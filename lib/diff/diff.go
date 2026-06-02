package diff

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
