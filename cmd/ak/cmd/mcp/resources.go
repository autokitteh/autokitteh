package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type rscTemplate struct {
	T mcp.ResourceTemplate
	H server.ResourceTemplateHandlerFunc
}

func addResources(srv *server.MCPServer) {
	for _, t := range templates {
		srv.AddResourceTemplate(t.T, t.H)
	}
}

var templates = []rscTemplate{
	{
		T: mcp.NewResourceTemplate(
			"projects://{id}",
			"Project details",
			mcp.WithTemplateDescription("Returns project information"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		H: func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			uri := request.Params.URI

			id := uri[len("projects://"):]

			ctx, cancel := common.WithLimitedContext(ctx)
			defer cancel()

			sdkID, err := sdktypes.ParseProjectID(id)
			if err != nil {
				return nil, fmt.Errorf("id: %w", err)
			}

			p, err := common.Client().Projects().GetByID(ctx, sdkID)
			if err != nil {
				return nil, err
			}

			j, err := p.MarshalJSON()
			if err != nil {
				return nil, fmt.Errorf("marshal json: %w", err)
			}

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      request.Params.URI,
					MIMEType: "application/json",
					Text:     string(j),
				},
			}, nil
		},
	},
}
