package export

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var Cmd = common.StandardCommand(&cobra.Command{
	Use:     "export <project name or ID>",
	Short:   "Export project",
	Aliases: []string{"ex"},
	Args:    cobra.ExactArgs(1),

	RunE: export,
})

func export(cmd *cobra.Command, args []string) error {
	r := resolver.Resolver{Client: common.Client()}
	ctx, cancel := common.LimitedContext()
	defer cancel()

	name := args[0]
	prj, _, err := r.ProjectNameOrID(ctx, name)
	if err != nil {
		return err
	}

	if !prj.IsValid() {
		return fmt.Errorf("project %q not found", name)
	}

	tf := sdkservices.ListTriggersFilter{
		ProjectID: prj.ID(),
	}
	triggers, err := r.Client.Triggers().List(ctx, tf)
	if err != nil {
		return err
	}

	// We print to buffer until the end so in case of errors we won't print anything but the error
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "version: v1")
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "project:")
	fmt.Fprintf(&buf, "  name: %s\n", prj.Name().String())

	cf := sdkservices.ListConnectionsFilter{
		ProjectID: prj.ID(),
	}
	conns, err := r.Client.Connections().List(ctx, cf)
	if err != nil {
		return err
	}

	if len(conns) > 0 {
		fmt.Fprintln(&buf, "  connections:")
		for _, c := range conns {
			fmt.Fprintf(&buf, "    - name: %s\n", yamlize(c.Name().String()))
			integ, err := r.Client.Integrations().GetByID(ctx, c.IntegrationID())
			if err != nil {
				return err
			}
			fmt.Fprintf(&buf, "      integration: %s\n", yamlize(integ.UniqueName().String()))
		}
	}

	fmt.Fprintln(&buf, "  triggers:")

	for _, t := range triggers {
		fmt.Fprintf(&buf, "    - name: %s\n", t.Name().String())
		fmt.Fprintf(&buf, "      call: %s\n", t.CodeLocation().CanonicalString())
		if filter := t.Filter(); filter != "" {
			fmt.Fprintf(&buf, "      filter: %s\n", filter)
		}
		if etype := t.EventType(); etype != "" {
			fmt.Fprintf(&buf, "      event_type: %s\n", etype)
		}

		switch t.SourceType() {
		case sdktypes.TriggerSourceTypeWebhook:
			fmt.Fprintln(&buf, "      webhook: {}")
			// TODO: More types
		case sdktypes.TriggerSourceTypeSchedule:
			fmt.Fprintf(&buf, "      schedule: %s\n", yamlize(t.Schedule()))
		case sdktypes.TriggerSourceTypeConnection:
			conn, err := r.Client.Connections().Get(ctx, t.ConnectionID())
			if err != nil {
				return err
			}
			fmt.Fprintf(&buf, "      connection: %s\n", yamlize(conn.Name().String()))
		}
	}

	envs, err := r.Client.Envs().List(ctx, prj.ID())
	if err != nil {
		return err
	}

	// Collect vars first, print only if there are some
	varsMap := make(map[string]string)
	for _, env := range envs {
		sid, err := sdktypes.Strict(sdktypes.ParseVarScopeID(env.ID().String()))
		if err != nil {
			return err
		}
		vars, err := r.Client.Vars().Get(ctx, sid)
		if err != nil {
			return err
		}

		for _, v := range vars {
			if v.IsSecret() {
				continue
			}
			varsMap[v.Name().String()] = v.Value()
		}
	}

	if len(varsMap) > 0 {
		fmt.Fprintln(&buf, "  vars:")
		for n, v := range varsMap {
			fmt.Fprintf(&buf, "    - name: %s\n", n)
			fmt.Fprintf(&buf, "      value: %s\n", yamlize(v))
		}
	}

	fmt.Println(buf.String())
	return nil
}

func yamlize(v string) string {
	data, _ := yaml.Marshal(v)
	// Trim newline added by yaml.Marshal
	return string(data[:len(data)-1])
}
