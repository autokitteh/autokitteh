package svc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/kittehs"

	"go.autokitteh.dev/autokitteh/internal/backend/basesvc"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

var opts = svc.NewOpts(kittehs.Must1(basesvc.LoadConfig("AK_TEST_APP", nil, "")), basesvc.RunOptions{Mode: configset.Test})

func TestFxOptions(t *testing.T) {
	if err := fx.ValidateApp(opts...); err != nil {
		t.Logf("validate error: %v", err)
		t.Fail()
	}
}

func TestNew(t *testing.T) {
	app := fx.New(opts...)
	assert.NotNil(t, app)
}
