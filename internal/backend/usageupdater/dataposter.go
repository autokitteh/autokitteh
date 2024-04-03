package usageupdater

import (
	"bytes"
	"crypto/tls"
	"net/http"
)

type poster struct {
	endpoint string
	client   http.Client
}

func newPoster(endpoint string) poster {
	tr := &http.Transport{
		DisableKeepAlives:   true,
		MaxIdleConnsPerHost: -1,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := http.Client{Transport: tr}

	return poster{
		endpoint: endpoint,
		client:   client,
	}
}

func (p poster) post(data []byte) {
	req, err := http.NewRequest("POST", p.endpoint, bytes.NewReader(data))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()
}
