package policygrpcsvc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/backend/policy"
	"go.autokitteh.dev/autokitteh/proto"
	policyv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/policy/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/policy/v1/policyv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	db     db.DB
	decide policy.DecideFunc
}

var _ policyv1connect.PolicyServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, decide policy.DecideFunc, db db.DB) {
	srv := server{db: db, decide: decide}

	path, handler := policyv1connect.NewPolicyServiceHandler(&srv)
	muxes.Auth.Handle(path, handler)
}

func (s *server) Decide(ctx context.Context, req *connect.Request[policyv1.DecideRequest]) (*connect.Response[policyv1.DecideResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	uid, err := sdktypes.Strict(sdktypes.ParseUserID(msg.UserId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(fmt.Errorf("user id: %w", err))
	}

	user, err := authz.Hydrate(ctx, s.db, uid, sdktypes.InvalidUser)
	if err != nil {
		return nil, sdkerrors.AsConnectError(fmt.Errorf("hydrate user: %w", err))
	}

	sid, err := sdktypes.ParseAnyID(msg.SubjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(fmt.Errorf("subject id: %w", err))
	}

	subj, err := authz.Hydrate(ctx, s.db, sid, sdktypes.InvalidUser)
	if err != nil {
		return nil, sdkerrors.AsConnectError(fmt.Errorf("hydrate subject: %w", err))
	}

	input := map[string]any{
		"user":    user,
		"subject": subj,
	}

	r, err := s.decide(ctx, msg.Path, input)
	if err != nil {
		return nil, sdkerrors.AsConnectError(fmt.Errorf("decide: %w", err))
	}

	pbv, err := structpb.NewValue(r)
	if err != nil {
		return nil, sdkerrors.AsConnectError(fmt.Errorf("unhandled result type: %w", err))
	}

	return connect.NewResponse(&policyv1.DecideResponse{Result: pbv}), nil
}
