// Package bedrock provides an embedding provider backed by AWS Bedrock
// using the Titan Embeddings model.
//
// This provider uses AWS Signature V4 authentication. Credentials are
// read from the standard AWS environment variables or credential chain.
//
// See https://docs.aws.amazon.com/bedrock/latest/userguide/titan-embedding-models.html
package bedrock

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultModel  = "amazon.titan-embed-text-v2:0"
	defaultRegion = "us-east-1"
	defaultTokens = 8192
	service       = "bedrock"
)

// Provider implements EmbeddingProvider using AWS Bedrock Titan.
type Provider struct {
	accessKey string
	secretKey string
	region    string
	model     string
	client    *http.Client
}

// Option configures a Bedrock Provider.
type Option func(*Provider)

// WithModel overrides the default Titan model.
func WithModel(model string) Option { return func(p *Provider) { p.model = model } }

// WithRegion sets the AWS region.
func WithRegion(region string) Option { return func(p *Provider) { p.region = region } }

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(p *Provider) { p.client = c } }

// New creates a Bedrock embedding provider. AWS credentials are read from
// parameters or AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY env vars.
func New(accessKey, secretKey string, opts ...Option) (*Provider, error) {
	if accessKey == "" {
		accessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if secretKey == "" {
		secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("bedrock: AWS credentials required (pass directly or set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY)")
	}
	p := &Provider{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    defaultRegion,
		model:     defaultModel,
		client:    http.DefaultClient,
	}
	for _, o := range opts {
		o(p)
	}
	return p, nil
}

type invokeRequest struct {
	InputText string `json:"inputText"`
}

type invokeResponse struct {
	Embedding []float64 `json:"embedding"`
	Message   string    `json:"message,omitempty"`
}

// EmbedText embeds a single text via the Bedrock InvokeModel API.
func (p *Provider) EmbedText(text string) ([]float64, error) {
	body, err := json.Marshal(invokeRequest{InputText: text})
	if err != nil {
		return nil, fmt.Errorf("bedrock: marshal: %w", err)
	}

	url := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke", p.region, p.model)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("bedrock: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	p.signRequest(req, body)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bedrock: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("bedrock: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bedrock: API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result invokeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("bedrock: decode response: %w", err)
	}

	if len(result.Embedding) == 0 {
		return nil, errors.New("bedrock: no embedding returned")
	}

	return result.Embedding, nil
}

// EmbedBatch embeds multiple texts sequentially (Bedrock doesn't support batch).
func (p *Provider) EmbedBatch(texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, errors.New("bedrock: no texts provided")
	}
	results := make([][]float64, len(texts))
	for i, t := range texts {
		emb, err := p.EmbedText(t)
		if err != nil {
			return nil, fmt.Errorf("bedrock: batch item %d: %w", i, err)
		}
		results[i] = emb
	}
	return results, nil
}

// Close is a no-op.
func (p *Provider) Close() {}

// GetMaxTokens returns the token limit.
func (p *Provider) GetMaxTokens() int { return defaultTokens }

// signRequest adds AWS Signature V4 headers to the request.
func (p *Provider) signRequest(req *http.Request, payload []byte) {
	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")

	req.Header.Set("X-Amz-Date", amzDate)

	host := req.URL.Host
	req.Header.Set("Host", host)

	payloadHash := sha256Hex(payload)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	canonicalHeaders := "content-type:" + req.Header.Get("Content-Type") + "\n" +
		"host:" + host + "\n" +
		"x-amz-content-sha256:" + payloadHash + "\n" +
		"x-amz-date:" + amzDate + "\n"
	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := strings.Join([]string{
		req.Method,
		req.URL.Path,
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	credentialScope := dateStamp + "/" + p.region + "/" + service + "/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + credentialScope + "\n" + sha256Hex([]byte(canonicalRequest))

	signingKey := deriveKey(p.secretKey, dateStamp, p.region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	req.Header.Set("Authorization",
		"AWS4-HMAC-SHA256 Credential="+p.accessKey+"/"+credentialScope+
			", SignedHeaders="+signedHeaders+
			", Signature="+signature)
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return mac.Sum(nil)
}

func deriveKey(secret, dateStamp, region, svc string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(svc))
	return hmacSHA256(kService, []byte("aws4_request"))
}
