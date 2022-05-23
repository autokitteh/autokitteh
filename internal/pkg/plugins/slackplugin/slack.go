package slackplugin

import (
	"context"

	"github.com/slack-go/slack"

	"go.autokitteh.dev/sdk/api/apivalues"
	"go.autokitteh.dev/sdk/pluginimpl"
)

var Plugin = &pluginimpl.Plugin{
	ID:  "slack",
	Doc: "TODO",
	Members: map[string]*pluginimpl.PluginMember{
		"open": pluginimpl.NewMethodMember(
			"TODO",
			func(
				ctx context.Context,
				name string,
				args []*apivalues.Value,
				kwargs map[string]*apivalues.Value,
				funcToValue pluginimpl.FuncToValueFunc,
			) (*apivalues.Value, error) {
				var token []byte

				if err := pluginimpl.UnpackArgs(
					args, kwargs,
					"token=", &token,
				); err != nil {
					return nil, err
				}

				api := slack.New(string(token))

				return pluginimpl.BuildStruct(
					funcToValue,
					"slack.client",
					pluginimpl.NewStructSimpleFuncMember(
						"post_message",
						"TODO",
						func(ctx context.Context, args []*apivalues.Value, kwargs map[string]*apivalues.Value) (*apivalues.Value, error) {
							var (
								channelID, text string
								esc             = true
							)

							err := pluginimpl.UnpackArgs(
								args, kwargs,
								"channel", &channelID,
								"text", &text,
								"escape?", &esc,
							)
							if err != nil {
								return nil, err
							}

							var resp struct {
								Channel   string `json:"channel"`
								Timestamp string `json:"timestamp"`
							}

							resp.Channel, resp.Timestamp, err = api.PostMessageContext(ctx, channelID, slack.MsgOptionText(text, esc))
							if err != nil {
								return nil, err
							}

							return wrap(resp)
						},
					),
					pluginimpl.NewStructSimpleFuncMember(
						"get_user_by_email",
						"TODO",
						func(ctx context.Context, args []*apivalues.Value, kwargs map[string]*apivalues.Value) (*apivalues.Value, error) {
							var email string

							if err := pluginimpl.UnpackArgs(args, kwargs, "email", &email); err != nil {
								return nil, err
							}

							user, err := api.GetUserByEmailContext(ctx, email)
							if err != nil {
								return nil, err
							}

							return wrap(user)
						},
					),
					pluginimpl.NewStructSimpleFuncMember(
						"get_user_by_email",
						"TODO",
						func(ctx context.Context, args []*apivalues.Value, kwargs map[string]*apivalues.Value) (*apivalues.Value, error) {
							var userID string

							if err := pluginimpl.UnpackArgs(args, kwargs, "user", &userID); err != nil {
								return nil, err
							}

							user, err := api.GetUserInfoContext(ctx, userID)
							if err != nil {
								return nil, err
							}

							return wrap(user)
						},
					),
				)
			},
		),
	},
}

func wrap(v interface{}) (*apivalues.Value, error) {
	return apivalues.Wrap(
		v,
		apivalues.WithWrapTranslate(
			func(in interface{}) (interface{}, error) {
				switch v := in.(type) {
				case slack.JSONTime:
					return v.Time(), nil
				default:
					return in, nil
				}
			},
		),
	)
}
