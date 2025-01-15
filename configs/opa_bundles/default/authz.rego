package authz

import rego.v1

#
# Helpers
#

is_active_org_member_of(org_id) if input.authn_user_orgs[org_id].status == "ACTIVE"

has_active_role_in_org(org_id, role) if {
	org := input.authn_user_orgs[org_id]
	org.status == "ACTIVE"
	role in org.roles
}

is_org_admin(org_id) if has_active_role_in_org(org_id, "admin")

is_active_member_of_resource_org := is_active_org_member_of(input.resource_org_id)

is_resource_org_admin := is_org_admin(input.resource_org_id)

# There is only a single unique org id associated, and the user is a member of it.
member_of_single_assosicated_org_id if {
	count(input.associated_org_ids) == 1
	is_active_org_member_of(input.associated_org_ids[0])
}

single_associated_project_id(pid) if {
	count(input.associated_project_ids) == 1
	input.associated_project_ids[0] == pid
}

#
# Base
#

default allow := false

# Users can do any read operation they like to objects they are a member of their org.
allow if {
	input.action_type == "read"
	is_active_member_of_resource_org
}

#
# Users
#

# A user can read its own user object.
allow if {
	input.kind == "usr"
	input.action_type == "read"
	input.authn_user_id == input.resource_id
}

# A user can update anything of its own user object except status.
allow if {
	input.kind == "usr"
	input.action == "update"
	input.authn_user_id == input.resource_id
	not "status" in input.data.field_mask
	input.data.status == "UNSPECIFIED"
}

# Allow any user to invite other users as long as they
# specify only the email.
allow if {
	input.kind == "usr"
	input.action == "create"
	input.data.status == "INVITED"
	not input.data.user.display_name
	not input.data.user.default_org_id
}

# Anyone can translate an email to an id.
allow if {
	input.kind == "usr"
	input.action = "get-id"
}

#
# Orgs
#

# Anyone who is either invited or active in an org can get the org.
allow if {
	input.kind == "org"
	input.action == "get"
	input.authn_user_orgs[input.resource_org_id].status in ["ACTIVE", "INVITED"]
}

# Org admins can perform delete, updates on it and remove members.
allow if {
	input.kind == "org"
	input.action in ["delete", "update", "remove-member"]
	is_resource_org_admin
}

# Anyone can create an org.
allow if {
	input.kind == "org"
	input.action == "create"
}

# New members must be invited.
allow if {
	input.kind == "org"
	input.action == "add-member"
	is_resource_org_admin
	input.data.org_member.status == "ORG_MEMBER_STATUS_INVITED"
}

# Only the invited user can accept or reject the invitation,
# and must not change its roles.
allow if {
	input.kind == "org"
	input.action == "update-member"
	input.authn_user_id == input.associations.user.id
	input.data.current_status == "INVITED"
	input.data.new_status in ["ACTIVE", "DECLINED"]
	input.data.field_mask == ["status"]
}

# Org admin can update any member, as long as they don't change the status.
allow if {
	input.kind == "org"
	input.action == "update-member"
	is_resource_org_admin
	not "status" in input.data.field_mask
	input.data.new_status == "UNSPECIFIED"
}

# Users of a specific org can see all other users who are active
# at that org.
allow if {
	input.kind == "org"
	input.action == "get-member"
	input.data.member_status == "ACTIVE"
	is_active_org_member_of(input.resource_id)
}

# Org admins can see any org member regardless of status.
allow if {
	input.kind == "org"
	input.action == "get-member"
	is_resource_org_admin
}

#
# Projects
#

allow if {
	input.kind == "prj"
	input.action == "create"
	member_of_single_assosicated_org_id
}

allow if {
	input.kind == "prj"
	input.action in ["set-resources", "build", "delete", "update"]
	is_active_member_of_resource_org
}

allow if {
	input.kind == "prj"
	input.action == "list"
	is_active_org_member_of(input.data.filter.org_id)
}

#
# Builds
#

allow if {
	input.kind == "bld"
	input.action == "save"
	member_of_single_assosicated_org_id
}

allow if {
	input.kind == "bld"
	input.action == "list"
	member_of_single_assosicated_org_id
}

allow if {
	input.kind == "bld"
	input.action_type == "delete"
	is_active_member_of_resource_org
}

#
# Integrations
#

allow if {
	input.kind == "int"
	input.action in ["get", "list"]
}

#
# Connections
#

allow if {
	input.kind == "con"
	input.action == "create"
	member_of_single_assosicated_org_id
}

allow if {
	input.kind == "con"
	input.action in ["delete", "test", "refresh", "update"]
	is_active_member_of_resource_org
}

allow if {
	input.kind == "con"
	input.action == "list"
	member_of_single_assosicated_org_id
}

#
# Deployments
#

allow if {
	input.kind == "dep"
	input.action == "create"
	member_of_single_assosicated_org_id
	single_associated_project_id(input.data.deployment.project_id)
}

allow if {
	input.kind == "dep"
	input.action in ["activate", "deactivate", "delete", "test"]
	is_active_member_of_resource_org
}

allow if {
	input.kind == "dep"
	input.action == "list"
	member_of_single_assosicated_org_id
}

#
# Triggers
#

allow if {
	input.kind == "trg"
	input.action == "create"
	member_of_single_assosicated_org_id
}

allow if {
	input.kind == "trg"
	input.action == "list"
	member_of_single_assosicated_org_id
}

allow if {
	input.kind == "trg"
	input.action in ["delete", "update"]
	is_active_member_of_resource_org
}

#
# Events
#
# saving is forbidden by default.

allow if {
	input.kind == "evt"
	input.action == "list"
	member_of_single_assosicated_org_id
}

#
# Sessions
#

allow if {
	input.kind == "ses"
	input.action == "start"
	member_of_single_assosicated_org_id
}

allow if {
	input.kind == "ses"
	input.action in ["stop", "delete"]
	is_active_member_of_resource_org
}

allow if {
	input.kind == "ses"
	input.action == "list"
	member_of_single_assosicated_org_id
}

#
# Vars
#

allow if {
	input.kind in ["prj", "con"]
	input.action in ["set-var", "delete-var", "delete-all-vars"]
	is_active_member_of_resource_org
}

#
# Dispatcher
#
# everything is forbidden by default.
