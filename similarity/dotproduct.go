package similarity

// DotProductSimilarity computes the dot product between two vectors.
// No normalization is applied, so results depend on vector magnitudes.
func DotProductSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot float32
	for i := range a {
		dot += a[i] * b[i]
	}

	return dot
}
