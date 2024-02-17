package api

type SlackResponse struct {
	OK               bool              `json:"ok"`
	Warning          string            `json:"warning,omitempty"`
	Error            string            `json:"error,omitempty"`
	ResponseMetadata *ResponseMetadata `json:"response_metadata,omitempty"`
}

type ResponseMetadata struct {
	Messages   []string `json:"messages,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	NextCursor string   `json:"next_cursor,omitempty"`
}
