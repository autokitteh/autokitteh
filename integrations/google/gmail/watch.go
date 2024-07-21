package gmail

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/api/gmail/v1"

	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// TODO(ENG-1203): Make this configurable! Env var?
	topic = "projects/autokitteh-gapis-integration/topics/gmail-api-push"
)

// UpdateWatch creates or updates a push notification watch on the user's mailbox.
func UpdateWatch(ctx context.Context, v sdkservices.Vars, cid sdktypes.ConnectionID) error {
	// Not a Gmail user? Nothing to do.
	if !gotGmailScope(ctx, v, cid) {
		return nil
	}

	watch, err := api{vars: v, cid: cid}.watch(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Gmail watch: %w", err)
	}

	l := extrazap.ExtractLoggerFromContext(ctx)
	l.Info("Created Gmail user mailbox watch", zap.Any("watch", watch))
	return nil
}

func gotGmailScope(ctx context.Context, v sdkservices.Vars, cid sdktypes.ConnectionID) bool {
	l := extrazap.ExtractLoggerFromContext(ctx)

	vs, err := v.Get(ctx, sdktypes.NewVarScopeID(cid), vars.UserScope)
	if err != nil {
		l.Error("Failed to get Google OAuth scopes", zap.Error(err))
		return false
	}

	return strings.Contains(vs.GetValue(vars.UserScope), "https://www.googleapis.com/auth/gmail")
}

// https://developers.google.com/gmail/api/reference/rest/v1/users/watch
func (a api) watch(ctx context.Context) (*gmail.WatchResponse, error) {
	client, err := a.gmailClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.Users.Watch("me", &gmail.WatchRequest{TopicName: topic}).Do()
	if err != nil {
		return nil, err
	}

	return resp, nil
}
