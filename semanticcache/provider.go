package semanticcache

// EmbeddingProvider defines the interface all embedding‐backends must satisfy.
type EmbeddingProvider interface {
	// EmbedText turns a piece of text into its embedding vector.
	EmbedText(text string) ([]float32, error)
	// Close frees any resources held by the provider.
	Close()
}
