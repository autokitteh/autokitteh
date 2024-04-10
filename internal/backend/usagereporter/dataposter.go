package usagereporter

import (
	"bytes"
	"fmt"
	"net/http"
)

func post(endpoint string, data []byte) error {
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("constructing request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	return nil
}
