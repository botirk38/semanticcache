# Contributing to SemanticCache

## Development Setup

```bash
# Clone the repository
git clone https://github.com/botirk38/semanticcache.git
cd semanticcache

# Download dependencies
go mod download

# Verify everything builds
go build ./...

# Run the test suite
go test ./...
```

## Requirements

- Go 1.25+
- Redis with RedisJSON and RediSearch modules (only for Redis backend development)

## Development Workflow

1. Fork the repository and create a feature branch from `master`.
2. Write your code following the conventions below.
3. Add or update tests for any changed behavior.
4. Ensure all checks pass before submitting a PR:

```bash
go build ./...
go test -race -count=1 ./...
go vet ./...
gofmt -l .
```

## Code Conventions

- Follow standard Go idioms and [Effective Go](https://go.dev/doc/effective_go).
- Use the **functional options pattern** (`options.With*`) for configuration.
- Maintain **generic type support** throughout the API.
- All exported functions and types must have doc comments.
- Use `context.Context` for all operations that may block or be cancelled.
- Keep interfaces small and focused (Interface Segregation Principle).

## Package Structure

| Package | Purpose |
|---------|---------|
| `semanticcache` (root) | Main `SemanticCache[K,V]` API |
| `options/` | Functional options (`With*` functions) |
| `types/` | Shared interfaces and types |
| `similarity/` | Similarity algorithms |
| `backends/inmemory/` | In-memory backends (LRU, LFU, FIFO) |
| `backends/remote/` | Remote backends (Redis) |
| `providers/openai/` | OpenAI embedding provider |
| `chunker/` | Text chunking strategies |
| `tokenizer/` | Token counting utilities |

## Adding a New Similarity Algorithm

1. Create a new file in `similarity/` (e.g., `similarity/jaccard.go`).
2. Implement: `func JaccardSimilarity(a, b []float64) float64`.
3. Add tests in `similarity/similarity_test.go`.

## Adding a New Backend

1. Implement the `types.Backend[K, V]` interface.
2. Add a constructor in `backends/`.
3. Add an `options.With*Backend` function in `options/options.go`.
4. Add integration tests.

## Adding a New Embedding Provider

1. Implement the `types.EmbeddingProvider` interface.
2. Optionally implement `types.BatchEmbeddingProvider` for batch support.
3. Create a package under `providers/` (e.g., `providers/cohere/`).
4. Add an `options.With*Provider` function in `options/options.go`.
5. Add tests with mock HTTP servers or unit tests.

## Running Tests

```bash
# All tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific package
go test -v ./similarity/

# Run benchmarks
go test -bench=. -benchmem ./...
```

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` new feature
- `fix:` bug fix
- `refactor:` code change that neither fixes a bug nor adds a feature
- `test:` adding or updating tests
- `docs:` documentation only changes
- `chore:` maintenance tasks

## Pull Requests

- Keep PRs focused on a single concern.
- Include tests for new functionality.
- Update documentation if the public API changes.
- All CI checks must pass before merge.
