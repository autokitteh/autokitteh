package common

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	HeaderAccept        = "Accept"
	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"

	ContentTypeForm            = "application/x-www-form-urlencoded"
	ContentTypeJSON            = "application/json"                // Accept
	ContentTypeJSONCharsetUTF8 = "application/json; charset=utf-8" // Content-Type

	HTTPTimeout = 3 * time.Second
	HTTPMaxSize = 1 << 23 // 2^23 bytes = 8 MiB
)

func HTTPGet(ctx context.Context, u, auth string) ([]byte, error) {
	return httpRequest(ctx, http.MethodGet, u, auth, "", nil)
}

func HTTPGetJSON(ctx context.Context, u, auth string, payload any) ([]byte, error) {
	if s, ok := payload.(string); ok {
		return httpGet(ctx, u, auth, ContentTypeJSONCharsetUTF8, []byte(s))
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON payload: %w", err)
	}
	return httpGet(ctx, u, auth, ContentTypeJSONCharsetUTF8, body)
}

func HTTPPostForm(ctx context.Context, u, auth string, payload url.Values) ([]byte, error) {
	body := []byte(payload.Encode())
	return httpPost(ctx, u, auth, ContentTypeForm, body)
}

func HTTPPostJSON(ctx context.Context, u, auth string, payload any) ([]byte, error) {
	if s, ok := payload.(string); ok {
		return httpPost(ctx, u, auth, ContentTypeJSONCharsetUTF8, []byte(s))
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON payload: %w", err)
	}
	return httpPost(ctx, u, auth, ContentTypeJSONCharsetUTF8, body)
}

func httpGet(ctx context.Context, u, auth, contentType string, body []byte) ([]byte, error) {
	return httpRequest(ctx, http.MethodGet, u, auth, contentType, body)
}

func httpPost(ctx context.Context, u, auth, contentType string, body []byte) ([]byte, error) {
	return httpRequest(ctx, http.MethodPost, u, auth, contentType, body)
}

// httpRequest sends an HTTP GET or POST request and returns the response's body.
// This function accepts only JSON responses, even though it doesn't parse them.
func httpRequest(ctx context.Context, method, u, auth, contentType string, body []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, HTTPTimeout)
	defer cancel()

	// Construct the request.
	req, err := http.NewRequestWithContext(ctx, method, u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to construct HTTP request: %w", err)
	}

	req.Header.Set(HeaderAccept, ContentTypeJSON)
	if auth != "" {
		req.Header.Set(HeaderAuthorization, auth)
	}
	if len(body) > 0 {
		req.Header.Set(HeaderContentType, contentType)
	}

	// Send the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response's body.
	payload, err := io.ReadAll(io.LimitReader(resp.Body, HTTPMaxSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response's body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		s := fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		if len(payload) > 0 {
			s = fmt.Sprintf("%s: %s", s, string(payload))
		}
		return nil, errors.New(s)
	}

	return payload, nil
}
