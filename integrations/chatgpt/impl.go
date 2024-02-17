package chatgpt

import (
	"context"
	"errors"

	openai "github.com/sashabaranov/go-openai"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

const (
	defaultModel = openai.GPT3Dot5Turbo
	defaultRole  = openai.ChatMessageRoleUser
)

// https://pkg.go.dev/github.com/sashabaranov/go-openai#Client.CreateChatCompletion
// https://platform.openai.com/docs/api-reference/chat/create
func (i integration) createChatCompletion(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		msg string
		req openai.ChatCompletionRequest
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"model?", &req.Model,
		"message?", &msg,
		"messages?", &req.Messages,
	)
	if err != nil {
		return nil, err
	}

	if msg != "" && len(req.Messages) > 0 {
		return nil, errors.New("cannot specify both 'message' and 'messages'")
	}

	if req.Model == "" {
		req.Model = defaultModel
	}
	if msg != "" {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    defaultRole,
			Content: msg,
		})
	}

	// Get auth details from secrets manager.
	token := sdkmodule.FunctionDataFromContext(ctx)
	auth, err := i.secrets.Get(ctx, i.scope, string(token))
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	client := openai.NewClient(auth["apiKey"])
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		jsonResp := ChatCompletionResponse{Error: err.Error()}
		return sdkvalues.Wrap(jsonResp)
	}

	// Parse and return the response.
	jsonResp := ChatCompletionResponse{
		ID:                resp.ID,
		Object:            resp.Object,
		Created:           resp.Created,
		Model:             resp.Model,
		Usage:             resp.Usage,
		SystemFingerprint: resp.SystemFingerprint,
	}
	for _, c := range resp.Choices {
		jsonResp.Choices = append(jsonResp.Choices, ChatCompletionChoice{
			Index:        c.Index,
			Message:      c.Message,
			FinishReason: string(c.FinishReason),
		})
	}
	return sdkvalues.Wrap(jsonResp)
}

// Workaround for a JSON conversion issue in the client library,
// and for passing errors back to the caller.
type ChatCompletionResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []ChatCompletionChoice `json:"choices"`
	Usage             openai.Usage           `json:"usage"`
	SystemFingerprint string                 `json:"system_fingerprint"`

	Error string `json:"error,omitempty"`
}

type ChatCompletionChoice struct {
	Index        int                          `json:"index"`
	Message      openai.ChatCompletionMessage `json:"message"`
	FinishReason string                       `json:"finish_reason"`
}
