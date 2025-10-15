package integrations

import (
	"go.uber.org/zap"
	"logur.dev/logur/integration/grpc"

	"go.autokitteh.dev/autokitteh/integrations/airtable"
	"go.autokitteh.dev/autokitteh/integrations/anthropic"
	"go.autokitteh.dev/autokitteh/integrations/asana"
	"go.autokitteh.dev/autokitteh/integrations/atlassian/confluence"
	"go.autokitteh.dev/autokitteh/integrations/atlassian/jira"
	"go.autokitteh.dev/autokitteh/integrations/auth0"
	"go.autokitteh.dev/autokitteh/integrations/aws"
	"go.autokitteh.dev/autokitteh/integrations/azurebot"
	"go.autokitteh.dev/autokitteh/integrations/chatgpt"
	"go.autokitteh.dev/autokitteh/integrations/discord"
	"go.autokitteh.dev/autokitteh/integrations/github"
	"go.autokitteh.dev/autokitteh/integrations/google"
	"go.autokitteh.dev/autokitteh/integrations/google/calendar"
	"go.autokitteh.dev/autokitteh/integrations/google/drive"
	"go.autokitteh.dev/autokitteh/integrations/google/forms"
	"go.autokitteh.dev/autokitteh/integrations/google/gemini"
	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/sheets"
	"go.autokitteh.dev/autokitteh/integrations/google/youtube"
	"go.autokitteh.dev/autokitteh/integrations/height"
	"go.autokitteh.dev/autokitteh/integrations/hubspot"
	"go.autokitteh.dev/autokitteh/integrations/kubernetes"
	"go.autokitteh.dev/autokitteh/integrations/linear"
	"go.autokitteh.dev/autokitteh/integrations/microsoft"
	"go.autokitteh.dev/autokitteh/integrations/microsoft/teams"
	"go.autokitteh.dev/autokitteh/integrations/notion"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/integrations/pipedrive"
	"go.autokitteh.dev/autokitteh/integrations/reddit"
	"go.autokitteh.dev/autokitteh/integrations/salesforce"
	"go.autokitteh.dev/autokitteh/integrations/slack"
	"go.autokitteh.dev/autokitteh/integrations/telegram"
	"go.autokitteh.dev/autokitteh/integrations/twilio"
	"go.autokitteh.dev/autokitteh/integrations/zoom"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type Integration struct {
	Name  string
	Init  any
	Start func(*zap.Logger, *muxes.Muxes, sdkservices.Vars, *oauth.OAuth, sdkservices.DispatchFunc)
}

var all = []Integration{
	{"airtable", airtable.New, airtable.Start},
	{"anthropic", anthropic.New, func(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, _ *oauth.OAuth, _ sdkservices.DispatchFunc) {
		anthropic.Start(l, m, v)
	}},
	{"asana", asana.New, func(l *zap.Logger, m *muxes.Muxes, _ sdkservices.Vars, _ *oauth.OAuth, _ sdkservices.DispatchFunc) {
		asana.Start(l, m)
	}},
	{"auth0", auth0.New, func(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, _ *oauth.OAuth, _ sdkservices.DispatchFunc) {
		auth0.Start(l, m, v)
	}},
	{"aws", aws.New, func(l *zap.Logger, m *muxes.Muxes, _ sdkservices.Vars, _ *oauth.OAuth, _ sdkservices.DispatchFunc) {
		aws.Start(l, m)
	}},
	{"azurebot", azurebot.New, func(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, _ *oauth.OAuth, d sdkservices.DispatchFunc) {
		azurebot.Start(l, m, v, d)
	}},
	{"calendar", calendar.New, nil},
	{"chatgpt", chatgpt.New, func(l *zap.Logger, m *muxes.Muxes, _ sdkservices.Vars, _ *oauth.OAuth, _ sdkservices.DispatchFunc) {
		chatgpt.Start(l, m)
	}},
	{"confluence", confluence.New, confluence.Start},
	{"discord", discord.New, func(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, _ *oauth.OAuth, d sdkservices.DispatchFunc) {
		discord.Start(l, m, v, d)
	}},
	{"drive", drive.New, nil},
	{"forms", forms.New, nil},
	{"gemini", gemini.New, func(l *zap.Logger, m *muxes.Muxes, _ sdkservices.Vars, _ *oauth.OAuth, _ sdkservices.DispatchFunc) {
		gemini.Start(l, m)
	}},
	{"github", github.New, github.Start},
	{"gmail", gmail.New, nil},
	{"google", google.New, google.Start},
	{"grpc", grpc.New, nil},
	{"height", height.New, height.Start},
	{"hubspot", hubspot.New, hubspot.Start},
	{"jira", jira.New, jira.Start},
	{"kubernetes", kubernetes.New, func(l *zap.Logger, m *muxes.Muxes, _ sdkservices.Vars, _ *oauth.OAuth, _ sdkservices.DispatchFunc) {
		kubernetes.Start(l, m)
	}},
	{"linear", linear.New, linear.Start},
	{"microsoft_teams", teams.New, nil},
	{"microsoft", microsoft.New, microsoft.Start},
	{"notion", notion.New, func(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, _ *oauth.OAuth, _ sdkservices.DispatchFunc) {
		notion.Start(l, m, v)
	}},
	{"pipedrive", pipedrive.New, func(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, _ *oauth.OAuth, _ sdkservices.DispatchFunc) {
		pipedrive.Start(l, m, v)
	}},
	{"reddit", reddit.New, func(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, _ *oauth.OAuth, _ sdkservices.DispatchFunc) {
		reddit.Start(l, m, v)
	}},
	{"salesforce", salesforce.New, salesforce.Start},
	{"sheets", sheets.New, nil},
	{"slack", slack.New, func(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, _ *oauth.OAuth, d sdkservices.DispatchFunc) {
		slack.Start(l, m, v, d)
	}},
	{"telegram", telegram.New, func(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, _ *oauth.OAuth, d sdkservices.DispatchFunc) {
		telegram.Start(l, m, v, d)
	}},
	{"twilio", twilio.New, func(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, _ *oauth.OAuth, d sdkservices.DispatchFunc) {
		twilio.Start(l, m, v, d)
	}},
	{"youtube", youtube.New, nil},
	{"zoom", zoom.New, zoom.Start},
}

func All() []Integration { return all }

func Names() []string {
	return kittehs.Transform(all, func(i Integration) string { return i.Name })
}
