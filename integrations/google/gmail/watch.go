package gmail

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/api/gmail/v1"

	"go.autokitteh.dev/autokitteh/integrations/google/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// pubsubTopicEnvVar is the name of an environment variable that
	// contains the GCP Pub/Sub topic name for Gmail notifications.
	pubsubTopicEnvVar = "GMAIL_PUBSUB_TOPIC"
)

// UpdateWatch creates or updates a push notification watch on the user's mailbox.
func UpdateWatch(ctx context.Context, v sdkservices.Vars, cid sdktypes.ConnectionID) error {
	l := extrazap.ExtractLoggerFromContext(ctx)

	// Not a Gmail user? Nothing to do.
	if !gotGmailScope(ctx, v, cid) {
		l.Debug("No Gmail scope, skipping Gmail user mailbox watch")
		return nil
	}

	watch, err := api{vars: v, cid: cid}.watch(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Gmail watch: %w", err)
	}

	hid := strconv.FormatUint(watch.HistoryId, 10)
	exp := time.Unix(watch.Expiration/1000, 0).UTC()
	vs := sdktypes.NewVars(
		sdktypes.NewVar(vars.GmailHistoryID).SetValue(hid),
		sdktypes.NewVar(vars.GmailWatchExpiration).SetValue(exp.Format(time.RFC3339)),
	).WithScopeID(sdktypes.NewVarScopeID(cid))

	if err := v.Set(ctx, vs...); err != nil {
		return fmt.Errorf("failed to save Gmail watch: %w", err)
	}

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

	resp, err := client.Users.Watch("me", &gmail.WatchRequest{
		TopicName: os.Getenv(pubsubTopicEnvVar),
	}).Do()
	if err != nil {
		return nil, err
	}

	return resp, nil
}
