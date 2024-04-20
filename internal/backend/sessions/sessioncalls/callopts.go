package sessioncalls

import (
	"fmt"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/akmodules/ak"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const callOptsArgName = "ak"

type CallOpts struct {
	Catch   bool          `json:"catch"`
	Timeout time.Duration `json:"timeout"`
}

// NOTE: If err is not nil, this might still have valid values in the other return values.
func parseCallSpec(spec sdktypes.SessionCallSpec) (v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value, opts *CallOpts, err error) {
	v, args, kwargs = spec.Data()

	opts = &CallOpts{}

	var found bool

	// First check if the call opts are in regular arguments. If so, they are set as the last
	// argument with a struct type with ctor with symbol `callOptsArgSymbol`.
	if len(args) > 0 {
		last := args[len(args)-1]

		if st := last.GetStruct(); st.IsValid() && st.Ctor().GetSymbol().Symbol() == ak.CallOptsCtorSymbol {
			if err = last.UnwrapInto(&opts); err != nil {
				err = fmt.Errorf("invalid opts: %w", err)
				return
			}

			args = args[:len(args)-1]
			found = true
		}
	}

	// Then check if they are mentioned in kwargs with `callOptsArgName` key.
	vopts := kwargs[callOptsArgName]
	if vopts.IsValid() && !vopts.IsNothing() {
		if found {
			// Already specified in args, gevalt!
			err = fmt.Errorf("call options found in both args and kwargs")
			return
		}

		delete(kwargs, callOptsArgName)

		if err = vopts.UnwrapInto(&opts); err != nil {
			err = fmt.Errorf("invalid opts: %w", err)
			return
		}

		return
	}

	// Each can be overwritten by a key in kwargs that begins with `callOptsArgName` and an underscore.
	for k, v := range kwargs {
		if !strings.HasPrefix(k, callOptsArgName+"_") {
			continue
		}

		delete(kwargs, k)

		switch k[len(callOptsArgName)+1:] {
		case "catch":
			if err = v.UnwrapInto(&opts.Catch); err != nil {
				err = fmt.Errorf("invalid call option catch value: %w", err)
			}
		case "timeout":
			if err = v.UnwrapInto(&opts.Timeout); err != nil {
				err = fmt.Errorf("invalid call option timeout value: %w", err)
			}
		default:
			err = fmt.Errorf("invalid call option: %s", k)
		}

		if err != nil {
			break
		}
	}

	return
}
