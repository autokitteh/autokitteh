package authz

import rego.v1

#
# Helpers
#

authn if input.user_id

user_is_owner := input.is_owner

user_is_filter_owner if {
	authn
	input.data.filter.owner_id == input.user_id
}

#
# Base
#

default allow := false

# Users can do any read operation they like to objects they own.
allow if {
	authn
	input.action_type == "read"
	user_is_owner
}

#
# Builds
#

# allow to save a build only if the user is the owner of both the build and assoicated project:

# - user is the owner of the build and the project.
allow if {
	authn
	input.kind == "bld"
	input.action == "save"
	input.data.build.owner_id == input.user_id
}

# - user is the owner of the build and there is no associated project.
allow if {
	authn
	input.kind == "bld"
	input.action == "save"
	input.data.build.owner_id in input.member_of_org_ids
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

allow if {
	authn
	input.kind == "prj"
	input.action == "resolve"
	input.data.owner_id == input.user_id
}

# allow to list only the authn user's project.
allow if {
	authn
	input.kind == "prj"
	input.action == "list"
	user_is_filter_owner
}
