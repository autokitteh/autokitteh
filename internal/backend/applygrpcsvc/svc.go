package applygrpcsvc

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/proto"
	applyv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/apply/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/apply/v1/applyv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	client sdkservices.Services
	applyv1connect.UnimplementedApplyServiceHandler
}

var _ applyv1connect.ApplyServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, client sdkservices.Services) {
	srv := server{client: client}

	path, namer := applyv1connect.NewApplyServiceHandler(&srv)
	muxes.Auth.Handle(path, namer)
}

func (s *server) Apply(ctx context.Context, req *connect.Request[applyv1.ApplyRequest]) (*connect.Response[applyv1.ApplyResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	oid, err := sdktypes.ParseOrgID(msg.OrgId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if !oid.IsValid() {
		oid = authcontext.GetAuthnInferredOrgID(ctx)
	}

	man, err := manifest.Read([]byte(msg.Manifest), msg.Path)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var logs []string

	actions, err := manifest.Plan(ctx, man, s.client, manifest.WithLogger(func(msg string) {
		logs = append(logs, "[plan] "+msg)
	}), manifest.WithProjectName(msg.ProjectName), manifest.WithOrgID(oid))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	effects, err := manifest.Execute(ctx, actions, s.client, func(msg string) {
		logs = append(logs, "[exec] "+msg)
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(&applyv1.ApplyResponse{
		Logs:       logs,
		ProjectIds: kittehs.TransformToStrings(effects.ProjectIDs()),
		Effects: kittehs.Transform(effects, func(e *manifest.Effect) *applyv1.Effect {
			return &applyv1.Effect{
				SubjectId: e.SubjectID.String(),
				Type:      string(e.Type),
				Text:      e.Text,
			}
		}),
	}), nil
}
