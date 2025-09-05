// Package similarity provides various similarity algorithms for comparing embedding vectors.
package similarity

// SimilarityFunc represents a function that computes similarity between two embedding vectors.
// It should return a float32 where higher values indicate greater similarity.
type SimilarityFunc func(a, b []float32) float32
