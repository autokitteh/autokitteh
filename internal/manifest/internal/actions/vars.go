package actions

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type SetVarAction struct {
	Key        string         `json:"key"`
	Project    string         `json:"project,omitempty"`
	OrgID      sdktypes.OrgID `json:"org_id,omitempty"`
	Connection string         `json:"connection,omitempty"`
	Var        sdktypes.Var   `json:"var"`
}

func (a SetVarAction) Type() string   { return "set_var" }
func (a SetVarAction) isAction()      {}
func (a SetVarAction) GetKey() string { return a.Key }

func init() { registerActionType[SetVarAction]() }

// ---

type DeleteVarAction struct {
	Key     string              `json:"key"`
	ScopeID sdktypes.VarScopeID `json:"scope_id"`
	Name    string              `json:"var_name"`
}

func (a DeleteVarAction) Type() string   { return "delete_var" }
func (a DeleteVarAction) isAction()      {}
func (a DeleteVarAction) GetKey() string { return a.Key }

func init() { registerActionType[DeleteVarAction]() }
