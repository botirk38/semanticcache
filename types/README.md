# types

Core interfaces for the semantic cache library.

## Interfaces

### Backend[K, V]

The storage interface every cache backend must implement. 9 methods:

- `Set(ctx, key, embedding, value)` -- store a value with its embedding vector
- `Get(ctx, key)` -- retrieve a value by key
- `Delete(ctx, key)` -- remove an entry
- `Contains(ctx, key)` -- check existence
- `Keys(ctx)` -- list all keys
- `GetEmbedding(ctx, key)` -- retrieve the embedding vector for a key
- `Flush(ctx)` -- remove all entries
- `Len(ctx)` -- count entries
- `Close()` -- release resources

### EmbeddingProvider

Turns text into embedding vectors:

- `EmbedText(ctx, text)` -- compute a single embedding
- `Close()` -- release resources

### BatchEmbeddingProvider

Optional extension for providers supporting batch embedding:

- Embeds `EmbeddingProvider`
- `EmbedBatch(ctx, texts)` -- embed multiple texts in one call

## Types

### Entry[V]

Holds an embedding vector alongside its cached value. Used internally by backends.
