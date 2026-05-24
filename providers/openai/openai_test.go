package openai

import (
	"testing"
)

func TestNewOpenAIProvider(t *testing.T) {
	t.Run("EmptyAPIKey", func(t *testing.T) {
		_, err := NewOpenAIProvider(OpenAIConfig{})
		if err == nil {
			t.Error("expected error for empty API key")
		}
	})

	t.Run("DefaultModel", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "test-key")
		p, err := NewOpenAIProvider(OpenAIConfig{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.model != DefaultOpenAIModel {
			t.Errorf("expected default model %s, got %s", DefaultOpenAIModel, p.model)
		}
	})

	t.Run("CustomModel", func(t *testing.T) {
		p, err := NewOpenAIProvider(OpenAIConfig{
			APIKey: "test-key",
			Model:  "text-embedding-ada-002",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.model != "text-embedding-ada-002" {
			t.Errorf("expected custom model, got %s", p.model)
		}
	})
}

func TestOpenAIProvider_Close(t *testing.T) {
	p, err := NewOpenAIProvider(OpenAIConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}
