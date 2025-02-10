package logring_test

import "testing"

func BenchmarkBubbleSort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.ReportAllocs()
	}
}
