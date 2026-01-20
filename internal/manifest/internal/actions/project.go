package actions

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

type CreateProjectAction struct {
	Key     string           `json:"key"`
	Project sdktypes.Project `json:"project"`
}

func (a CreateProjectAction) Type() string   { return "create_project" }
func (a CreateProjectAction) isAction()      {}
func (a CreateProjectAction) GetKey() string { return a.Key }

func init() { registerActionType[CreateProjectAction]() }

// ---

type UpdateProjectAction struct {
	Key     string           `json:"key"`
	Project sdktypes.Project `json:"project"`
}

func (a UpdateProjectAction) Type() string   { return "update_project" }
func (a UpdateProjectAction) isAction()      {}
func (a UpdateProjectAction) GetKey() string { return a.Key }

func init() { registerActionType[UpdateProjectAction]() }

// ---

type TouchedProjectAction struct {
	Key       string             `json:"key"`
	ProjectID sdktypes.ProjectID `json:"project_id"`
}

func (a TouchedProjectAction) Type() string   { return "touch_project" }
func (a TouchedProjectAction) isAction()      {}
func (a TouchedProjectAction) GetKey() string { return a.Key }

func init() { registerActionType[TouchedProjectAction]() }
