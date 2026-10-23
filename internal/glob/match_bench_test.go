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
