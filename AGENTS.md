# SemanticCache - AI Agent Instructions

This is a high-performance semantic caching library for Go that uses vector embeddings to find semantically similar content.

## Dev environment tips
- Use `go mod tidy` to ensure dependencies are up to date
- Run `go test ./...` to execute the full test suite across all packages
- Use `go test -cover ./...` for test coverage analysis
- Run `go build ./...` to verify all packages compile correctly
- Use `go run` with the main examples for quick testing

## Project structure
The project follows Go best practices with modular package organization:
- `cache.go` - Main cache implementation with `New()` function
- `options/` - All functional options (`options.With*()` functions)
- `similarity/` - Similarity algorithms (`similarity.*Similarity` functions)  
- `backends/` - Storage backends (in-memory and Redis)
- `providers/` - Embedding providers (OpenAI)
- `types/` - Shared interfaces and types

## Testing instructions
- Run the full test suite: `go test ./...`
- Test specific packages: `go test ./cache_test.go ./cache.go -v`
- Run benchmarks: `go test -bench=. ./...`
- Check test coverage: `go test -cover ./...`
- All tests must pass before merging changes
- Add tests for any new functionality you implement
- Use mock providers for testing without external API dependencies

## Code style guidelines
- Follow standard Go conventions and idioms
- Use the functional options pattern for configuration
- Maintain generic type support throughout the API
- All public functions and types must have Go doc comments
- Use context.Context for all operations that might be cancelled
- Keep interface segregation principle - small, focused interfaces

## Build and development commands
- `go mod download` - Download dependencies
- `go build ./...` - Build all packages
- `go test ./...` - Run all tests
- `go test -race ./...` - Run tests with race detection
- `go vet ./...` - Run static analysis
- `go fmt ./...` - Format code

## Package import structure
When adding new code, use these imports:
```go
import (
    "github.com/botirkhaltaev/semanticcache"           // Main cache
    "github.com/botirkhaltaev/semanticcache/options"   // Configuration options
    "github.com/botirkhaltaev/semanticcache/similarity" // Similarity algorithms
    "github.com/botirkhaltaev/semanticcache/types"     // Shared types
)
```

## Performance considerations
- LRU backend for time-locality patterns
- LFU backend for frequency-based access patterns  
- FIFO backend for simple queue-like usage
- Use batch operations (`SetBatch`, `GetBatch`) for multiple items
- Adjust similarity thresholds based on precision requirements
- Always use context for request timeouts and cancellation

## Common development tasks

### Adding new similarity algorithms
1. Create new file in `similarity/` package
2. Implement function with signature: `func(a, b []float32) float32`
3. Add comprehensive tests in `similarity_test.go`
4. Export function for use in options

### Adding new backends
1. Implement the `Backend` interface from `types/types.go`
2. Create corresponding option function in `options/options.go`
3. Add integration tests with the main cache
4. Update documentation and examples

### Adding new providers
1. Implement the `EmbeddingProvider` interface from `types/types.go`
2. Create option function in `options/options.go`
3. Add tests with mock implementations
4. Consider rate limiting and error handling

## Error handling
- Return descriptive errors for configuration issues
- Handle network timeouts and API rate limits gracefully
- Validate inputs and provide clear error messages
- Use Go's standard error handling patterns

## Security considerations
- Never log or expose API keys or sensitive data
- Validate all inputs to prevent injection attacks
- Use secure defaults for configurations
- Handle authentication failures appropriately

## Dependencies
The project uses minimal external dependencies:
- Standard library for core functionality
- Context package for cancellation and timeouts
- Generic type support (Go 1.18+)
- External APIs (OpenAI) are optional and pluggable