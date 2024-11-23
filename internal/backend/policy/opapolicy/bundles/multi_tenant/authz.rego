package authz

import rego.v1

use_authn_for_default_list_filter_owner := true

#
# Helpers
#

authn if input.user_id

has_owner if input.owner_id

user_is_owner if {
	has_owner
	input.user_id == input.owner_id
}

has_project_owner if input.data.project_owner_id

user_is_project_owner if {
	authn
	has_project_owner
	input.data.project_owner_id == input.user_id
}

#
# Base
#

default allow := false

# Users can do anything they like to objects they own.
allow if {
	authn
	user_is_owner
}

#
# Users
#

allow if {
	authn
	input.kind == "usr"
	input.action == "get"
}

#
# Builds
#

allow if {
	authn
	input.kind == "bld"
	input.action == "list"
	input.data.filter.owner_id == input.user_id
}

allow if {
	authn
	input.kind == "bld"
	input.action == "save"
	user_is_project_owner
}

allow if {
	authn
	input.kind == "bld"
	input.action == "save"
	not has_project_owner
}

#
# Projects
#

allow if {
	authn
	input.kind == "prj"
	input.action == "create"
}
