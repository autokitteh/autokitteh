package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func resolveOrg(org string) (sdktypes.OrgID, error) {
	if org == "" {
		return sdktypes.InvalidOrgID, nil
	}

	ctx, cancel := common.LimitedContext()
	defer cancel()

	r := resolver.Resolver{Client: common.Client()}
	oid, err := r.Org(ctx, org)
	if err != nil {
		return sdktypes.InvalidOrgID, fmt.Errorf("resolve org: %w", err)
	}

	return oid, nil
}

func getOptionalParam[T any](args map[string]any, key string) (*T, error) {
	value, ok := args[key]
	if !ok {
		return nil, nil
	}

	result, ok := value.(T)
	if !ok {
		return nil, fmt.Errorf("expected %s to be of type %T", key, result)
	}

	return &result, nil
}

func getMandatoryParam[T any](args map[string]any, key string) (T, error) {
	var t T

	value, ok := args[key]
	if !ok {
		return t, fmt.Errorf("missing mandatory parameter %s", key)
	}

	result, ok := value.(T)
	if !ok {
		return t, fmt.Errorf("expected %s to be of type %T", key, result)
	}

	return result, nil
}

func toolResultErrorf(msg string, args ...any) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(fmt.Sprintf(msg, args...)), nil
}

func toolResultTextf(msg string, args ...any) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(fmt.Sprintf(msg, args...)), nil
}
