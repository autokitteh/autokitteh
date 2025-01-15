package db

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func Populate(ctx context.Context, db DB, objs ...sdktypes.Object) (err error) {
	for _, obj := range objs {
		switch obj := obj.(type) {
		case sdktypes.Session:
			err = db.CreateSession(ctx, obj)
		case sdktypes.Connection:
			err = db.CreateConnection(ctx, obj)
		case sdktypes.Trigger:
			err = db.CreateTrigger(ctx, obj)
		case sdktypes.Event:
			err = db.SaveEvent(ctx, obj)
		case sdktypes.Var:
			err = db.SetVars(ctx, []sdktypes.Var{obj})
		case sdktypes.Deployment:
			err = db.CreateDeployment(ctx, obj)
		case sdktypes.Project:
			err = db.CreateProject(ctx, obj)
		case sdktypes.User:
			_, err = db.CreateUser(ctx, obj)
		case sdktypes.Org:
			_, err = db.CreateOrg(ctx, obj)
		case sdktypes.OrgMember:
			err = db.AddOrgMember(ctx, obj)
		default:
			err = sdkerrors.NewInvalidArgumentError("unsupported object type: %T", obj)
		}

		if err != nil {
			return
		}
	}

	return
}
