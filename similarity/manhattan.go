package similarity

import "math"

// ManhattanSimilarity computes similarity based on Manhattan (L1) distance.
// Returns 1 / (1 + distance) to convert distance to similarity.
func ManhattanSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var sum float32
	for i := range a {
		sum += float32(math.Abs(float64(a[i] - b[i])))
	}

	return 1 / (1 + sum)
}
