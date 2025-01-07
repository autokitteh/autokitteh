package orgs

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var orgsCmd = common.StandardCommand(&cobra.Command{
	Use:   "orgs",
	Short: "Orgs: create, get, add-member, list-members, get-member, remove-member, update-member",
	Args:  cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(orgsCmd)
}

func init() {
	// Subcommands.
	orgsCmd.AddCommand(addMemberCmd)
	orgsCmd.AddCommand(createCmd)
	orgsCmd.AddCommand(deleteCmd)
	orgsCmd.AddCommand(getCmd)
	orgsCmd.AddCommand(getMemberCmd)
	orgsCmd.AddCommand(listMembersCmd)
	orgsCmd.AddCommand(removeMemberCmd)
	orgsCmd.AddCommand(updateCmd)
	orgsCmd.AddCommand(updateMemberCmd)
}

func orgs() sdkservices.Orgs { return common.Client().Orgs() }
