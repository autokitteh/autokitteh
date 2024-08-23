package kittehs

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterpolate(t *testing.T) {
	tests := []struct {
		in, out   string
		err, verr bool
	}{
		{},
		{in: `hello`, out: `hello`},
		{in: `{{x,}} y`, out: `X, y`},
		{in: `x, {{y}}`, out: `x, Y`},
		{in: `x, {{y}} `, out: `x, Y `},
		{in: `hello, \{{world}}`, out: `hello, {{world}}`},
		{in: `{{hello, {{ world }}}}`, out: `HELLO, {{ WORLD }}`},
		{in: `{\{hello, {{world}} }}`, out: `{{hello, WORLD }}`},
		{in: `{{hello}}, {{world}}`, out: `HELLO, WORLD`},
		{in: `{{`, err: true},
		{in: `{{ meow`, err: true},
		{in: `hello }}`, out: `hello }}`},
		{in: `{{ hello }`, err: true},
		{in: `\`, err: true},
		{in: `x\`, err: true},
		{in: `{{}}`, out: ``},
		{in: `1{{}}2`, out: `12`},
		{in: "meow {{!woof}}", out: "meow !WOOF", verr: true},
	}

	i := Interpolator{
		Left:  "{{",
		Right: "}}",
		EvaluateExpr: func(expr string) (string, error) {
			return strings.ToUpper(expr), nil
		},
		ValidateExpr: func(in string) error {
			if strings.HasPrefix(in, "!") {
				return fmt.Errorf("invalid expression: %s", in)
			}

			return nil
		},
	}

	for j, test := range tests {
		t.Run(fmt.Sprintf("%d", j), func(t *testing.T) {
			out, err := i.Execute(test.in)
			if test.err {
				assert.Error(t, err)
			} else if assert.NoError(t, err) {
				assert.Equal(t, test.out, out)
			}

			err = i.Validate(test.in)
			if test.verr || test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
