package projectsstoregorm

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"

	"github.com/autokitteh/autokitteh/sdk/api/apiaccount"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
)

type project struct {
	ID          string `gorm:"primaryKey"`
	AccountName string `gorm:"index;not null"`
	Enabled     bool   `gorm:"index"`
	MainPath    string
	Predecls    datatypes.JSON
	Plugins     datatypes.JSON
	Name        string `gorm:"index"`
	Memo        datatypes.JSON
	CreatedAt   time.Time // only set in initial creation, then carried over.
	UpdatedAt   time.Time // set on updates.
}

func marshalMemo(memo map[string]string) (datatypes.JSON, error) {
	if memo == nil {
		return nil, nil
	}

	bs, err := json.Marshal(memo)
	if err != nil {
		return nil, fmt.Errorf("memo marshal: %w", err)
	}

	return datatypes.JSON(bs), nil
}

func unmarshalMemo(f datatypes.JSON) (map[string]string, error) {
	if len(f) == 0 {
		return nil, nil
	}

	var m map[string]string
	if err := json.Unmarshal(f, &m); err != nil {
		return nil, fmt.Errorf("memo unmarshal: %w", err)
	}

	return m, nil
}

func unmarshalPredecls(f datatypes.JSON) (map[string]*apivalues.Value, error) {
	if len(f) == 0 {
		return nil, nil
	}

	var m map[string]*apivalues.Value
	if err := json.Unmarshal(f, &m); err != nil {
		return nil, fmt.Errorf("predecls unmarshal: %w", err)
	}

	return m, nil
}

func marshalPredecls(p map[string]*apivalues.Value) (datatypes.JSON, error) {
	if p == nil {
		return nil, nil
	}

	bs, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("predecls marshal: %w", err)
	}

	return datatypes.JSON(bs), nil
}

func marshalPlugins(pls []*apiproject.ProjectPlugin) (datatypes.JSON, error) {
	if pls == nil {
		return nil, nil
	}

	bs, err := json.Marshal(pls)
	if err != nil {
		return nil, fmt.Errorf("plugins marshal: %w", err)
	}

	return datatypes.JSON(bs), nil
}

func unmarshalPlugins(f datatypes.JSON) ([]*apiproject.ProjectPlugin, error) {
	if len(f) == 0 {
		return nil, nil
	}

	var pls []*apiproject.ProjectPlugin
	if err := json.Unmarshal(f, &pls); err != nil {
		return nil, fmt.Errorf("plugins unmarshal: %w", err)
	}

	return pls, nil
}

func decodeProject(p *project) (*apiproject.Project, error) {
	memo, err := unmarshalMemo(p.Memo)
	if err != nil {
		return nil, err
	}

	predecls, err := unmarshalPredecls(p.Predecls)
	if err != nil {
		return nil, err
	}

	plugins, err := unmarshalPlugins(p.Plugins)
	if err != nil {
		return nil, err
	}

	mainPath, err := apiprogram.ParsePathString(p.MainPath)
	if err != nil {
		return nil, fmt.Errorf("invalid project main path: %w", err)
	}

	d := (&apiproject.ProjectSettings{}).
		SetName(p.Name).
		SetMemo(memo).
		SetMainPath(mainPath).
		SetPredecls(predecls).
		SetEnabled(p.Enabled).
		SetPlugins(plugins)

	var updatedAt *time.Time

	if !p.UpdatedAt.IsZero() {
		updatedAt = &p.UpdatedAt
	}

	aa, err := apiproject.NewProject(
		apiproject.ProjectID(p.ID),
		apiaccount.AccountName(p.AccountName),
		d,
		p.CreatedAt,
		updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid project record: %w", err)
	}

	return aa, nil
}
