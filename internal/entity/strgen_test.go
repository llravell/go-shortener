package entity

import "testing"

func BenchmarkRandomStringGenerator(b *testing.B) {
	gen := NewRandomStringGenerator()

	b.ResetTimer()

	for range b.N {
		//nolint:errcheck
		gen.Generate()
	}
}
