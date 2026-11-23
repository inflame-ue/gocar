package glob

import "testing"

var sink bool

func BenchmarkMatch(b *testing.B) {
	pattern := []string{"**", "**", "**", "**", "**", "**", "**", "**", "**",
		"**", "**", "**", "**", "**", "**", "main.go"}
	path := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "1",
		"2", "3", "4", "5", "6", "main.py"}

	b.Run("memoized", func(b *testing.B) {
		b.ReportAllocs()
		var ok bool
		for i := 0; i < b.N; i++ {
			ok, _ = match(pattern, path)
		}
		sink = ok
	})

	b.Run("naive", func(b *testing.B) {
		b.ReportAllocs()
		var ok bool
		for i := 0; i < b.N; i++ {
			ok, _ = naiveMatch(pattern, path)
		}
		sink = ok
	})
}

// naiveMatch is the exponential baseline used by BenchmarkMatch.
func naiveMatch(pattern, path []string) (bool, error) {
	if len(pattern) == 0 {
		return len(path) == 0, nil
	}

	if pattern[0] == "**" {
		if ok, err := naiveMatch(pattern[1:], path); err != nil || ok {
			return ok, err
		}

		if len(path) > 0 {
			return naiveMatch(pattern, path[1:])
		}

		return false, nil
	}

	if len(path) == 0 {
		return false, nil
	}

	segMatched, err := segMatch(pattern[0], path[0])
	if err != nil {
		return false, err
	}

	matched, err := naiveMatch(pattern[1:], path[1:])
	if err != nil {
		return false, err
	}

	return segMatched && matched, nil
}
