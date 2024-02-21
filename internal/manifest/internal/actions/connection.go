package actions

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

type CreateConnectionAction struct {
	Key            string              `json:"key"`
	Connection     sdktypes.Connection `json:"connection"`
	ProjectKey     string              `json:"project"`
	IntegrationKey string              `json:"integration"`
}

func (a CreateConnectionAction) Type() string   { return "create_connection" }
func (a CreateConnectionAction) isAction()      {}
func (a CreateConnectionAction) GetKey() string { return a.Key }

func init() { registerActionType[CreateConnectionAction]() }

// ---

type UpdateConnectionAction struct {
	Key        string              `json:"key"`
	Connection sdktypes.Connection `json:"connection"`
}

func (a UpdateConnectionAction) Type() string   { return "update_connection" }
func (a UpdateConnectionAction) isAction()      {}
func (a UpdateConnectionAction) GetKey() string { return a.Key }

func init() { registerActionType[UpdateConnectionAction]() }

// ---

type DeleteConnectionAction struct {
	Key          string                `json:"key"`
	ConnectionID sdktypes.ConnectionID `json:"connection_id"`
}

func (a DeleteConnectionAction) Type() string   { return "delete_connection" }
func (a DeleteConnectionAction) isAction()      {}
func (a DeleteConnectionAction) GetKey() string { return a.Key }

func init() { registerActionType[DeleteConnectionAction]() }
