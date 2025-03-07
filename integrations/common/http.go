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
)

const (
	HeaderAccept        = "Accept"
	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"

	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeJSON = "application/json"
)

func HTTPPostForm(ctx context.Context, u, auth string, payload url.Values) ([]byte, error) {
	body := []byte(payload.Encode())
	return HTTPPost(ctx, u, auth, ContentTypeForm, body)
}

func HTTPPostJSON(ctx context.Context, u, auth string, payload any) ([]byte, error) {
	var body []byte
	if s, ok := payload.(string); ok {
		body = []byte(s)
	} else {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON payload: %w", err)
		}
	}
	return HTTPPost(ctx, u, auth, ContentTypeJSON, body)
}

func HTTPPost(ctx context.Context, u, auth, contentType string, body []byte) ([]byte, error) {
	// Construct the request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to construct HTTP request: %w", err)
	}

	req.Header.Set(HeaderAccept, "application/json")
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

	// Read the response's body, up to 8 MiB.
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<23))
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
