package gemini

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		resp := embedContentResponse{}
		resp.Embedding.Values = []float64{0.1, 0.2, 0.3, 0.4}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
}

func TestEmbedText(t *testing.T) {
	srv := fakeServer(t)
	defer srv.Close()

	p, err := New("test-key", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	emb, err := p.EmbedText("hello world")
	if err != nil {
		t.Fatal(err)
	}
	if len(emb) != 4 {
		t.Fatalf("expected 4 dims, got %d", len(emb))
	}
}

func TestEmbedBatch(t *testing.T) {
	srv := fakeServer(t)
	defer srv.Close()

	p, err := New("test-key", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}

	results, err := p.EmbedBatch([]string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestMissingAPIKey(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}
