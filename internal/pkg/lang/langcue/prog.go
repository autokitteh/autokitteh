package langcue

import (
	"encoding/json"
	"fmt"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
)

type progModule struct {
	Path    string                 `json:"path"`
	Name    string                 `json:"name"`
	Sources map[string]string      `json:"sources"`
	Context map[string]interface{} `json:"context"`
}

type compiledModule struct {
	Path      *apiprogram.Path            `json:"path"`
	ValueName string                      `json:"value"`
	Sources   map[string]string           `json:"sources"`
	Context   map[string]*apivalues.Value `json:"context"`
}

type prog struct {
	Consts  map[string]interface{} `json:"consts"`
	Modules []progModule           `json:"modules"`
}

type compiled struct {
	Consts  map[string]*apivalues.Value
	Modules []*compiledModule
}

func (c *compiled) encode() ([]byte, error) {
	return json.Marshal(c)
}

func (c *compiled) decode(src []byte) error {
	*c = compiled{}
	return json.Unmarshal(src, c)
}

// TODO: more validation!
func (p *prog) compile() (c *compiled, err error) {
	c = &compiled{
		Modules: make([]*compiledModule, len(p.Modules)),
		Consts:  make(map[string]*apivalues.Value, len(p.Consts)),
	}

	if err = apivalues.WrapValuesMap(c.Consts, p.Consts); err != nil {
		err = fmt.Errorf("values: %w", err)
		return
	}

	for i, pp := range p.Modules {
		if c.Modules[i], err = pp.compile(); err != nil {
			err = fmt.Errorf("setup #%d: %w", i, err)
			return
		}
	}

	return
}

func (s *progModule) compile() (c *compiledModule, err error) {
	c = &compiledModule{
		ValueName: s.Name,
		Sources:   s.Sources,
		Context:   make(map[string]*apivalues.Value, len(s.Context)),
	}

	if c.Path, err = apiprogram.ParsePathString(s.Path); err != nil {
		err = fmt.Errorf("load path: %w", err)
		return
	}

	if err = apivalues.WrapValuesMap(c.Context, s.Context); err != nil {
		err = fmt.Errorf("context: %w", err)
		return
	}

	return
}
