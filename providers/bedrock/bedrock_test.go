package bedrock

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// roundTripFunc lets us redirect requests to a test server.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func fakeServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req invokeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := invokeResponse{Embedding: []float64{0.1, 0.2, 0.3, 0.4}}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
}

func newTestProvider(t *testing.T, srvURL string) *Provider {
	t.Helper()
	p, err := New("AKID", "SECRET")
	if err != nil {
		t.Fatal(err)
	}
	// Replace client with one that rewrites all requests to the test server
	p.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			// Read original body
			var body []byte
			if req.Body != nil {
				body, _ = io.ReadAll(req.Body)
			}
			// Create new request pointing to test server
			newReq, err := http.NewRequest(req.Method, srvURL+req.URL.Path, bytes.NewReader(body))
			if err != nil {
				return nil, err
			}
			newReq.Header = req.Header
			return http.DefaultTransport.RoundTrip(newReq)
		}),
	}
	return p
}

func TestEmbedText(t *testing.T) {
	srv := fakeServer(t)
	defer srv.Close()

	p := newTestProvider(t, srv.URL)
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

	p := newTestProvider(t, srv.URL)
	results, err := p.EmbedBatch([]string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2, got %d", len(results))
	}
}

func TestMissingCredentials(t *testing.T) {
	_, err := New("", "")
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
}

func TestGetMaxTokens(t *testing.T) {
	p, _ := New("AKID", "SECRET")
	if p.GetMaxTokens() != defaultTokens {
		t.Fatalf("expected %d, got %d", defaultTokens, p.GetMaxTokens())
	}
}

func TestSignRequest(t *testing.T) {
	p, _ := New("AKID", "SECRET")
	body := []byte(`{"inputText":"test"}`)
	req, _ := http.NewRequest(http.MethodPost, "https://bedrock-runtime.us-east-1.amazonaws.com/model/test/invoke", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	p.signRequest(req, body)

	auth := req.Header.Get("Authorization")
	if auth == "" {
		t.Fatal("expected Authorization header")
	}
	if !bytes.Contains([]byte(auth), []byte("AWS4-HMAC-SHA256")) {
		t.Fatal("expected SigV4 auth header")
	}
}
