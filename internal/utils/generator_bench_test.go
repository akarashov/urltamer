package utils

import "testing"

func BenchmarkGenerateShortURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateShortURL()
	}
}
