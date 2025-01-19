package sdktypes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalAnyObject(t *testing.T) {
	objs := []AnyObject{
		{NewProject().WithNewID()},
		{NewOrg().WithNewID()},
		{NewUser().WithNewID()},
		{NewOrgMember(NewOrgID(), NewUserID())},
	}

	bs, err := json.Marshal(objs)
	require.NoError(t, err)
	require.NotNil(t, bs)

	var aobjs []AnyObject
	if assert.NoError(t, json.Unmarshal(bs, &aobjs)) && assert.Len(t, aobjs, len(objs)) {
		for _, aobj := range aobjs {
			assert.Equal(t, aobj.ProtoMessage(), aobj.ProtoMessage())
		}
	}
}
