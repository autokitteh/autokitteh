package authz

import rego.v1

use_authn_for_default_list_filter_owner := false

# is user authenticated?
authn if input.user_id != ""

default allow := false

# allow all operations except creates. these needs to be validated as
# registered as created by their respective owners.
allow if {
	authn
	input.action_type in ["read", "delete", "write"]
}

#
# Projects
#

# allow to create and update a project only if the user is the owner.
allow if {
	authn
	input.kind == "prj"
	input.action in ["create", "update"]
	input.data.project.owner_id == input.user_id
}

#
# Builds
#

# allow to save a build only if the user is the owner.
allow if {
	authn
	input.kind == "bld"
	input.action == "save"
	input.data.build.owner_id == input.user_id
}
