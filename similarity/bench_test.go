package similarity

import "testing"

func benchVecs(dim int) ([]float64, []float64) {
	a := make([]float64, dim)
	b := make([]float64, dim)
	for i := range a {
		a[i] = float64(i) / float64(dim)
		b[i] = float64(dim-i) / float64(dim)
	}
	return a, b
}

func BenchmarkCosineSimilarity_1536(b *testing.B) {
	a, bv := benchVecs(1536)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CosineSimilarity(a, bv)
	}
}

func BenchmarkEuclideanSimilarity_1536(b *testing.B) {
	a, bv := benchVecs(1536)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EuclideanSimilarity(a, bv)
	}
}

func BenchmarkDotProductSimilarity_1536(b *testing.B) {
	a, bv := benchVecs(1536)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DotProductSimilarity(a, bv)
	}
}

func BenchmarkManhattanSimilarity_1536(b *testing.B) {
	a, bv := benchVecs(1536)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ManhattanSimilarity(a, bv)
	}
}

func BenchmarkPearsonCorrelationSimilarity_1536(b *testing.B) {
	a, bv := benchVecs(1536)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PearsonCorrelationSimilarity(a, bv)
	}
}
