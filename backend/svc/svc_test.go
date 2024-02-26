package svc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/backend/basesvc"
	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var opts = svc.NewOpts(kittehs.Must1(basesvc.LoadConfig("AK_TEST_APP", nil, "")), basesvc.RunOptions{})

func TestFxOptions(t *testing.T) {
	if err := fx.ValidateApp(opts...); err != nil {
		t.Logf("validate error: %v", err)
		t.Fail()
        // test
	}
}

func TestNew(t *testing.T) {
	app := fx.New(opts...)
	assert.NotNil(t, app)
}
