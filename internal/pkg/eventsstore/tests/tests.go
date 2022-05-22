//go:build unit_norace

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autokitteh/autokitteh/sdk/api/apievent"
	"github.com/autokitteh/autokitteh/sdk/api/apieventsrc"
	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
)

func TestAll(t *testing.T, s func() eventsstore.Store) {
	t.Run("addgetevent", func(t *testing.T) { TestAddAndGetEvent(t, s) })
	t.Run("listevents", func(t *testing.T) { TestListEvents(t, s) })
	t.Run("states", func(t *testing.T) { TestStates(t, s) })
	t.Run("liststates", func(t *testing.T) { TestListStates(t, s) })
}

func TestAddAndGetEvent(t *testing.T, sf func() eventsstore.Store) {
	ctx := context.Background()
	s := sf()
	srcid := apieventsrc.NewEventSourceID()

	id, err := s.Add(
		ctx,
		srcid,
		"",
		"",
		"test",
		map[string]*apivalues.Value{"v": apivalues.MustWrap(1)},
		map[string]string{"test": "true"},
	)
	require.NoError(t, err)

	e, err := s.Get(ctx, id)
	require.NoError(t, err)

	assert.Equal(t, id, e.ID())
	assert.Equal(t, srcid, e.SourceID())
	assert.EqualValues(t, map[string]string{"test": "true"}, e.Memo())
	assert.EqualValues(t, map[string]*apivalues.Value{"v": apivalues.MustWrap(1)}, e.Data())
}

func TestListEvents(t *testing.T, sf func() eventsstore.Store) {
	ctx := context.Background()
	s := sf()
	srcid := apieventsrc.NewEventSourceID()

	const N = 16
	ids := make([]apievent.EventID, N)
	for i := 0; i < N; i++ {
		var err error
		ids[N-i-1], err = s.Add(
			ctx,
			srcid,
			"",
			"",
			"test",
			map[string]*apivalues.Value{"i": apivalues.MustWrap(i)},
			nil,
		)
		require.NoError(t, err)
	}

	rids, err := s.List(ctx, nil, 0, 0)
	if assert.NoError(t, err) {
		assert.EqualValues(t, ids, rids)
	}

	rids, err = s.List(ctx, nil, 0, 1)
	if assert.NoError(t, err) {
		assert.EqualValues(t, ids[:1], rids)
	}

	rids, err = s.List(ctx, nil, 0, 5)
	if assert.NoError(t, err) {
		assert.EqualValues(t, ids[:5], rids)
	}

	rids, err = s.List(ctx, nil, 0, 500)
	if assert.NoError(t, err) {
		assert.EqualValues(t, ids, rids)
	}

	rids, err = s.List(ctx, nil, 5, 0)
	if assert.NoError(t, err) {
		assert.EqualValues(t, ids[5:], rids)
	}

	rids, err = s.List(ctx, nil, 5, 3)
	if assert.NoError(t, err) {
		assert.EqualValues(t, ids[5:8], rids)
	}
}

func TestStates(t *testing.T, sf func() eventsstore.Store) {
	ctx := context.Background()
	srcid := apieventsrc.NewEventSourceID()
	s := sf()
	projid := apiproject.NewProjectID()

	id, err := s.Add(
		ctx,
		srcid,
		"",
		"",
		"test",
		map[string]*apivalues.Value{"v": apivalues.MustWrap(1)},
		map[string]string{"test": "true"},
	)
	require.NoError(t, err)

	rs, err := s.GetStateForProject(ctx, id, projid)
	if assert.NoError(t, err) {
		assert.Len(t, rs, 0)
	}

	err = s.UpdateStateForProject(ctx, id, projid, apievent.NewProcessingProjectEventState())
	require.NoError(t, err)

	rs, err = s.GetStateForProject(ctx, id, projid)
	if assert.NoError(t, err) {
		assert.Len(t, rs, 1)
	}

	err = s.UpdateStateForProject(ctx, id, projid, apievent.NewProcessedProjectEventState(nil))
	require.NoError(t, err)

	rs, err = s.GetStateForProject(ctx, id, projid)
	if assert.NoError(t, err) {
		assert.Len(t, rs, 2)
	}
}

func TestListStates(t *testing.T, sf func() eventsstore.Store) {
	ctx := context.Background()
	s := sf()
	srcid := apieventsrc.NewEventSourceID()

	const N = 4
	ids := make([]apievent.EventID, N)
	projids := make([]apiproject.ProjectID, N)
	for i := 0; i < N; i++ {
		projid := apiproject.NewProjectID()

		id, err := s.Add(
			ctx,
			srcid,
			"",
			"",
			"test",
			map[string]*apivalues.Value{"i": apivalues.MustWrap(i)},
			nil,
		)
		require.NoError(t, err)

		ids[N-i-1] = id
		projids[N-i-1] = projid

		if i%2 == 0 {
			err = s.UpdateStateForProject(ctx, id, projid, apievent.NewProcessingProjectEventState())
			require.NoError(t, err)

			if i%4 == 0 {
				err = s.UpdateStateForProject(ctx, id, projid, apievent.NewProcessedProjectEventState(nil))
				require.NoError(t, err)
			}
		}
	}

	rids, err := s.List(ctx, &projids[N-1], 0, 0)
	if assert.NoError(t, err) {
		assert.EqualValues(t, ids[N-1:], rids)
	}

	rids, err = s.List(ctx, &projids[0], 0, 0)
	if assert.NoError(t, err) {
		assert.Empty(t, rids)
	}

	rids, err = s.List(ctx, &projids[1], 0, 0)
	if assert.NoError(t, err) {
		assert.EqualValues(t, ids[1:2], rids)
	}
}
