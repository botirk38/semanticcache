# similarity -- Agent Instructions

## What this package does
Provides similarity functions for comparing embedding vectors. Each function has the signature `func(a, b []float64) float64`.

## Available functions
- `CosineSimilarity` -- default, angle-based
- `EuclideanSimilarity` -- inverse distance
- `DotProductSimilarity` -- raw dot product
- `ManhattanSimilarity` -- inverse L1 distance
- `PearsonCorrelationSimilarity` -- correlation coefficient

## Rules
- One function per file.
- All functions must return 0 for mismatched lengths or empty vectors.
- Add tests in `similarity_test.go` for any new function.
- The `SimilarityFunc` type is defined in `similarity.go`.

## Testing
```
go test ./similarity/
```
