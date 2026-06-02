package diff

type OpType int

const (
	Equal OpType = iota
	Delete
	Insert
)

type Operation struct {
	Type OpType
	Line string
}

func myersDiff(a, b []string) []Operation {
	N := len(a)
	M := len(b)
	MAX := N + M

	v := make([]int, 2*MAX+1)

	var traces [][]int

	found := false
	var dStep int

	for d := 0; d <= MAX; d++ {
		vCopy := make([]int, len(v))
		copy(vCopy, v)
		traces = append(traces, vCopy)

		for k := -d; k <= d; k += 2 {
			kIdx := k + MAX
			var x int

			if k == -d || (k != d && v[kIdx-1] < v[kIdx+1]) {
				x = v[kIdx+1]
			} else {
				x = v[kIdx-1] + 1
			}

			y := x - k

			for x < N && y < M && a[x] == b[y] {
				x++
				y++
			}

			v[kIdx] = x

			if x >= N && y >= M {
				found = true
				dStep = d
				break
			}
		}
		if found {
			break
		}
	}

	return backtrack(traces, a, b, dStep, N, M, MAX)
}

func backtrack(traces [][]int, a, b []string, dStep, N, M, MAX int) []Operation {
	var script []Operation
	x, y := N, M

	for d := dStep; d > 0; d-- {
		v := traces[d]
		k := x - y
		kIdx := k + MAX

		var prevK int
		if k == -d || (k != d && v[kIdx-1] < v[kIdx+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}

		prevX := v[prevK+MAX]
		prevY := prevX - prevK

		for x > prevX && y > prevY {
			script = append([]Operation{{Type: Equal, Line: a[x-1]}}, script...)
			x--
			y--
		}

		if x > prevX {
			script = append([]Operation{{Type: Delete, Line: a[x-1]}}, script...) // Horizontal
		} else if y > prevY {
			script = append([]Operation{{Type: Insert, Line: b[y-1]}}, script...) // Vertical
		}

		x, y = prevX, prevY
	}

	for x > 0 && y > 0 {
		script = append([]Operation{{Type: Equal, Line: a[x-1]}}, script...)
		x--
		y--
	}

	return script
}
