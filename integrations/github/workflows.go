package github

import (
	"context"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://docs.github.com/en/rest/actions/workflows?apiVersion=2022-11-28#create-a-workflow-dispatch-event
func (i integration) triggerWorkflow(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo, workflowName string
		req                       github.CreateWorkflowDispatchEventRequest
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"ref", &req.Ref,
		"workflow_name", &workflowName,
		"inputs?", &req.Inputs,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	_, err = gh.Actions.CreateWorkflowDispatchEventByFileName(ctx, owner, repo, workflowName, req)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.Nothing, nil
}

// https://docs.github.com/en/rest/actions/workflows#list-repository-workflows
func (i integration) listWorkflows(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		owner, repo string
		listOptions github.ListOptions
	)

	if err := sdkmodule.UnpackArgs(args, kwargs,
		"owner", &owner,
		"repo", &repo,
		"per_page?", &listOptions.PerPage,
		"page?", &listOptions.Page,
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	workflows, _, err := gh.Actions.ListWorkflows(ctx, owner, repo, &listOptions)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.WrapValue(workflows)
}
