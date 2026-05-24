# tokenizer -- Agent Instructions

## What this package does
Token counting implementations for different LLM providers. Used by the chunker.

## Tokenizers
- `OpenAITokenizer` -- local counting via tiktoken (cl100k_base). No API call.
- `AnthropicTokenizer` -- counts via Anthropic API. Requires client.
- `GeminiTokenizer` -- counts via Gemini API. Requires client + model.

## Rules
- OpenAI tokenizer should remain local-only (no network calls).
- Anthropic and Gemini tokenizers require API clients -- they make network calls.
- No tests currently exist. When adding tests, mock the API clients.

## Testing
```
go test ./tokenizer/
```
