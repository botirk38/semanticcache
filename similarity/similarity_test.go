package similarity

import (
	"math"
	"testing"
)

// Test similarity functions with known vectors
func TestSimilarityFunctions(t *testing.T) {
	// Test vectors
	vec1 := []float64{1, 0, 0}
	vec2 := []float64{0, 1, 0}
	vec3 := []float64{1, 0, 0} // Same as vec1

	t.Run("CosineSimilarity", func(t *testing.T) {
		// Test orthogonal vectors (should be 0)
		sim := CosineSimilarity(vec1, vec2)
		if sim != 0 {
			t.Errorf("Expected 0, got %f", sim)
		}

		// Test identical vectors (should be 1)
		sim = CosineSimilarity(vec1, vec3)
		if math.Abs(float64(sim-1)) > 0.001 {
			t.Errorf("Expected 1, got %f", sim)
		}

		// Test empty vectors
		sim = CosineSimilarity([]float64{}, []float64{})
		if sim != 0 {
			t.Errorf("Expected 0 for empty vectors, got %f", sim)
		}

		// Test different length vectors
		sim = CosineSimilarity(vec1, []float64{1, 0})
		if sim != 0 {
			t.Errorf("Expected 0 for different length vectors, got %f", sim)
		}
	})

	t.Run("EuclideanSimilarity", func(t *testing.T) {
		// Test identical vectors (should be 1)
		sim := EuclideanSimilarity(vec1, vec3)
		if sim != 1 {
			t.Errorf("Expected 1, got %f", sim)
		}

		// Test different vectors (should be less than 1)
		sim = EuclideanSimilarity(vec1, vec2)
		if sim >= 1 {
			t.Errorf("Expected < 1, got %f", sim)
		}

		// Test empty vectors
		sim = EuclideanSimilarity([]float64{}, []float64{})
		if sim != 0 {
			t.Errorf("Expected 0 for empty vectors, got %f", sim)
		}
	})

	t.Run("DotProductSimilarity", func(t *testing.T) {
		// Test orthogonal vectors (should be 0)
		sim := DotProductSimilarity(vec1, vec2)
		if sim != 0 {
			t.Errorf("Expected 0, got %f", sim)
		}

		// Test identical unit vectors (should be 1)
		sim = DotProductSimilarity(vec1, vec3)
		if sim != 1 {
			t.Errorf("Expected 1, got %f", sim)
		}
	})

	t.Run("ManhattanSimilarity", func(t *testing.T) {
		// Test identical vectors (should be 1)
		sim := ManhattanSimilarity(vec1, vec3)
		if sim != 1 {
			t.Errorf("Expected 1, got %f", sim)
		}

		// Test different vectors (should be less than 1)
		sim = ManhattanSimilarity(vec1, vec2)
		if sim >= 1 {
			t.Errorf("Expected < 1, got %f", sim)
		}
	})

	t.Run("PearsonCorrelationSimilarity", func(t *testing.T) {
		// Test with longer vectors for meaningful correlation
		a := []float64{1, 2, 3, 4, 5}
		b := []float64{2, 4, 6, 8, 10} // Perfect positive correlation

		sim := PearsonCorrelationSimilarity(a, b)
		if math.Abs(float64(sim-1)) > 0.001 {
			t.Errorf("Expected ~1 for perfect correlation, got %f", sim)
		}

		// Test negative correlation
		c := []float64{5, 4, 3, 2, 1}
		sim = PearsonCorrelationSimilarity(a, c)
		if math.Abs(float64(sim+1)) > 0.001 {
			t.Errorf("Expected ~-1 for negative correlation, got %f", sim)
		}
	})
}
