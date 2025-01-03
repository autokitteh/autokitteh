package authz

import rego.v1

#
# Helpers
#

authn if input.user_id

is_org_member_of(org_id) := org_id in input.user_org_ids

is_org_member_of_resource := is_org_member_of(input.resource_org_id)

# There is only a single unique org id associated, and the user is a member of it.
member_of_single_assosicated_org_id if {
	count(input.associated_org_ids) == 1
	is_org_member_of(input.associated_org_ids[0])
}

single_associated_project_id(pid) if {
	count(input.associated_project_ids) == 1
	input.associated_project_ids[0] == pid
}

#
# Base
#

default allow := false

# Users can do any read operation they like to objects they own.
allow if {
	authn
	input.action_type == "read"
	is_org_member_of_resource
}

#
# Users
#

allow if {
	authn
	input.kind == "usr"
	input.action_type == "read"
	input.user_id == input.resource_id
}

# Allow any user to invite other users as long as they
# specify only the email.
allow if {
	authn
	input.action == "create"
	input.data.status == "INVITED"
	not input.data.user.display_name
	not input.data.user.default_org_id
}

# do not allow users to change their own status.
allow if {
	authn
	input.action == "update"
	not "status" in input.data.field_mask
	input.data.status == "UNSPECIFIED"
}

#
# Orgs
#

allow if {
	authn
	input.kind == "org"
	is_org_member_of_resource
}

allow if {
	authn
	input.kind == "org"
	input.action == "create"
}

#
# Projects
#

allow if {
	authn
	input.kind == "prj"
	input.action == "create"
	member_of_single_assosicated_org_id
}

allow if {
	authn
	input.kind == "prj"
	input.action in ["set-resources", "build", "delete", "update"]
	is_org_member_of_resource
}

allow if {
	authn
	input.kind == "prj"
	input.action == "list"
	is_org_member_of(input.data.filter.org_id)
}

#
# Builds
#

allow if {
	authn
	input.kind == "bld"
	input.action == "save"
	member_of_single_assosicated_org_id
}

allow if {
	authn
	input.kind == "bld"
	input.action == "list"
	member_of_single_assosicated_org_id
}

allow if {
	authn
	input.kind == "bld"
	input.action_type == "delete"
	is_org_member_of_resource
}

#
# Integrations
#

allow if {
	authn
	input.kind == "int"
	input.action in ["get", "list"]
}

#
# Connections
#

allow if {
	authn
	input.kind == "con"
	input.action == "create"
	member_of_single_assosicated_org_id
}

allow if {
	authn
	input.kind == "con"
	input.action in ["delete", "test", "refresh", "update"]
	is_org_member_of_resource
}

allow if {
	authn
	input.kind == "con"
	input.action == "list"
	member_of_single_assosicated_org_id
}

#
# Deployments
#

allow if {
	authn
	input.kind == "dep"
	input.action == "create"
	member_of_single_assosicated_org_id
	single_associated_project_id(input.data.deployment.project_id)
}

allow if {
	authn
	input.kind == "dep"
	input.action in ["activate", "deactivate", "delete", "test"]
	is_org_member_of_resource
}

allow if {
	authn
	input.kind == "dep"
	input.action == "list"
	member_of_single_assosicated_org_id
}

#
# Triggers
#

allow if {
	authn
	input.kind == "trg"
	input.action == "create"
	member_of_single_assosicated_org_id
}

allow if {
	authn
	input.kind == "trg"
	input.action == "list"
	member_of_single_assosicated_org_id
}

allow if {
	authn
	input.kind == "trg"
	input.action in ["delete", "update"]
	is_org_member_of_resource
}

#
# Events
#
# saving is forbidden by default.

allow if {
	authn
	input.kind == "evt"
	input.action == "list"
	member_of_single_assosicated_org_id
}

#
# Sessions
#

allow if {
	authn
	input.kind == "ses"
	input.action == "start"
	member_of_single_assosicated_org_id
}

allow if {
	authn
	input.kind == "ses"
	input.action in ["stop", "delete"]
	is_org_member_of_resource
}

allow if {
	authn
	input.kind == "ses"
	input.action == "list"
	member_of_single_assosicated_org_id
}

#
# Vars
#

allow if {
	authn
	input.kind in ["prj", "con"]
	input.action in ["set-var", "delete-var", "delete-all-vars"]
	is_org_member_of_resource
}

#
# Dispatcher
#
# everything is forbidden by default.
