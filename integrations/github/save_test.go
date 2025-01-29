package github

import (
	"context"
	"net/url"
	"testing"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestSaveClientIDAndSecret(t *testing.T) {
	type fields struct {
		logger *zap.Logger
		oauth  sdkservices.OAuth
		vars   sdkservices.Vars
	}
	type args struct {
		ctx  context.Context
		c    sdkintegrations.ConnectionInit
		form url.Values
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "invalid connection ID",
			fields: fields{
				vars: &mockVars{},
			},
			args: args{
				ctx: context.Background(),
				c: sdkintegrations.ConnectionInit{
					ConnectionID: "invalid",
				},
				form: url.Values{},
			},
			wantErr: true,
		},
		{
			name: "basic save",
			fields: fields{
				vars: &mockVars{},
			},
			args: args{
				ctx: context.Background(),
				c: sdkintegrations.ConnectionInit{
					ConnectionID: "con_01234567890123456789012345",
				},
				form: url.Values{
					"client_id":     {"test-client"},
					"client_secret": {"test-secret"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handler{
				logger: tt.fields.logger,
				oauth:  tt.fields.oauth,
				vars:   tt.fields.vars,
			}
			if err := h.saveClientIDAndSecret(tt.args.ctx, tt.args.c, tt.args.form); (err != nil) != tt.wantErr {
				t.Errorf("handler.saveClientIDAndSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockVars struct{}

func (m *mockVars) Set(ctx context.Context, vars ...sdktypes.Var) error {
	return nil
}

func (m *mockVars) Delete(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) error {
	return nil
}

func (m *mockVars) Get(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) (sdktypes.Vars, error) {
	return nil, nil
}

func (m *mockVars) FindConnectionIDs(ctx context.Context, iid sdktypes.IntegrationID, name sdktypes.Symbol, value string) ([]sdktypes.ConnectionID, error) {
	return nil, nil
}
