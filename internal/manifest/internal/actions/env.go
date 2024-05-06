package actions

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type CreateEnvAction struct {
	Key        string       `json:"key"`
	ProjectKey string       `json:"project"`
	Env        sdktypes.Env `json:"env"`
}

func (a CreateEnvAction) Type() string   { return "create_env" }
func (a CreateEnvAction) isAction()      {}
func (a CreateEnvAction) GetKey() string { return a.Key }

func init() { registerActionType[CreateEnvAction]() }

// ---

type UpdateEnvAction struct {
	Key string       `json:"key"`
	Env sdktypes.Env `json:"env"`
}

func (a UpdateEnvAction) Type() string   { return "update_env" }
func (a UpdateEnvAction) isAction()      {}
func (a UpdateEnvAction) GetKey() string { return a.Key }

func init() { registerActionType[UpdateEnvAction]() }

// ---

type DeleteEnvAction struct {
	Key   string         `json:"key"`
	EnvID sdktypes.EnvID `json:"env_id"`
}

func (a DeleteEnvAction) Type() string   { return "delete_env" }
func (a DeleteEnvAction) isAction()      {}
func (a DeleteEnvAction) GetKey() string { return a.Key }

func init() { registerActionType[DeleteEnvAction]() }
