# SemanticCache - Claude Development Instructions

## Project Overview
This is a high-performance semantic caching library for Go that uses vector embeddings to find semantically similar content. The library supports multiple backends (in-memory and Redis) and embedding providers (OpenAI).

## Project Structure
The project is organized into modular packages following Go best practices:

```
semanticcache/
├── similarity/                    # Similarity algorithms package
│   ├── similarity.go             # SimilarityFunc type definition
│   ├── cosine.go                 # Individual algorithm implementations
│   ├── euclidean.go, dotproduct.go, manhattan.go, pearson.go
│   └── similarity_test.go        # Complete test suite
├── options/                       # Functional options package  
│   ├── options.go                # All With* option functions
│   └── options_test.go           # Options test suite
├── backends/                      # Storage backends
│   ├── inmemory/                 # In-memory backends (LRU, LFU, FIFO)
│   └── remote/                   # Remote backends (Redis)
├── providers/                     # Embedding providers
│   └── openai/                   # OpenAI provider implementation
├── types/                        # Shared types and interfaces
├── cache.go                      # Main cache implementation
├── cache_test.go                 # Main cache tests
└── README.md                     # Updated with new structure
```

## Development Guidelines

### Package Usage
- **Main cache**: `semanticcache.New()` - creates new cache instances
- **Options**: `options.With*()` - all functional options for configuration
- **Similarity**: `similarity.*Similarity` - similarity algorithm functions
- **Types**: `types.*` - shared interfaces and types

### Testing Commands
- Run all tests: `go test ./...`
- Run specific package tests: `go test ./cache_test.go ./cache.go -v`
- Test coverage: `go test -cover ./...`

### Code Style
- Follow Go idioms and conventions
- Use functional options pattern for configuration
- Maintain backward compatibility when possible
- Add comprehensive tests for new features
- Document public APIs with Go comments

### Key Design Patterns
1. **Functional Options**: All configuration uses the options pattern
2. **Interface Segregation**: Separate interfaces for backends, providers, similarity functions
3. **Generic Types**: Full generic support for key/value types
4. **Context Awareness**: All operations support context.Context
5. **Modular Architecture**: Clear separation of concerns across packages

### Common Tasks

#### Adding New Similarity Algorithm
1. Create new file in `similarity/` package
2. Implement function with signature `func(a, b []float32) float32`
3. Add tests in `similarity_test.go`
4. Export function for use in options

#### Adding New Backend
1. Implement the `Backend` interface in `types/types.go`
2. Create option function in `options/options.go`
3. Add comprehensive tests
4. Update documentation

#### Adding New Provider
1. Implement the `EmbeddingProvider` interface in `types/types.go`
2. Create option function in `options/options.go`
3. Add integration tests
4. Update documentation

### Testing Strategy
- Unit tests for individual components
- Integration tests for end-to-end functionality
- Benchmark tests for performance-critical paths
- Mock providers for testing without external dependencies

### Import Guidelines
When working with the codebase:
```go
import (
    "github.com/botirkhaltaev/semanticcache"
    "github.com/botirkhaltaev/semanticcache/options"
    "github.com/botirkhaltaev/semanticcache/similarity"
    "github.com/botirkhaltaev/semanticcache/types"
)
```

### Recent Restructuring
The project was recently restructured from a monolithic approach to modular packages:
- Moved similarity functions to `similarity/` package
- Moved functional options to `options/` package
- Updated all imports and references throughout the codebase
- Maintained API compatibility where possible

### Build and Test
Ensure all tests pass before making changes:
```bash
go test ./...
go build ./...
```

The project should build without warnings and all tests should pass.