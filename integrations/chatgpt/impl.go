package chatgpt

import (
	"context"
	"errors"

	openai "github.com/sashabaranov/go-openai"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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
		return sdktypes.InvalidValue, err
	}

	if msg != "" && len(req.Messages) > 0 {
		return sdktypes.InvalidValue, errors.New("cannot specify both 'message' and 'messages'")
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

	// Retrieve auth details based on connection ID.
	cid, err := sdkmodule.FunctionConnectionIDFromContext(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	cvars, err := i.vars.Reveal(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client := openai.NewClient(cvars.GetValue(apiKeyVar))
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		jsonResp := ChatCompletionResponse{Error: err.Error()}
		return sdktypes.WrapValue(jsonResp)
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
	return sdktypes.WrapValue(jsonResp)
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
