package actions

import (
	"encoding/json"
	"fmt"
)

type Actions []Action

var unmarshallers = make(map[string]func([]byte) (Action, error))

func registerActionType[A Action]() {
	var a A
	unmarshallers[a.Type()] = func(data []byte) (Action, error) {
		var a A
		if err := json.Unmarshal(data, &a); err != nil {
			return nil, err
		}
		return a, nil
	}
}

func (as Actions) MarshalJSON() ([]byte, error) {
	type action struct {
		Type   string `json:"type"`
		Action `json:"action"`
	}

	actions := make([]action, len(as))
	for i, a := range as {
		actions[i] = action{
			Type:   a.Type(),
			Action: a,
		}
	}

	return json.Marshal(actions)
}

func unmarshal(t string, data []byte) (Action, error) {
	unmarshal, ok := unmarshallers[t]
	if !ok {
		return nil, fmt.Errorf("unknown action type: %s", t)
	}

	return unmarshal(data)
}

func UnmarshalActionJSON(data []byte) (Action, error) {
	type anyAction struct {
		Type   string          `json:"type"`
		Action json.RawMessage `json:"action"`
	}

	var a anyAction
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, err
	}

	return unmarshal(a.Type, a.Action)
}

func (as *Actions) UnmarshalJSON(data []byte) error {
	type action struct {
		Type   string          `json:"type"`
		Action json.RawMessage `json:"action"`
	}

	var actions []action
	if err := json.Unmarshal(data, &actions); err != nil {
		return err
	}

	*as = make(Actions, len(actions))
	for i, a := range actions {
		var err error
		if (*as)[i], err = unmarshal(a.Type, a.Action); err != nil {
			return err
		}
	}

	return nil
}
