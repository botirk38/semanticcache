package similarity

// DotProductSimilarity computes the dot product between two vectors.
// No normalization is applied, so results depend on vector magnitudes.
func DotProductSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot float64
	for i := range a {
		dot += a[i] * b[i]
	}

	return dot
}
