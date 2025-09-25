package mcp

import (
	"context"
	_ "embed"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

//go:embed dearai.txt
var prompt string

//go:embed mcp.txt
var mcpDoc string

func addPrompts(srv *server.MCPServer) {
	srv.AddPrompt(
		mcp.Prompt{
			Name:        "autokitteh",
			Description: "How to work with AutoKitteh",
		},
		func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return &mcp.GetPromptResult{
				Description: "How to work with AutoKitteh",
				Messages: []mcp.PromptMessage{
					{
						Role: mcp.RoleAssistant,
						Content: mcp.TextContent{
							Type: "text",
							Text: mcpDoc + prompt,
						},
					},
				},
			}, nil
		},
	)
}
