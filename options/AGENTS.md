# options -- Agent Instructions

## What this package does
Provides functional options (`Option[K, V]`) for configuring cache instances. Each `With*` function returns a closure that configures a `Config[K, V]`.

## Key types
- `Option[K, V]` -- `func(*Config[K, V]) error`
- `Config[K, V]` -- holds Backend, Provider, Comparator

## Rules
- When adding a new backend or provider, add a corresponding `With*` function here.
- Errors for nil arguments are defined in this package (`ErrNilBackend`, `ErrNilProvider`, `ErrNilComparator`).
- Default similarity is `CosineSimilarity`.
- Validate requires non-nil Backend and Provider.

## Testing
```
go test ./options/
```
Tests use mock implementations defined in `options_test.go`.
