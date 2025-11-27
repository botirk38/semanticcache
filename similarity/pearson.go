package similarity

import "math"

// PearsonCorrelationSimilarity computes the Pearson correlation coefficient.
// Returns a value between -1 and 1, where 1 means perfect positive correlation.
func PearsonCorrelationSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	n := float64(len(a))

	// Calculate means
	var meanA, meanB float64
	for i := range a {
		meanA += a[i]
		meanB += b[i]
	}
	meanA /= n
	meanB /= n

	// Calculate correlation components
	var numerator, sumSqA, sumSqB float64
	for i := range a {
		diffA := a[i] - meanA
		diffB := b[i] - meanB
		numerator += diffA * diffB
		sumSqA += diffA * diffA
		sumSqB += diffB * diffB
	}

	denominator := math.Sqrt(sumSqA * sumSqB)
	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}
