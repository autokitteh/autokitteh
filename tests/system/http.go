package systest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	httpClientTimeout = 5 * time.Second
)

type httpRequest struct {
	method  string
	url     string
	headers map[string]string
	body    string
}

type httpResponse struct {
	resp *http.Response
	body string
}

func sendRequest(akAddr string, r httpRequest) (*httpResponse, error) {
	u := r.url
	var err error
	if !strings.HasPrefix(u, "http") {
		u, err = url.JoinPath("http://"+akAddr, u)
		if err != nil {
			return nil, fmt.Errorf("failed to construct request URL: %w", err)
		}
	}

	var body io.Reader
	if r.body != "" {
		body = io.NopCloser(strings.NewReader(r.body))
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, r.method, u, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create new HTTP request: %w", err)
	}
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body: %w", err)
	}

	return &httpResponse{resp: resp, body: string(b)}, nil
}
