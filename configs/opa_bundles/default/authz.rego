package authz

import rego.v1

#
# Helpers
#

authn if input.user_id

is_member(org_id) := org_id in input.user_org_ids

#
# Base
#

default allow := false

# Users can do any read operation they like to objects they own.
allow if {
	authn
	input.action_type == "read"
	is_member(input.resource_org_id)
}

#
# Projects
#

allow if {
	authn
	input.kind == "prj"
	input.action == "create"
	is_member(input.data.project.org_id)
}

allow if {
	authn
	input.kind == "prj"
	input.action in ["set-resources", "build"]
	is_member(input.resource_org_id)
}

allow if {
	authn
	input.kind == "prj"
	input.action == "list"
	is_member(input.data.filter.org_id)
}

allow if {
	authn
	input.kind == "prj"
	input.action == "delete"
	is_member(input.resource_org_id)
}

#
# Builds
#

allow if {
	authn
	input.kind == "bld"
	input.action == "save"
	is_member(input.data.build.org_id)

	# allow to create builds only for projects in the same org.
	input.data.build.org_id == input.associated_project_org_id
}
