package cohere

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var req embedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		embs := make([][]float64, len(req.Texts))
		for i := range embs {
			embs[i] = make([]float64, 4)
			for j := range embs[i] {
				embs[i][j] = float64(i+1) * 0.1 * float64(j+1)
			}
		}

		resp := embedResponse{}
		resp.Embeddings.Float = embs
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

	emb, err := p.EmbedText("hello")
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

	results, err := p.EmbedBatch([]string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestMissingAPIKey(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestGetMaxTokens(t *testing.T) {
	p, _ := New("test-key")
	if p.GetMaxTokens() != defaultTokens {
		t.Fatalf("expected %d, got %d", defaultTokens, p.GetMaxTokens())
	}
}
