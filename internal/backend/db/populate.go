package db

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func Populate(ctx context.Context, db DB, objs ...sdktypes.Object) (err error) {
	for _, obj := range objs {
		switch obj := obj.(type) {
		case sdktypes.Project:
			err = db.CreateProject(ctx, obj)
		case sdktypes.Env:
			err = db.CreateEnv(ctx, obj)
		case sdktypes.Event:
			err = db.SaveEvent(ctx, obj)
		case sdktypes.Var:
			err = db.SetVars(ctx, []sdktypes.Var{obj})
		case sdktypes.Trigger:
			err = db.CreateTrigger(ctx, obj)
		case sdktypes.Connection:
			err = db.CreateConnection(ctx, obj)
		case sdktypes.Deployment:
			err = db.CreateDeployment(ctx, obj)
		case sdktypes.Session:
			err = db.CreateSession(ctx, obj)
		default:
			sdklogger.Panic("unknown object type")
		}

		if err != nil {
			return
		}
	}

	return
}
