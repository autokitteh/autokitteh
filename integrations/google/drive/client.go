package drive

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/google/connections"
	"go.autokitteh.dev/autokitteh/integrations/google/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	integrationName = "googledrive"
)

var (
	IntegrationID = sdktypes.NewIntegrationIDFromName(integrationName)

	desc = common.Descriptor(integrationName, "Google Drive", "/static/images/google_drive.svg")
)

type api struct {
	logger *zap.Logger
	vars   sdkservices.Vars
	cid    sdktypes.ConnectionID
}

func New(cvars sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(),
		connections.ConnStatus(cvars),
		connections.ConnTest(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

func (a api) driveClient(ctx context.Context) (*drive.Service, error) {
	data, err := a.connectionData(ctx)
	if err != nil {
		return nil, err
	}

	var src oauth2.TokenSource
	if data.OAuthData != "" {
		src, err = oauthTokenSource(ctx, data.OAuthData)
	} else {
		src, err = jwtTokenSource(ctx, data.JSON)
	}
	if err != nil {
		return nil, err
	}

	svc, err := drive.NewService(ctx, option.WithTokenSource(src))
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func oauthTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	d, err := sdkintegrations.DecodeOAuthData(data)
	if err != nil {
		return nil, err
	}

	return oauthConfig().TokenSource(ctx, d.Token), nil
}

func jwtTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	scopes := oauthConfig().Scopes

	cfg, err := google.JWTConfigFromJSON([]byte(data), scopes...)
	if err != nil {
		return nil, err
	}

	return cfg.TokenSource(ctx), nil
}

func oauthConfig() *oauth2.Config {
	addr := os.Getenv("WEBHOOK_ADDRESS")
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf("https://%s/oauth/redirect/google", addr),
		// https://developers.google.com/drive/api/guides/api-specific-auth
		Scopes: []string{
			// Non-sensitive.
			googleoauth2.OpenIDScope,
			googleoauth2.UserinfoEmailScope,
			googleoauth2.UserinfoProfileScope,
			// Sensitive.
			drive.DriveFileScope,
			// drive.DriveScope, // See ENG-1701
		},
	}
}

func (a api) connectionData(ctx context.Context) (*vars.Vars, error) {
	cid, err := sdkmodule.FunctionConnectionIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if !cid.IsValid() {
		cid = a.cid // Fallback during authentication flows.
	}

	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return nil, err
	}

	var decoded vars.Vars
	vs.Decode(&decoded)
	return &decoded, nil
}

// https://developers.google.com/drive/api/guides/push
// https://developers.google.com/drive/api/reference/rest/v3/changes/watch
func (a api) watchEvents(ctx context.Context, connID sdktypes.ConnectionID, userEmail string) (*drive.Channel, error) {
	client, err := a.driveClient(ctx)
	if err != nil {
		return nil, err
	}

	startToken, err := client.Changes.GetStartPageToken().Do()
	if err != nil {
		return nil, err
	}

	// Save the start page token for the next request.
	v := sdktypes.NewVar(vars.DriveChangesStartPageToken).SetValue(startToken.StartPageToken)
	if err = a.vars.Set(ctx, v.WithScopeID(sdktypes.NewVarScopeID(a.cid))); err != nil {
		return nil, err
	}

	addr := os.Getenv("WEBHOOK_ADDRESS")
	watchID := connID.String() + "/events"
	req := client.Changes.Watch(startToken.StartPageToken, &drive.Channel{
		Id:      watchID,
		Token:   userEmail + "/" + watchID,
		Address: fmt.Sprintf("https://%s/googledrive/notif", addr),
		Type:    "web_hook",
		// Expiration: time.Now().Add(time.Hour*24*7).Unix() * 1000,
		Expiration: time.Now().Add(time.Hour*24).Unix() * 1000,
	})

	// check if the channel already exists
	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(connID), vars.DriveEventsWatchID)
	if err != nil {
		return nil, err
	}
	wid := vs.Get(vars.DriveEventsWatchID).Value()
	if wid != "" {
		// sanity check
		if watchID != wid {
			return nil, fmt.Errorf("watch ID mismatch: %s != %s", watchID, wid)
		}
		// channel already exists, stop it
		if err := a.stopWatch(ctx, connID); err != nil {
			return nil, fmt.Errorf("stop existing watch channel: %w", err)
		}
	}

	resp, err := req.Do()
	if err != nil {
		gerr, _ := err.(*googleapi.Error)
		a.logger.Warn("Google Drive watch channel creation error", zap.Any("googleApiError", gerr))
		return nil, err
	}

	return resp, nil
}

// https://developers.google.com/drive/api/reference/rest/v3/channels/stop
func (a api) stopWatch(ctx context.Context, cid sdktypes.ConnectionID) error {
	client, err := a.driveClient(ctx)
	if err != nil {
		return err
	}

	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(cid), vars.DriveEventsWatchResID)
	if err != nil {
		return err
	}

	err = client.Channels.Stop(&drive.Channel{
		Id:         cid.String() + "/events",
		ResourceId: vs.Get(vars.DriveEventsWatchResID).Value(),
	}).Do()
	if err != nil {
		return err
	}
	return nil
}

func (a api) createChangeListRequest(client *drive.Service, pageToken string) *drive.ChangesListCall {
	return client.Changes.List(pageToken).
		IncludeItemsFromAllDrives(true).
		SupportsAllDrives(true).
		IncludeCorpusRemovals(true)
}

func (a api) initializeChangeTracking(ctx context.Context) error {
	client, err := a.driveClient(ctx)
	if err != nil {
		return err
	}

	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(a.cid))
	if err != nil {
		return err
	}
	startPageToken := vs.Get(vars.DriveChangesStartPageToken).Value()

	// Initial request
	resp, err := a.createChangeListRequest(client, startPageToken).Do()
	if err != nil {
		return err
	}

	// Handle pagination
	for resp.NextPageToken != "" {
		a.logger.Debug("Requesting next page of Google Drive changes",
			zap.String("pageToken", resp.NextPageToken),
		)

		resp, err = a.createChangeListRequest(client, resp.NextPageToken).Do()
		if err != nil {
			return err
		}
	}

	// After processing all pages, save the newStartPageToken for future change requests
	if resp.NewStartPageToken != "" {
		v := sdktypes.NewVar(vars.DriveChangesStartPageToken).SetValue(resp.NewStartPageToken)
		if err = a.vars.Set(ctx, v.WithScopeID(sdktypes.NewVarScopeID(a.cid))); err != nil {
			return err
		}
		a.logger.Debug("Updated Google Drive changes start page token",
			zap.String("cid", a.cid.String()),
			zap.String("newStartPageToken", resp.NewStartPageToken),
		)
	}

	return nil
}

func (a api) listChanges(ctx context.Context) ([]*drive.Change, error) {
	client, err := a.driveClient(ctx)
	if err != nil {
		return nil, err
	}

	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(a.cid))
	if err != nil {
		return nil, err
	}

	startPageToken := vs.Get(vars.DriveChangesStartPageToken).Value()
	a.logger.Debug("Google Drive connection's existing start page token",
		zap.String("cid", a.cid.String()),
		zap.String("startPageToken", startPageToken),
	)

	var changes []*drive.Change

	// Initial request with the stored page token
	req := a.createChangeListRequest(client, startPageToken)
	resp, err := req.Do()
	if err != nil {
		return nil, err
	}
	changes = append(changes, resp.Changes...)

	// Handle pagination
	for resp.NextPageToken != "" {
		a.logger.Debug("Requesting next page of Google Drive changes",
			zap.String("pageToken", resp.NextPageToken),
		)

		req = a.createChangeListRequest(client, resp.NextPageToken)
		resp, err = req.Do()
		if err != nil {
			return nil, err
		}

		changes = append(changes, resp.Changes...)
	}

	// Save the new start page token for future requests
	if resp.NewStartPageToken != "" {
		v := sdktypes.NewVar(vars.DriveChangesStartPageToken).SetValue(resp.NewStartPageToken)
		if err := a.vars.Set(ctx, v.WithScopeID(sdktypes.NewVarScopeID(a.cid))); err != nil {
			return nil, err
		}
		a.logger.Debug("Updated Google Drive changes start page token",
			zap.String("cid", a.cid.String()),
			zap.String("newStartPageToken", resp.NewStartPageToken),
		)
	}
	return changes, nil
}
