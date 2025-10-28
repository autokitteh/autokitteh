package store

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbtest"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	o = sdktypes.NewOrg().WithName(sdktypes.NewSymbol("o")).WithNewID()

	ps = []sdktypes.Project{
		sdktypes.NewProject().WithName(sdktypes.NewSymbol("p0")).WithOrgID(o.ID()).WithNewID(),
		sdktypes.NewProject().WithName(sdktypes.NewSymbol("p1")).WithOrgID(o.ID()).WithNewID(),
	}

	pids = kittehs.Transform(ps, func(p sdktypes.Project) sdktypes.ProjectID { return p.ID() })

	ivs = []sdktypes.Value{
		sdktypes.NewIntegerValue(0),
		sdktypes.NewIntegerValue(1),
		sdktypes.NewIntegerValue(2),
		sdktypes.NewIntegerValue(3),
	}
)

func TestMain(m *testing.M) {
	authz.DisableCheckForTesting()
	os.Exit(m.Run())
}

func TestMutate(t *testing.T) {
	db := dbtest.NewTestDB(t, o, ps[0], ps[1])

	store := New(db, zap.NewNop())

	// Each test is not independent - it relies on the previous state.
	tests := []struct {
		key      string
		op       string
		pid      sdktypes.ProjectID
		operands []sdktypes.Value
		ret      sdktypes.Value
		err      string
	}{
		{
			key: "k1",
			op:  "set",
			pid: pids[0],
			err: "missing value to set",
		},
		{
			key:      "k1",
			op:       "set",
			pid:      pids[0],
			operands: []sdktypes.Value{ivs[0], ivs[1]},
			err:      "too many operands",
		},
		{
			key:      "k1",
			op:       "get",
			pid:      pids[0],
			operands: []sdktypes.Value{ivs[0]},
			err:      "too many operands",
		},
		{
			key: "k1",
			op:  "get",
			pid: pids[0],
			ret: sdktypes.Nothing,
		},
		{
			key:      "k1",
			op:       "set",
			pid:      pids[0],
			operands: []sdktypes.Value{ivs[0]},
			ret:      sdktypes.Nothing,
		},
		{
			key: "k1",
			op:  "get",
			pid: pids[0],
			ret: ivs[0],
		},
		{
			key:      "k1",
			op:       "set",
			pid:      pids[1],
			operands: []sdktypes.Value{ivs[1]},
			ret:      sdktypes.Nothing,
		},
		{
			key:      "k2",
			op:       "set",
			pid:      pids[1],
			operands: []sdktypes.Value{ivs[2]},
			ret:      sdktypes.Nothing,
		},
		{
			key: "k1",
			op:  "get",
			pid: pids[0],
			ret: ivs[0],
		},
		{
			key: "k1",
			op:  "get",
			pid: pids[1],
			ret: ivs[1],
		},
		{
			key: "k1",
			op:  "del",
			pid: pids[0],
			ret: sdktypes.Nothing,
		},
		{
			key: "k1",
			op:  "get",
			pid: pids[0],
			ret: sdktypes.Nothing,
		},
		{
			key: "k1",
			op:  "get",
			pid: pids[1],
			ret: ivs[1],
		},
		{
			key: "k2",
			op:  "get",
			pid: pids[1],
			ret: ivs[2],
		},
		{
			key:      "k2",
			op:       "set",
			pid:      pids[1],
			operands: []sdktypes.Value{sdktypes.Nothing},
			ret:      sdktypes.Nothing,
		},
		{
			key: "k2",
			op:  "get",
			pid: pids[1],
			ret: sdktypes.Nothing,
		},
		{
			key:      "k3",
			op:       "add",
			pid:      pids[1],
			operands: []sdktypes.Value{ivs[1]},
			ret:      ivs[1],
		},
		{
			key: "k3",
			op:  "get",
			pid: pids[1],
			ret: ivs[1],
		},
		{
			key:      "k3",
			op:       "add",
			pid:      pids[1],
			operands: []sdktypes.Value{ivs[2]},
			ret:      ivs[3],
		},
		{
			key: "k3",
			op:  "get",
			pid: pids[1],
			ret: ivs[3],
		},
		{
			key:      "cas0",
			op:       "check_and_set",
			pid:      pids[0],
			operands: []sdktypes.Value{ivs[0], ivs[1]},
			ret:      sdktypes.FalseValue,
		},
		{
			key: "cas0",
			op:  "get",
			pid: pids[0],
			ret: sdktypes.Nothing,
		},
		{
			key:      "cas1",
			op:       "check_and_set",
			pid:      pids[0],
			operands: []sdktypes.Value{ivs[0], sdktypes.Nothing},
			ret:      sdktypes.TrueValue,
		},
		{
			key: "cas1",
			op:  "get",
			pid: pids[0],
			ret: ivs[0],
		},
		{
			key:      "cas1",
			op:       "check_and_set",
			pid:      pids[0],
			operands: []sdktypes.Value{ivs[1], ivs[0]},
			ret:      sdktypes.TrueValue,
		},
		{
			key: "cas1",
			op:  "get",
			pid: pids[0],
			ret: ivs[1],
		},
		{
			key:      "cas1",
			op:       "check_and_set",
			pid:      pids[0],
			operands: []sdktypes.Value{ivs[1], ivs[0]},
			ret:      sdktypes.FalseValue,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ret, err := store.Mutate(
				t.Context(),
				test.pid,
				test.key,
				test.op,
				test.operands...,
			)

			if test.err != "" {
				assert.Equal(t, test.err, err.Error())
			} else if assert.NoError(t, err) {
				assert.True(t, test.ret.Equal(ret), "expected %s, got %s", test.ret, ret)
			}
		})
	}
}
