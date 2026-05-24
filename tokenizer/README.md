# tokenizer

Token counting implementations for different LLM providers. Used by the chunker to determine text length in tokens.

## Tokenizers

### OpenAITokenizer

Local token counting using tiktoken (cl100k_base encoding). No API call needed.

```go
t := tokenizer.NewOpenAITokenizer()
count, err := t.CountTokens(ctx, messages)
```

### AnthropicTokenizer

Counts tokens via Anthropic's API. Requires an Anthropic client.

```go
t := tokenizer.NewAnthropicTokenizer(anthropicClient)
count, err := t.CountTokens(ctx, messages)
```

### GeminiTokenizer

Counts tokens via Google's Gemini API. Requires a Gemini client and model name.

```go
t := tokenizer.NewGeminiTokenizer(geminiClient, "gemini-1.5-flash")
count, err := t.CountTokens(ctx, contents)
```
