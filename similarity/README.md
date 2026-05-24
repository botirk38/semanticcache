# similarity

Similarity functions for comparing embedding vectors. Each function has the signature `func(a, b []float64) float64`.

## Functions

| Function | Range | Description |
|----------|-------|-------------|
| `CosineSimilarity` | [-1, 1] | Angle between vectors. Default for the cache. |
| `EuclideanSimilarity` | (0, 1] | `1 / (1 + distance)`. Higher is more similar. |
| `DotProductSimilarity` | unbounded | Raw dot product. Depends on vector magnitudes. |
| `ManhattanSimilarity` | (0, 1] | `1 / (1 + L1 distance)`. |
| `PearsonCorrelationSimilarity` | [-1, 1] | Pearson correlation coefficient. |

All functions return `0` for mismatched lengths or empty vectors.

## Custom functions

Pass any `func(a, b []float64) float64` to `options.WithSimilarityComparator`.
