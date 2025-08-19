package mcp

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	projectsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	orgNameArg = mcp.WithString(
		"org_name",
		mcp.Title("Organization name"),
		mcp.Description("Organization name for the project, must be a valid symbol. If not provided, user's default organization will be used."),
		mcp.Pattern(sdktypes.SymbolREPattern),
	)

	projectNameArg = mcp.WithString(
		"project_name",
		mcp.Title("Project name"),
		mcp.Description("Project name, must be a valid symbol."),
		mcp.Pattern(sdktypes.SymbolREPattern),
	)

	projectIDArg = mcp.WithString(
		"project_id",
		mcp.Title("Project id"),
		mcp.Description("Project ID. No need to specify org_name or project_name if using project_id."),
	)
)

var tools = []server.ServerTool{
	{
		Tool: mcp.NewTool(
			"create_project",
			orgNameArg,
			projectNameArg,
			mcp.WithDescription("Create an autokitteh project"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, oid, err := resolveProjectArgs(ctx, request)
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			p := sdktypes.NewProject().WithName(name).WithOrgID(oid)

			pid, err := common.Client().Projects().Create(ctx, p)
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("Created project with id %q", pid)
		},
	},
	{
		Tool: mcp.NewTool(
			"get_project",
			orgNameArg,
			projectNameArg,
			projectIDArg,
			mcp.WithDescription("Get an autokitteh project's details"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			if !p.IsValid() {
				return toolResultErrorf("project not found")
			}

			j, err := p.MarshalJSON()
			if err != nil {
				return toolResultErrorf("marshal json: %v", err)
			}

			return mcp.NewToolResultResource(
				"project info",
				mcp.TextResourceContents{
					URI:      "projects://" + p.ID().String(),
					MIMEType: "application/json",
					Text:     string(j),
				},
			), nil
		},
	},
	{
		Tool: mcp.NewTool(
			"set_project_resources",
			orgNameArg,
			projectNameArg,
			projectIDArg,
			mcp.WithDescription("Set project resources, which contains code and additional data. This overrides all previous resources in the project"),
			mcp.WithArray(
				"resources",
				mcp.Title("Resources"),
				mcp.Description("Resources to set in the project, which are usually code files and additional data"),
				mcp.Required(),
				mcp.Items(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type": "string",
						},
						"data": map[string]any{
							"type": "string",
						},
					},
				}),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			args, err := getMandatoryParam[[]any](request.Params.Arguments, "resources")
			if err != nil {
				return toolResultErrorf("resources: %v", err)
			}

			type rsc struct {
				Path string `json:"path"`
				Data string `json:"data"`
			}

			var rscs []rsc
			if err := mapstructure.Decode(args, &rscs); err != nil {
				return toolResultErrorf("decode resources: %v", err)
			}

			if err := common.Client().Projects().SetResources(ctx, p.ID(), kittehs.ListToMap(
				rscs, func(v rsc) (string, []byte) { return v.Path, []byte(v.Data) },
			)); err != nil {
				return toolResultErrorf("%v", err)
			}

			return mcp.NewToolResultText("Success!"), nil
		},
	},
	{
		Tool: mcp.NewTool(
			"get_project_resources",
			orgNameArg,
			projectNameArg,
			projectIDArg,
			mcp.WithDescription("Get project resources, which contains code and additional data."),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			rscs, err := common.Client().Projects().DownloadResources(ctx, p.ID())
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			j, err := json.Marshal(rscs)
			if err != nil {
				return toolResultErrorf("marshal json: %v", err)
			}

			return mcp.NewToolResultResource(
				"project resources",
				mcp.TextResourceContents{
					URI:      fmt.Sprintf("projects://%s/resources", p.ID()),
					MIMEType: "application/json",
					Text:     string(j),
				},
			), nil
		},
	},
	{
		Tool: mcp.NewTool(
			"build_project",
			orgNameArg,
			projectNameArg,
			projectIDArg,
			mcp.WithDescription("Build project, which creates a build"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			bid, err := common.Client().Projects().Build(ctx, p.ID(), false)
			if err != nil {
				return toolResultErrorf("%v", err)
			}
			return toolResultTextf("Built project, new build created with id %q", bid)
		},
	},
	{
		Tool: mcp.NewTool(
			"export_project",
			orgNameArg,
			projectNameArg,
			projectIDArg,
			mcp.WithString(
				"include_vars_contents",
				mcp.Title("Include variables contents"),
				mcp.Description("If true, export include variable contents"),
			),
			mcp.WithDescription("Get all project data, including resources and manifest file describing the current state of the project"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			includeVarsContents, err := getOptionalParam[bool](request.Params.Arguments, "include_vars_contents")
			if err != nil {
				return toolResultErrorf("include_vars_contents: %v", err)
			}

			data, err := common.Client().Projects().Export(ctx, p.ID(), includeVarsContents)
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			rscs := make(map[string][]byte, len(r.File))

			for _, f := range r.File {
				fr, err := f.Open()
				if err != nil {
					return toolResultErrorf("%q: %v", f.Name, err)
				}

				defer fr.Close()

				b, err := io.ReadAll(fr)
				if err != nil {
					return toolResultErrorf("%q: %v", f.Name, err)
				}

				rscs[f.Name] = b
			}

			j, err := json.Marshal(rscs)
			if err != nil {
				return toolResultErrorf("marshal json: %v", err)
			}

			return mcp.NewToolResultResource(
				"project resources",
				mcp.TextResourceContents{
					URI:      fmt.Sprintf("project://%s/data", p.ID()),
					MIMEType: "application/json",
					Text:     string(j),
				},
			), nil
		},
	},
	{
		Tool: mcp.NewTool(
			"lint_existing_project",
			orgNameArg,
			projectNameArg,
			projectIDArg,
			mcp.WithDescription("Lint existing project, making sure it is setup correctly"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			vs, err := common.Client().Projects().Lint(ctx, p.ID(), nil, "")
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			if len(vs) == 0 {
				return toolResultTextf("Project is valid")
			}

			return toolResultTextf("%s", strings.Join(kittehs.Transform(vs, func(v *sdktypes.CheckViolation) string {
				loc, _ := sdktypes.CodeLocationFromProto(v.Location)

				return fmt.Sprintf("%s %s (%s) %s", loc.CanonicalString(), projectsv1.CheckViolation_Level_name[int32(v.Level)], v.RuleId, v.Message)
			}), "\n"))
		},
	},
	{
		Tool: mcp.NewTool(
			"apply_manifest",
			mcp.WithDescription("Apply a project manifest, which creates or updates a project"),
			mcp.WithString(
				"manifest",
				mcp.Title("Manifest"),
				mcp.Description("Manifest to apply, in YAML format"),
				mcp.Required(),
			),
			projectNameArg,
			orgNameArg,
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, oid, err := resolveProjectArgs(ctx, request)
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			m, err := manifest.Read([]byte(request.Params.Arguments["manifest"].(string)))
			if err != nil {
				return nil, err
			}

			var log []string

			logFunc := func(msg string) { log = append(log, msg) }

			acts, err := manifest.Plan(
				ctx, m, common.Client(),
				manifest.WithLogger(logFunc),
				manifest.WithProjectName(name.String()),
				manifest.WithOrgID(oid),
			)
			if err != nil {
				return toolResultErrorf("plan: %v", err)
			}

			if _, err := manifest.Execute(ctx, acts, common.Client(), logFunc); err != nil {
				return toolResultErrorf("execute: %v", err)
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "Manifest applied successfully",
					},
					mcp.TextContent{
						Type: "text",
						Text: "Log: " + strings.Join(log, "\n"),
					},
				},
			}, nil
		},
	},
	{
		Tool: mcp.NewTool(
			"create_project_deployment",
			mcp.WithDescription("Create an inactive deployment for existing project"),
			projectNameArg,
			projectIDArg,
			orgNameArg,
			mcp.WithString(
				"build_id",
				mcp.Title("Build ID"),
				mcp.Description("Build ID to associate with deployment, must be built from the project specified"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			bidArg, err := getMandatoryParam[string](request.Params.Arguments, "build_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			bid, err := sdktypes.ParseBuildID(bidArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			d := sdktypes.NewDeployment(sdktypes.InvalidDeploymentID, p.ID(), bid)

			did, err := common.Client().Deployments().Create(ctx, d)
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("Created deployment with id %q", did)
		},
	},
	{
		Tool: mcp.NewTool(
			"activate_deployment",
			mcp.WithDescription("Activate an existing deployment"),
			mcp.WithString(
				"deployment_id",
				mcp.Title("Deployment ID"),
				mcp.Description("Deployment ID to activate"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			didArg, err := getMandatoryParam[string](request.Params.Arguments, "deployment_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			did, err := sdktypes.ParseDeploymentID(didArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			if err := common.Client().Deployments().Activate(ctx, did); err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("Activated deployment with id %q", did)
		},
	},
	{
		Tool: mcp.NewTool(
			"get_deployment",
			mcp.WithDescription("Describe an existing deployment"),
			mcp.WithString(
				"deployment_id",
				mcp.Title("Deployment ID"),
				mcp.Description("Deployment ID to activate"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			didArg, err := getMandatoryParam[string](request.Params.Arguments, "deployment_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			did, err := sdktypes.ParseDeploymentID(didArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			d, err := common.Client().Deployments().Get(ctx, did)
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			j, err := d.MarshalJSON()
			if err != nil {
				return toolResultErrorf("marshal json: %v", err)
			}

			return mcp.NewToolResultResource(
				"deployment info",
				mcp.TextResourceContents{
					URI:      fmt.Sprintf("deployments://%s", d.ID()),
					MIMEType: "application/json",
					Text:     string(j),
				},
			), nil
		},
	},
	{
		Tool: mcp.NewTool(
			"deactivate_deployment",
			mcp.WithDescription("Deactivate an existing deployment"),
			mcp.WithString(
				"deployment_id",
				mcp.Title("Deployment ID"),
				mcp.Description("Deployment ID to deactivate"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			didArg, err := getMandatoryParam[string](request.Params.Arguments, "deployment_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			did, err := sdktypes.ParseDeploymentID(didArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			if err := common.Client().Deployments().Deactivate(ctx, did); err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("Deactivated deployment with id %q", did)
		},
	},
	{
		Tool: mcp.NewTool(
			"list_projects",
			orgNameArg,
			mcp.WithDescription("List existing projects"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			oid, err := resolveOrgArg(ctx, request)
			if err != nil {
				return toolResultErrorf("org: %v", err)
			}

			prjs, err := common.Client().Projects().List(ctx, oid)
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("Projects: %s", strings.Join(kittehs.Transform(prjs, func(p sdktypes.Project) string {
				return string(kittehs.Must1(p.MarshalJSON()))
			}), "\n"))
		},
	},
	{
		Tool: mcp.NewTool(
			"list_project_deployments",
			projectNameArg,
			projectIDArg,
			orgNameArg,
			mcp.WithDescription("List existing deployments for project"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			deps, err := common.Client().Deployments().List(ctx, sdkservices.ListDeploymentsFilter{
				OrgID:     p.OrgID(),
				ProjectID: p.ID(),
			})
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("Deployments: %s", strings.Join(kittehs.Transform(deps, func(d sdktypes.Deployment) string {
				return string(kittehs.Must1(d.MarshalJSON()))
			}), "\n"))
		},
	},
	{
		Tool: mcp.NewTool(
			"list_project_connections",
			projectNameArg,
			projectIDArg,
			orgNameArg,
			mcp.WithDescription("List connections for an existing project"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			cs, err := common.Client().Connections().List(ctx, sdkservices.ListConnectionsFilter{
				OrgID:     p.OrgID(),
				ProjectID: p.ID(),
			})
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("Triggers: %s", strings.Join(kittehs.Transform(cs, func(c sdktypes.Connection) string {
				return string(kittehs.Must1(c.MarshalJSON()))
			}), "\n"))
		},
	},
	{
		Tool: mcp.NewTool(
			"list_project_triggers",
			projectNameArg,
			projectIDArg,
			orgNameArg,
			mcp.WithDescription("List triggers for an existing project"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			ts, err := common.Client().Triggers().List(ctx, sdkservices.ListTriggersFilter{
				OrgID:     p.OrgID(),
				ProjectID: p.ID(),
			})
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("Triggers: %s", strings.Join(kittehs.Transform(ts, func(t sdktypes.Trigger) string {
				return string(kittehs.Must1(t.MarshalJSON()))
			}), "\n"))
		},
	},
	{
		Tool: mcp.NewTool(
			"get_trigger",
			mcp.WithDescription("Get a specific trigger for an existing project"),
			mcp.WithString(
				"trigger_id",
				mcp.Title("Trigger ID"),
				mcp.Description("Trigger ID to get"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tidArg, err := getMandatoryParam[string](request.Params.Arguments, "trigger_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			tid, err := sdktypes.ParseTriggerID(tidArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			t, err := common.Client().Triggers().Get(ctx, tid)
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			j, err := t.MarshalJSON()
			if err != nil {
				return toolResultErrorf("marshal json: %v", err)
			}

			return mcp.NewToolResultResource(
				"trigger info",
				mcp.TextResourceContents{
					URI:      fmt.Sprintf("triggers://%s", t.ID()),
					MIMEType: "application/json",
					Text:     string(j),
				},
			), nil
		},
	},
	{
		Tool: mcp.NewTool(
			"get_connection",
			mcp.WithDescription("Get a specific connection for an existing project"),
			mcp.WithString(
				"connection_id",
				mcp.Title("Connection ID"),
				mcp.Description("Connection ID to get"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			cidArg, err := getMandatoryParam[string](request.Params.Arguments, "connection_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			cid, err := sdktypes.ParseConnectionID(cidArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			t, err := common.Client().Connections().Get(ctx, cid)
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			j, err := t.MarshalJSON()
			if err != nil {
				return toolResultErrorf("marshal json: %v", err)
			}

			return mcp.NewToolResultResource(
				"connection info",
				mcp.TextResourceContents{
					URI:      fmt.Sprintf("connections://%s", t.ID()),
					MIMEType: "application/json",
					Text:     string(j),
				},
			), nil
		},
	},
	{
		Tool: mcp.NewTool(
			"list_project_sessions",
			projectNameArg,
			projectIDArg,
			orgNameArg,
			mcp.WithDescription("List sessions for a project"),
			mcp.WithString(
				"deployment_id",
				mcp.Title("Deployment ID"),
				mcp.Description("Only list sessions for this deployment"),
			),
			mcp.WithString(
				"event_id",
				mcp.Title("Event ID"),
				mcp.Description("Only list sessions for this event"),
			),
			mcp.WithString(
				"build_id",
				mcp.Title("Deployment ID"),
				mcp.Description("Only list sessions for this build"),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			var (
				did sdktypes.DeploymentID
				eid sdktypes.EventID
				bid sdktypes.BuildID
			)

			didArg, err := getOptionalParam[string](request.Params.Arguments, "deployment_id")
			if err != nil {
				return toolResultErrorf("deployment_id: %v", err)
			}

			if did, err = sdktypes.ParseDeploymentID(didArg); err != nil {
				return toolResultErrorf("deployment_id: %v", err)
			}

			eidArg, err := getOptionalParam[string](request.Params.Arguments, "event_id")
			if err != nil {
				return toolResultErrorf("event_id: %v", err)
			}

			if eid, err = sdktypes.ParseEventID(eidArg); err != nil {
				return toolResultErrorf("event_id: %v", err)
			}

			bidArg, err := getOptionalParam[string](request.Params.Arguments, "build_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			if bid, err = sdktypes.ParseBuildID(bidArg); err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			r, err := common.Client().Sessions().List(ctx, sdkservices.ListSessionsFilter{
				OrgID:        p.OrgID(),
				ProjectID:    p.ID(),
				DeploymentID: did,
				EventID:      eid,
				BuildID:      bid,
			})
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("Sessions: %s", strings.Join(kittehs.Transform(r.Sessions, func(t sdktypes.Session) string {
				return string(kittehs.Must1(t.MarshalJSON()))
			}), "\n"))
		},
	},
	{
		Tool: mcp.NewTool(
			"get_session",
			mcp.WithDescription("Get a session"),
			mcp.WithString(
				"session_id",
				mcp.Title("Session ID"),
				mcp.Description("Session ID to get"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sidArg, err := getMandatoryParam[string](request.Params.Arguments, "session_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			sid, err := sdktypes.ParseSessionID(sidArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			t, err := common.Client().Sessions().Get(ctx, sid)
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			j, err := t.MarshalJSON()
			if err != nil {
				return toolResultErrorf("marshal json: %v", err)
			}

			return mcp.NewToolResultResource(
				"session info",
				mcp.TextResourceContents{
					URI:      fmt.Sprintf("sessions://%s", t.ID()),
					MIMEType: "application/json",
					Text:     string(j),
				},
			), nil
		},
	},
	{
		Tool: mcp.NewTool(
			"get_session_logs",
			mcp.WithDescription("Get a session's execution log"),
			mcp.WithString(
				"session_id",
				mcp.Title("Session ID"),
				mcp.Description("Session ID to get"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sidArg, err := getMandatoryParam[string](request.Params.Arguments, "session_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			sid, err := sdktypes.ParseSessionID(sidArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			logs, err := common.Client().Sessions().GetLog(ctx, sdkservices.SessionLogRecordsFilter{
				SessionID: sid,
			})
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			j, err := json.Marshal(logs)
			if err != nil {
				return toolResultErrorf("marshal json: %v", err)
			}

			return mcp.NewToolResultResource(
				"session log",
				mcp.TextResourceContents{
					URI:      fmt.Sprintf("sessions://%s/logs", sid),
					MIMEType: "application/json",
					Text:     string(j),
				},
			), nil
		},
	},
	{
		Tool: mcp.NewTool(
			"get_session_prints",
			mcp.WithDescription("Get a session's prints"),
			mcp.WithString(
				"session_id",
				mcp.Title("Session ID"),
				mcp.Description("Session ID to get"),
				mcp.Required(),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sidArg, err := getMandatoryParam[string](request.Params.Arguments, "session_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			sid, err := sdktypes.ParseSessionID(sidArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			prints, err := common.Client().Sessions().GetPrints(ctx, sid, sdktypes.PaginationRequest{})
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			j, err := json.Marshal(prints)
			if err != nil {
				return toolResultErrorf("marshal json: %v", err)
			}

			return mcp.NewToolResultResource(
				"session log",
				mcp.TextResourceContents{
					URI:      fmt.Sprintf("sessions://%s/prints", sid),
					MIMEType: "application/json",
					Text:     string(j),
				},
			), nil
		},
	},
	{
		Tool: mcp.NewTool(
			"stop_session",
			mcp.WithDescription("Stop a session"),
			mcp.WithString(
				"session_id",
				mcp.Title("Session ID"),
				mcp.Description("Session ID to get"),
				mcp.Required(),
			),
			mcp.WithString(
				"reason",
				mcp.Title("reason for stop"),
				mcp.Description("Reason for stopping the session"),
			),
			mcp.WithString(
				"force",
				mcp.Title("try to force stop the session"),
				mcp.Description("If true, will attempt to force stop the session"),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sidArg, err := getMandatoryParam[string](request.Params.Arguments, "session_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			sid, err := sdktypes.ParseSessionID(sidArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			reasonArg, err := getOptionalParam[string](request.Params.Arguments, "reason")
			if err != nil {
				return toolResultErrorf("reason: %v", err)
			}

			forceArg, err := getOptionalParam[bool](request.Params.Arguments, "force")
			if err != nil {
				return toolResultErrorf("reason: %v", err)
			}

			if err := common.Client().Sessions().Stop(ctx, sid, reasonArg, forceArg, time.Second*5); err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("done")
		},
	},
	{
		Tool: mcp.NewTool(
			"list_project_events",
			projectNameArg,
			projectIDArg,
			orgNameArg,
			mcp.WithDescription("List events for a project"),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := resolveExistingProject(ctx, request)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			es, err := common.Client().Events().List(ctx, sdkservices.ListEventsFilter{
				OrgID:     p.OrgID(),
				ProjectID: p.ID(),
			})
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("Events: %s", strings.Join(kittehs.Transform(es, func(t sdktypes.Event) string {
				return string(kittehs.Must1(t.MarshalJSON()))
			}), "\n"))
		},
	},
	{
		Tool: mcp.NewTool(
			"redispatch_event",
			mcp.WithString(
				"event_id",
				mcp.Title("Event ID"),
				mcp.Description("Event ID to redispatch"),
				mcp.Required(),
			),
			mcp.WithDescription("Redispatch event"),
			mcp.WithString(
				"deployment_id",
				mcp.Title("Deployment ID"),
				mcp.Description("Redispatch to this deployment"),
			),
		),
		Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			eidArg, err := getMandatoryParam[string](request.Params.Arguments, "event_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			eid, err := sdktypes.ParseEventID(eidArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			didArg, err := getOptionalParam[string](request.Params.Arguments, "deployment_id")
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			did, err := sdktypes.ParseDeploymentID(didArg)
			if err != nil {
				return toolResultErrorf("build_id: %v", err)
			}

			eid, err = common.Client().Dispatcher().Redispatch(ctx, eid, &sdkservices.DispatchOptions{
				DeploymentID: did,
			})
			if err != nil {
				return toolResultErrorf("%v", err)
			}

			return toolResultTextf("Event redispatched as event id %q", eid)
		},
	},
}

func resolveOrgArg(ctx context.Context, request mcp.CallToolRequest) (oid sdktypes.OrgID, err error) {
	args := request.Params.Arguments

	orgArg, err := getOptionalParam[string](args, "org_name")
	if err != nil {
		err = fmt.Errorf("org: %v", err)
		return
	}

	if oid, err = resolveOrg(orgArg); err != nil {
		err = fmt.Errorf("org: %v", err)
	}

	return
}

func resolveProjectArgs(ctx context.Context, request mcp.CallToolRequest) (name sdktypes.Symbol, oid sdktypes.OrgID, err error) {
	args := request.Params.Arguments

	var nameArg string
	if nameArg, err = getOptionalParam[string](args, "project_name"); err != nil {
		err = fmt.Errorf("name: %v", err)
		return
	}

	if name, err = sdktypes.ParseSymbol(nameArg); err != nil {
		err = fmt.Errorf("project name: %v", err)
		return
	}

	if oid, err = resolveOrgArg(ctx, request); err != nil {
		err = fmt.Errorf("org: %v", err)
		return
	}

	return
}

func resolveExistingProject(ctx context.Context, request mcp.CallToolRequest) (sdktypes.Project, error) {
	args := request.Params.Arguments

	p := sdktypes.NewProject()

	id, err := getOptionalParam[string](args, "project_id")
	if err != nil {
		return sdktypes.InvalidProject, fmt.Errorf("id: %v", err)
	}

	sdkID, err := sdktypes.ParseProjectID(id)
	if err != nil {
		return sdktypes.InvalidProject, fmt.Errorf("project id: %v", err)
	}

	p = p.WithID(sdkID)

	name, err := getOptionalParam[string](args, "project_name")
	if err != nil {
		return sdktypes.InvalidProject, fmt.Errorf("name: %v", err)
	}

	sym, err := sdktypes.ParseSymbol(name)
	if err != nil {
		return sdktypes.InvalidProject, fmt.Errorf("project name: %v", err)
	}

	p = p.WithName(sym)

	org, err := getOptionalParam[string](args, "org_name")
	if err != nil {
		return sdktypes.InvalidProject, fmt.Errorf("org: %v", err)
	}

	oid, err := resolveOrg(org)
	if err != nil {
		return sdktypes.InvalidProject, fmt.Errorf("org: %v", err)
	}

	p = p.WithOrgID(oid)

	switch {
	case p.ID().IsValid():
		if p, err = common.Client().Projects().GetByID(ctx, p.ID()); err != nil {
			return sdktypes.InvalidProject, err
		}
	case !p.Name().IsValid():
		return sdktypes.InvalidProject, errors.New("no id or name provided")
	default:
		if p, err = common.Client().Projects().GetByName(ctx, p.OrgID(), p.Name()); err != nil {
			return sdktypes.InvalidProject, err
		}
	}

	if !p.IsValid() {
		return sdktypes.InvalidProject, errors.New("project not found")
	}

	return p, nil
}

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

func getOptionalParam[T any](args map[string]any, key string) (r T, err error) {
	value, found := args[key]
	if !found {
		return
	}

	var ok bool
	if r, ok = value.(T); !ok {
		err = fmt.Errorf("expected %s to be of type %T", key, r)
		return
	}

	return
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
