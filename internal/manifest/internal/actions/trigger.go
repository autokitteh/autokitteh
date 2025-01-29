package actions

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

type CreateTriggerAction struct {
	Key           string           `json:"key"`
	ConnectionKey *string          `json:"connection"`
	ProjectKey    string           `json:"project"`
	OrgID         sdktypes.OrgID   `json:"org_id,omitempty"`
	Trigger       sdktypes.Trigger `json:"trigger"`
}

func (a CreateTriggerAction) Type() string   { return "create_trigger" }
func (a CreateTriggerAction) isAction()      {}
func (a CreateTriggerAction) GetKey() string { return a.Key }

func init() { registerActionType[CreateTriggerAction]() }

// ---

type UpdateTriggerAction struct {
	Key           string           `json:"key"`
	ConnectionKey *string          `json:"connection"`
	ProjectKey    string           `json:"project"`
	OrgID         sdktypes.OrgID   `json:"org_id"`
	Trigger       sdktypes.Trigger `json:"trigger"`
}

func (a UpdateTriggerAction) Type() string   { return "update_trigger" }
func (a UpdateTriggerAction) isAction()      {}
func (a UpdateTriggerAction) GetKey() string { return a.Key }

func init() { registerActionType[UpdateTriggerAction]() }

// ---

type DeleteTriggerAction struct {
	Key       string             `json:"key"`
	TriggerID sdktypes.TriggerID `json:"trigger_id"`
}

func (a DeleteTriggerAction) Type() string   { return "delete_trigger" }
func (a DeleteTriggerAction) isAction()      {}
func (a DeleteTriggerAction) GetKey() string { return a.Key }

func init() { registerActionType[DeleteTriggerAction]() }
