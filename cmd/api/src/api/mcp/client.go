package mcp

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type client struct {
	httpClient *http.Client
	baseURL    string
	tokenID    string
	tokenKey   string
}

func newClient(baseURL, tokenID, tokenKey string) *client {
	return &client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		tokenID:    tokenID,
		tokenKey:   tokenKey,
	}
}

type bhResponse struct {
	Data json.RawMessage `json:"data"`
}

func (c *client) Get(ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
	reqURL := c.baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	return c.doRequest(req)
}

func (c *client) Post(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.doJSON(ctx, "POST", path, body)
}

func (c *client) Patch(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.doJSON(ctx, "PATCH", path, body)
}

func (c *client) doJSON(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	return c.doRequest(req)
}

func (c *client) doRequest(req *http.Request) (json.RawMessage, error) {
	if err := c.signRequest(req); err != nil {
		return nil, fmt.Errorf("signing request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var envelope bhResponse
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return respBody, nil
	}
	if envelope.Data != nil {
		return envelope.Data, nil
	}
	return respBody, nil
}

// signRequest signs using the BloodHound HMAC-SHA256 scheme.
// NOTE: Uses req.URL.RequestURI() (includes query params) to match server-side validation.
func (c *client) signRequest(req *http.Request) error {
	now := time.Now().UTC()
	datetimeFormatted := now.Format(time.RFC3339)

	// OperationKey = HMAC-SHA256(tokenKey, method + requestURI)
	digester := hmac.New(sha256.New, []byte(c.tokenKey))
	digester.Write([]byte(req.Method + req.URL.RequestURI()))

	// DateKey = HMAC-SHA256(operationKey, datetime[:13])
	digester = hmac.New(sha256.New, digester.Sum(nil))
	digester.Write([]byte(datetimeFormatted[:13]))

	// BodySignature = HMAC-SHA256(dateKey, requestBody)
	digester = hmac.New(sha256.New, digester.Sum(nil))
	if req.Body != nil {
		var buf bytes.Buffer
		tee := io.TeeReader(req.Body, &buf)
		if _, err := io.Copy(digester, tee); err != nil {
			return fmt.Errorf("hmac body: %w", err)
		}
		req.Body = io.NopCloser(&buf)
	}

	req.Header.Set("Authorization", "bhesignature "+c.tokenID)
	req.Header.Set("RequestDate", datetimeFormatted)
	req.Header.Set("Signature", base64.StdEncoding.EncodeToString(digester.Sum(nil)))
	req.Header.Set("Content-Type", "application/json")
	return nil
}
