package mistral

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
		data := make([]struct {
			Embedding []float64 `json:"embedding"`
		}, len(req.Input))
		for i := range data {
			data[i].Embedding = []float64{0.1, 0.2, 0.3}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"data": data}); err != nil {
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
	if len(emb) != 3 {
		t.Fatalf("expected 3 dims, got %d", len(emb))
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
		t.Fatalf("expected 2, got %d", len(results))
	}
}

func TestMissingAPIKey(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Fatal("expected error")
	}
}
