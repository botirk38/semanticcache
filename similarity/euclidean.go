package similarity

import "math"

// EuclideanSimilarity computes similarity based on Euclidean distance.
// Returns 1 / (1 + distance) to convert distance to similarity (higher = more similar).
// Result is always between 0 and 1, where 1 means identical vectors.
func EuclideanSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var sum float32
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	distance := float32(math.Sqrt(float64(sum)))
	return 1 / (1 + distance)
}
