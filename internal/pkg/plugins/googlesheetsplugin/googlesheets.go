package googlesheetsplugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"go.autokitteh.dev/sdk/api/apivalues"
	"go.autokitteh.dev/sdk/pluginimpl"

	"github.com/autokitteh/autokitteh/internal/pkg/googleoauth"
)

var (
	oauthConfig     *oauth2.Config
	onceOAuthConfig sync.Once
)

func clientFromToken(ctx context.Context, token []byte) (*http.Client, error) {
	var err error

	onceOAuthConfig.Do(func() {
		var cfg googleoauth.Config

		// must be the same as [# google-oauth-config #]
		if err = envconfig.Process("AKD_GOOGLE_OAUTH", &cfg); err != nil {
			return
		}

		if cfg.ClientID == "" || cfg.ClientSecret == "" {
			err = errors.New("missing client id and/or secret in oauth env variables")
			return
		}

		oauthConfig = googleoauth.MakeConfig(cfg)
	})

	if err != nil {
		return nil, err
	}

	var oauthToken oauth2.Token
	if err := json.Unmarshal(token, &oauthToken); err != nil {
		return nil, fmt.Errorf("unmarshal token: %w", err)
	}

	return oauthConfig.Client(ctx, &oauthToken), nil
}

var Plugin = &pluginimpl.Plugin{
	ID:  "googlesheets",
	Doc: "TODO",
	Members: map[string]*pluginimpl.PluginMember{
		"open": pluginimpl.NewMethodMember(
			"TODO",
			func(
				_ context.Context,
				name string,
				args []*apivalues.Value,
				kwargs map[string]*apivalues.Value,
				funcToValue pluginimpl.FuncToValueFunc,
			) (*apivalues.Value, error) {
				var token []byte

				if err := pluginimpl.UnpackArgs(args, kwargs, "token", &token); err != nil {
					return nil, err
				}

				// clientFromToken seems to retain the context after it returns, so in order
				// to avoid context canceled we just use context.Background here.
				// Futher calls will use .Context() construct to use actual context.
				ctx := context.Background()

				client, err := clientFromToken(ctx, token)
				if err != nil {
					return nil, err
				}

				srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
				if err != nil {
					return nil, fmt.Errorf("new sheets srv: %w", err)
				}

				return newSheets(funcToValue, srv), nil
			},
		),
	},
}

func newSheets(funcToValue pluginimpl.FuncToValueFunc, srv *sheets.Service) *apivalues.Value {
	return apivalues.Struct(
		apivalues.Symbol("sheets"),
		map[string]*apivalues.Value{
			// TODO: maybe can autogenerate functions?
			// See https://workflows.googleapis.com/$discovery/rest?version=v1.
			// See https://github.com/googleapis/google-api-go-client/blob/main/api-list.json#L5462.
			// See https://github.com/googleapis/google-api-go-client/blob/main/google-api-go-generator/gen.go.
			"values": pluginimpl.MustBuildStruct(
				funcToValue,
				"sheets.values",
				pluginimpl.NewStructSimpleFuncMember(
					"get",
					"TODO",
					func(ctx context.Context, args []*apivalues.Value, kws map[string]*apivalues.Value) (*apivalues.Value, error) {
						var id, readRange string

						if err := pluginimpl.UnpackArgs(args, kws, "id", &id, "range", &readRange); err != nil {
							return nil, err
						}

						resp, err := srv.Spreadsheets.Values.Get(id, readRange).Context(ctx).Do()
						if err != nil {
							return nil, fmt.Errorf("get: %w", err)
						}

						return apivalues.Wrap(resp.Values)
					},
				),
				pluginimpl.NewStructSimpleFuncMember(
					"clear",
					"TODO",
					func(ctx context.Context, args []*apivalues.Value, kws map[string]*apivalues.Value) (*apivalues.Value, error) {
						var id, clearRange string

						if err := pluginimpl.UnpackArgs(args, kws, "id", &id, "range", &clearRange); err != nil {
							return nil, err
						}

						_, err := srv.Spreadsheets.Values.Get(id, clearRange).Context(ctx).Do()
						if err != nil {
							return nil, fmt.Errorf("clear: %w", err)
						}

						return apivalues.None, nil
					},
				),
				pluginimpl.NewStructSimpleFuncMember(
					"update",
					"TODO",
					func(ctx context.Context, args []*apivalues.Value, kws map[string]*apivalues.Value) (*apivalues.Value, error) {
						var (
							id, updateRange string
							values          *apivalues.Value
							vio             = "RAW"
						)

						if err := pluginimpl.UnpackArgs(args, kws, "id", &id, "range", &updateRange, "values", &values, "value_input_option?", &vio); err != nil {
							return nil, err
						}

						var govs [][]interface{}

						if vvvs, ok := values.Get().(apivalues.ListValue); ok {
							govs = make([][]interface{}, len(vvvs))
							for i, vvs := range vvvs {
								if vs, ok := vvs.Get().(apivalues.ListValue); ok {
									l := make([]interface{}, len(vs))

									for j, v := range vs {
										l[j] = v.String()
									}

									govs[i] = l
									continue
								}

								govs = nil
								break
							}
						}

						if govs == nil {
							return nil, fmt.Errorf("values must be a list of list of strings")
						}

						_, err := srv.Spreadsheets.
							Values.
							Update(id, updateRange, &sheets.ValueRange{Values: govs}).
							ValueInputOption(vio).
							Context(ctx).Do()
						if err != nil {
							return nil, fmt.Errorf("update: %w", err)
						}

						return apivalues.None, nil
					},
				),
				pluginimpl.NewStructSimpleFuncMember(
					"append",
					"TODO",
					func(ctx context.Context, args []*apivalues.Value, kws map[string]*apivalues.Value) (*apivalues.Value, error) {
						var (
							id, updateRange string
							values          *apivalues.Value
							vio             = "RAW"
						)

						if err := pluginimpl.UnpackArgs(args, kws, "id", &id, "range", &updateRange, "values", &values, "value_input_option?", &vio); err != nil {
							return nil, err
						}

						var govs [][]interface{}

						if vvvs, ok := values.Get().(apivalues.ListValue); ok {
							govs = make([][]interface{}, len(vvvs))
							for i, vvs := range vvvs {
								if vs, ok := vvs.Get().(apivalues.ListValue); ok {
									l := make([]interface{}, len(vs))

									for j, v := range vs {
										l[j] = v.String()
									}

									govs[i] = l
									continue
								}

								govs = nil
								break
							}
						}

						if govs == nil {
							return nil, fmt.Errorf("values must be a list of list of strings")
						}

						_, err := srv.Spreadsheets.
							Values.
							Append(id, updateRange, &sheets.ValueRange{Values: govs}).
							ValueInputOption(vio).
							Context(ctx).Do()
						if err != nil {
							return nil, fmt.Errorf("append: %w", err)
						}

						return apivalues.None, nil
					},
				),
			),
		},
	)
}
