//go:build unit_norace && !race

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiaccount"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsrcsstore"
)

func TestAll(t *testing.T, sf func() eventsrcsstore.Store) {
	t.Run("src", func(t *testing.T) { TestSrc(t, sf) })
	t.Run("binding", func(t *testing.T) { TestBinding(t, sf) })
}

func TestSrc(t *testing.T, sf func() eventsrcsstore.Store) {
	ctx := context.Background()
	s := sf()

	aid := apiaccount.NewAccountID()

	id, err := s.Add(ctx, aid, eventsrcsstore.AutoEventSourceID, &apieventsrc.EventSourceData{})
	require.NoError(t, err)

	src, err := s.Get(ctx, id)
	if assert.NoError(t, err) {
		assert.Equal(t, id, src.ID())
		assert.Equal(t, aid, src.AccountID())
		assert.False(t, src.Data().Enabled())
	}

	require.NoError(t, s.Update(ctx, id, (&apieventsrc.EventSourceData{}).SetEnabled(true)))

	src, err = s.Get(ctx, id)
	if assert.NoError(t, err) {
		assert.Equal(t, id, src.ID())
		assert.Equal(t, aid, src.AccountID())
		assert.True(t, src.Data().Enabled())
	}

	ids, err := s.List(ctx, &aid)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []apieventsrc.EventSourceID{id}, ids)
	}

	aid2 := apiaccount.NewAccountID()

	id1, err := s.Add(ctx, aid, eventsrcsstore.AutoEventSourceID, &apieventsrc.EventSourceData{})
	require.NoError(t, err)

	id2, err := s.Add(ctx, aid2, eventsrcsstore.AutoEventSourceID, &apieventsrc.EventSourceData{})
	require.NoError(t, err)

	ids, err = s.List(ctx, &aid)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []apieventsrc.EventSourceID{id, id1}, ids)
	}

	ids, err = s.List(ctx, &aid2)
	if assert.NoError(t, err) {
		assert.EqualValues(t, []apieventsrc.EventSourceID{id2}, ids)
	}
}

func TestBinding(t *testing.T, sf func() eventsrcsstore.Store) {
	ctx := context.Background()
	s := sf()

	sid1 := apieventsrc.NewEventSourceID()
	sid2 := apieventsrc.NewEventSourceID()
	pid1 := apiproject.NewProjectID()
	pid2 := apiproject.NewProjectID()
	pid3 := apiproject.NewProjectID()

	require.NoError(t, s.AddProjectBinding(ctx, sid1, pid1, "", "", (&apieventsrc.EventSourceProjectBindingData{}).SetName("f1to1").SetEnabled(true)))
	require.NoError(t, s.AddProjectBinding(ctx, sid1, pid2, "", "", (&apieventsrc.EventSourceProjectBindingData{}).SetName("f1to2").SetEnabled(true)))
	require.NoError(t, s.AddProjectBinding(ctx, sid2, pid3, "", "", (&apieventsrc.EventSourceProjectBindingData{}).SetName("f2to3").SetEnabled(true)))
	require.NoError(t, s.UpdateProjectBinding(ctx, sid1, pid1, (&apieventsrc.EventSourceProjectBindingData{}).SetName("f1to1_updated").SetEnabled(false)))

	bs, err := s.GetProjectBindings(ctx, &sid1, &pid1, "")
	if assert.NoError(t, err) {
		for i, b := range bs {
			bs[i] = b.WithoutTimes()
		}

		assert.EqualValues(
			t,
			bs,
			[]*apieventsrc.EventSourceProjectBinding{
				apieventsrc.MustNewEventSourceProjectBinding(
					sid1, pid1, "", "", (&apieventsrc.EventSourceProjectBindingData{}).SetName("f1to1_updated"), time.Now(), nil,
				).WithoutTimes(),
			},
		)
	}

	bs, err = s.GetProjectBindings(ctx, &sid1, nil, "")
	if assert.NoError(t, err) {
		for i, b := range bs {
			bs[i] = b.WithoutTimes()
		}

		assert.EqualValues(
			t,
			bs,
			[]*apieventsrc.EventSourceProjectBinding{
				apieventsrc.MustNewEventSourceProjectBinding(
					sid1, pid1, "", "", (&apieventsrc.EventSourceProjectBindingData{}).SetName("f1to1_updated"), time.Time{}, nil,
				).WithoutTimes(),
				apieventsrc.MustNewEventSourceProjectBinding(
					sid1, pid2, "", "", (&apieventsrc.EventSourceProjectBindingData{}).SetName("f1to2").SetEnabled(true), time.Time{}, nil,
				).WithoutTimes(),
			},
		)
	}

	bs, err = s.GetProjectBindings(ctx, nil, &pid1, "")
	if assert.NoError(t, err) {
		for i, b := range bs {
			bs[i] = b.WithoutTimes()
		}

		assert.EqualValues(
			t,
			bs,
			[]*apieventsrc.EventSourceProjectBinding{
				apieventsrc.MustNewEventSourceProjectBinding(
					sid1, pid1, "", "", (&apieventsrc.EventSourceProjectBindingData{}).SetName("f1to1_updated"), time.Time{}, nil,
				).WithoutTimes(),
			},
		)
	}

	bs, err = s.GetProjectBindings(ctx, nil, &pid3, "")
	if assert.NoError(t, err) {
		for i, b := range bs {
			bs[i] = b.WithoutTimes()
		}

		assert.EqualValues(
			t,
			bs,
			[]*apieventsrc.EventSourceProjectBinding{
				apieventsrc.MustNewEventSourceProjectBinding(
					sid2, pid3, "", "", (&apieventsrc.EventSourceProjectBindingData{}).SetName("f2to3").SetEnabled(true), time.Time{}, nil,
				).WithoutTimes(),
			},
		)
	}
}
