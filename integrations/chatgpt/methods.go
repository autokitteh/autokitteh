package chatgpt

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIAPI struct {
	APIKey string
}

func (api OpenAIAPI) Test(ctx context.Context) error {
	client := openai.NewClient(api.APIKey)
	_, err := client.ListModels(ctx) // This will test the API key validity
	if err != nil {
		return fmt.Errorf("failed to authenticate with OpenAI: %w", err)
	}
	return nil
}
