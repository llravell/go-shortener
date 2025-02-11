package entity

import "testing"

func BenchmarkRandomStringGenerator(b *testing.B) {
	gen := NewRandomStringGenerator()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		//nolint:errcheck
		gen.Generate()
	}
}
