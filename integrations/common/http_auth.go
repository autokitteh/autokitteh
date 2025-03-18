package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func HTTPGetWithAuth(ctx context.Context, u, username, password, contentType string, body []byte) ([]byte, error) {
	return httpRequestWithBasicAuth(ctx, http.MethodGet, u, username, password, contentType, body)
}

// httpRequestWithBasicAuth sends an HTTP request with basic authentication and returns the response's body.
// This function accepts only JSON responses, even though it doesn't parse them.
func httpRequestWithBasicAuth(ctx context.Context, method, u, username, password, contentType string, body []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, HTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to construct HTTP request: %w", err)
	}

	req.Header.Set(HeaderAccept, ContentTypeJSON)

	req.SetBasicAuth(username, password)

	if len(body) > 0 {
		req.Header.Set(HeaderContentType, contentType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(http.MaxBytesReader(nil, resp.Body, HTTPMaxSize))
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
