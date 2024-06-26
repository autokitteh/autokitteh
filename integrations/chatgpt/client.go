package chatgpt

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct{ vars sdkservices.Vars }

var integrationID = sdktypes.NewIntegrationIDFromName("chatgpt")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "chatgpt",
	DisplayName:   "OpenAI ChatGPT",
	Description:   "ChatGPT is a conversational AI model that can generates human-like responses based on prompts.",
	LogoUrl:       "/static/images/chatgpt.svg",
	UserLinks: map[string]string{
		"1 OpenAI developer platform": "https://platform.openai.com/",
		"2 Go client API":             "https://pkg.go.dev/github.com/sashabaranov/go-openai",
	},
	ConnectionUrl: "/chatgpt/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(vars sdkservices.Vars) sdkservices.Integration {
	i := integration{vars: vars}
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(
			sdkmodule.ExportFunction(
				"create_chat_completion",
				i.createChatCompletion,
				sdkmodule.WithFuncDoc("https://pkg.go.dev/github.com/sashabaranov/go-openai#Client.CreateChatCompletion"),
				sdkmodule.WithArgs("model?", "message?", "messages?"),
			),
		),
		sdkintegrations.WithConnectionConfigFromVars(vars),
	)
}
