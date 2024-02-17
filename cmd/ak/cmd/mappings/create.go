package mappings

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	module string
	events []string
)

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create <--env=...> <--connection=...> <--module=...> <--event=... [--event=... [...]]>",
	Short:   "Create connection mapping",
	Aliases: []string{"c"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		e, eid, err := r.EnvNameOrID(env, "")
		if err != nil {
			return err
		}
		if e == nil {
			err = fmt.Errorf("environment %q not found", env)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		c, cid, err := r.ConnectionNameOrID(connection)
		if err != nil {
			if errors.As(err, resolver.NotFoundErrorType) {
				err = common.NewExitCodeError(common.NotFoundExitCode, err)
			}
			return err
		}
		if c == nil {
			err = fmt.Errorf("connection %q not found", connection)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		_, err = sdktypes.ParseSymbol(module)
		if err != nil {
			return fmt.Errorf("invalid module name %q: %w", module, err)
		}

		es, err := kittehs.TransformError(events, mappingEventToProto)
		if err != nil {
			return fmt.Errorf("invalid event: %w", err)
		}

		m, err := sdktypes.StrictMappingFromProto(&sdktypes.MappingPB{
			Events:       es,
			EnvId:        eid.String(),
			ConnectionId: cid.String(),
			ModuleName:   module,
		})
		if err != nil {
			return fmt.Errorf("invalid mapping: %w", err)
		}

		mid, err := mappings().Create(context.Background(), m)
		if err != nil {
			return fmt.Errorf("create mapping: %w", err)
		}

		common.RenderKVIfV("mapping_id", mid)
		return nil
	},
})

func init() {
	// Command-specific flags.
	createCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	kittehs.Must0(createCmd.MarkFlagRequired("env"))

	createCmd.Flags().StringVarP(&connection, "connection", "n", "", "connection name or ID")
	kittehs.Must0(createCmd.MarkFlagRequired("connection"))

	createCmd.Flags().StringVarP(&module, "module", "m", "", "module name")
	kittehs.Must0(createCmd.MarkFlagRequired("module"))

	createCmd.Flags().StringSliceVarP(&events, "event", "p", nil, `one or more of: "<event_type>@<file_path>:<function_name>"`)
	kittehs.Must0(createCmd.MarkFlagRequired("event"))
}

func mappingEventToProto(e string) (*sdktypes.MappingEventPB, error) {
	eventType, entryPoint, ok := strings.Cut(e, "@")
	if !ok {
		return nil, errors.New(`must separate event type and code location with "@"`)
	}
	if eventType == "" {
		return nil, errors.New("event type cannot be empty")
	}
	if entryPoint == "" {
		return nil, errors.New("entry-point cannot be empty")
	}

	cl, err := sdktypes.ParseCodeLocation(entryPoint)
	if err != nil {
		return nil, fmt.Errorf("invalid entry-point %q: %w", entryPoint, err)
	}

	return &sdktypes.MappingEventPB{
		EventType:    eventType,
		CodeLocation: cl.ToProto(),
	}, nil
}
