package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req embedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := embedResponse{Embeddings: [][]float64{{0.1, 0.2, 0.3}}}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
}

func TestEmbedText(t *testing.T) {
	srv := fakeServer(t)
	defer srv.Close()
	p := New(WithBaseURL(srv.URL))
	emb, err := p.EmbedText("hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(emb) != 3 {
		t.Fatalf("expected 3 dims, got %d", len(emb))
	}
}

func TestEmbedBatch(t *testing.T) {
	srv := fakeServer(t)
	defer srv.Close()
	p := New(WithBaseURL(srv.URL))
	results, err := p.EmbedBatch([]string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2, got %d", len(results))
	}
}

func TestDefaults(t *testing.T) {
	p := New()
	if p.GetMaxTokens() != defaultTokens {
		t.Fatalf("expected %d, got %d", defaultTokens, p.GetMaxTokens())
	}
}
