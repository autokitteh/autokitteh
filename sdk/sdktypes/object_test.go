package sdktypes_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var _ sdktypes.Object = sdktypes.Project{}

func TestMisc(t *testing.T) {
	var zero sdktypes.Project

	assert.True(t, !zero.IsValid())
	assert.Nil(t, zero.ToProto())
	assert.Nil(t, sdktypes.ToProto(zero))

	p1id := sdktypes.NewProjectID().String()

	p1, err := sdktypes.ProjectFromProto(&sdktypes.ProjectPB{
		ProjectId: p1id,
		Name:      "p1",
	})
	assert.NoError(t, err)

	p11, err := sdktypes.ProjectFromProto(&sdktypes.ProjectPB{
		ProjectId: p1.ID().String(),
		Name:      "p1",
	})
	assert.NoError(t, err)

	p2, err := sdktypes.ProjectFromProto(&sdktypes.ProjectPB{Name: "p2"})
	assert.NoError(t, err)

	assert.Equal(t, "", zero.ID().String())

	assert.True(t, p1.Equal(p11))
	assert.True(t, !p1.Equal(p2))

	p3 := p1.WithName(kittehs.Must1(sdktypes.ParseSymbol("p3")))
	assert.Equal(t, "p3", p3.Name().String())
	assert.Equal(t, "p1", p1.Name().String())
	assert.Equal(t, p1id, p1.ID().String())
	assert.Equal(t, p1id, p1.ID().String())

	assert.Equal(t, "", zero.Hash())

	jsonP3 := `{"project_id":"` + p3.ID().String() + `","name":"p3"}`
	bs, err := json.Marshal(p3)
	if assert.NoError(t, err) {
		assert.JSONEq(t, jsonP3, string(bs))
	}

	var p4 sdktypes.Project
	bs, err = json.Marshal(p4)
	if assert.NoError(t, err) {
		assert.Equal(t, `null`, string(bs))
	}

	if assert.NoError(t, p4.UnmarshalJSON([]byte(jsonP3))) {
		assert.True(t, p3.Equal(p4))
	}
}

// If this test fails, the hash function changed. This will cause incompability with existing data.
func TestStableObjectHash(t *testing.T) {
	n := kittehs.Must1(sdktypes.ParseSymbol("test"))
	p := sdktypes.NewProject(sdktypes.InvalidProjectID).WithName(n)
	assert.Equal(t, "357785dbbe12ae6fa6b63134a5163e5ea70db04ae8b151864a5b1e72c3a5bd6e", p.Hash())
}
