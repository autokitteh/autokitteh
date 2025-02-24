package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	var buf bytes.Buffer
	cmd := exec.Command("ak", "-J", "trigger", "list")
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: trigger list - %s\n", err)
		os.Exit(1)
	}

	var slug string
	dec := json.NewDecoder(&buf)
	for {
		var trigger struct {
			WebHook string `json:"webhook_slug"`
		}
		err := dec.Decode(&trigger)
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: parse trigger - %s\n", err)
			os.Exit(1)
		}

		if trigger.WebHook != "" {
			slug = trigger.WebHook
			break
		}
	}

	if slug == "" {
		fmt.Fprintln(os.Stderr, "error: can't find webhook slug")
		os.Exit(1)
	}

	url := fmt.Sprintf("http://localhost:9980/webhooks/%s", slug)
	fmt.Println(url)

	payload := map[string]string{
		"user":  "joe",
		"event": "login",
	}
	buf.Reset()
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		fmt.Fprintf(os.Stderr, "error: can't encode - %s\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(url, "application/json", &buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: can't call - %s\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusAccepted {
		fmt.Fprintf(os.Stderr, "error: bad status - %s\n", resp.Status)
		os.Exit(1)
	}
}
