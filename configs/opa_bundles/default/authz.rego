package authz

import rego.v1

#
# Helpers
#

has_active_role_in_org(user, org_id, "") if {
	m := user.org_memberships[org_id]
	m.status == "ACTIVE"
}

has_active_role_in_org(user, org_id, role) if {
	m := user.org_memberships[org_id]
	m.status == "ACTIVE"
	role in m.roles
}

is_active_org_member_of(org_id) if has_active_role_in_org(input.authn_user, org_id, "")

is_org_admin(org_id) if has_active_role_in_org(input.authn_user, org_id, "admin")

is_active_member_of_resource_org := is_active_org_member_of(input.resource.org_id)

is_resource_org_admin := is_org_admin(input.resource.org_id)

# There is only a single unique org id associated, and the user is a member of it.
is_active_member_of_single_assosicated_org_id if {
	# build a set of all specified org ids.
	oids := {oid | oid := input.associations[_].org_id}

	# there is only a single item in the set.
	count(oids) == 1

	every oid in oids { is_active_org_member_of(oid) }
}

#
# Base
#

default allow := false

# Users can do any read operation they like to objects they are a member of their org.
allow if {
	input.action.type == "read"
	is_active_member_of_resource_org
}

#
# Users
#

# Users can read their own user object.
allow if {
	input.resource.kind == "usr"
	input.action.type == "read"
	input.authn_user.id == input.resource.id
}

# Users can update anything of their own user object except status.
allow if {
	input.resource.kind == "usr"
	input.action.name == "update"
	input.authn_user.id == input.resource.id
	not "status" in input.data.field_mask
	input.data.status == "UNSPECIFIED"
}

# Allow any user to invite other users as long as they
# specify only the email.
allow if {
	input.resource.kind == "usr"
	input.action.name == "create"
	input.data.status == "INVITED"
	not input.data.user.display_name
	not input.data.user.default_org_id
}

# Anyone can translate an email to an id.
allow if {
	input.resource.kind == "usr"
	input.action.name = "get-id"
}

# A user can get any another user.
allow if {
	input.resource.kind == "usr"
	input.action.name == "get"
}

#
# Orgs
#

# Anyone who is either invited or active in an org can get the org.
allow if {
	input.resource.kind == "org"
	input.action.name == "get"
	input.authn_user.org_memberships[input.resource.id].status in ["ACTIVE", "INVITED"]
}

# Org admins can perform delete, updates on it and remove members.
allow if {
	input.resource.kind == "org"
	input.action.name in ["delete", "update", "remove-member"]
	is_resource_org_admin
}

# Anyone can create an org.
allow if {
	input.resource.kind == "org"
	input.action.name == "create"
}

# New members must be invited.
allow if {
	input.resource.kind == "org"
	input.action.name == "add-member"
	input.data.org_member.status == "ORG_MEMBER_STATUS_INVITED"
	not input.data.org_member.roles
	is_resource_org_admin
}

# Org admins can perform delete, updates on it and remove members.
allow if {
	input.kind == "org"
	input.action in ["delete", "update", "remove-member"]
	is_resource_org_admin
}

# Anyone can create an org.
allow if {
	input.resource.kind == "org"
	input.action.name == "add-member"
	input.data.org_member.status == "ORG_MEMBER_STATUS_INVITED"
	is_resource_org_admin
}

# Only the invited user can accept or reject the invitation,
# and must not change its roles.
allow if {
	input.resource.kind == "org"
	input.action.name == "update-member"
	input.authn_user.id == input.associations.user.id
	input.data.current_status == "INVITED"
	input.data.new_status in ["ACTIVE", "DECLINED"]
	input.data.field_mask == ["status"]
}

# Org admin can update any member, as long as they don't change the status.
allow if {
	input.resource.kind == "org"
	input.action.name == "update-member"
	is_resource_org_admin
	not "status" in input.data.field_mask
	input.data.new_status == "UNSPECIFIED"
}

# Users of a specific org can see all other users who are active
# at that org.
allow if {
	input.resource.kind == "org"
	input.action.name == "get-member"
	input.data.member_status == "ACTIVE"
	is_active_org_member_of(input.resource.id)
}

# Org admins can see any org member regardless of status.
allow if {
	input.resource.kind == "org"
	input.action.name == "get-member"
	is_resource_org_admin
}

#
# Projects
#

allow if {
	input.resource.kind == "prj"
	input.action.name == "create"
	is_active_org_member_of(input.data.project.org_id)
}

allow if {
	input.resource.kind == "prj"
	input.action.name in ["set-resources", "build", "delete", "update"]
	is_active_member_of_resource_org
}

allow if {
	input.resource.kind == "prj"
	input.action.name == "list"
	is_active_org_member_of(input.data.filter.org_id)
}

#
# Builds
#

allow if {
	input.resource.kind == "bld"
	input.action.name == "save"
	is_active_member_of_single_assosicated_org_id
}

allow if {
	input.resource.kind == "bld"
	input.action.name == "list"
	is_active_member_of_single_assosicated_org_id
}

allow if {
	input.resource.kind == "bld"
	input.action.type == "delete"
	is_active_member_of_resource_org
}

#
# Integrations
#

allow if {
	input.resource.kind == "int"
	input.action.name in ["get", "list"]
}

#
# Connections
#

allow if {
	input.resource.kind == "con"
	input.action.name == "create"
	is_active_org_member_of(input.associations.project.org_id)
}

allow if {
	input.resource.kind == "con"
	input.action.name in ["delete", "test", "refresh", "update"]
	is_active_member_of_resource_org
}

allow if {
	input.resource.kind == "con"
	input.action.name == "list"
	is_active_member_of_single_assosicated_org_id
}

#
# Deployments
#

allow if {
	input.resource.kind == "dep"
	input.action.name == "create"
	is_active_member_of_single_assosicated_org_id
}

allow if {
	input.resource.kind == "dep"
	input.action.name in ["activate", "deactivate", "delete", "test"]
	is_active_member_of_resource_org
}

allow if {
	input.resource.kind == "dep"
	input.action.name == "list"
	is_active_member_of_single_assosicated_org_id
}

#
# Triggers
#

allow if {
	input.resource.kind == "trg"
	input.action.name == "create"
	is_active_member_of_single_assosicated_org_id
}

allow if {
	input.resource.kind == "trg"
	input.action.name == "list"
	is_active_member_of_single_assosicated_org_id
}

allow if {
	input.resource.kind == "trg"
	input.action.name in ["delete", "update"]
	is_active_member_of_resource_org
}

#
# Events
#
# saving is forbidden by default.

allow if {
	input.resource.kind == "evt"
	input.action.name == "list"
	is_active_member_of_single_assosicated_org_id
}

#
# Sessions
#

allow if {
	input.resource.kind == "ses"
	input.action.name == "start"
	is_active_member_of_single_assosicated_org_id
}

allow if {
	input.resource.kind == "ses"
	input.action.name in ["stop", "delete"]
	is_active_member_of_resource_org
}

allow if {
	input.resource.kind == "ses"
	input.action.name == "list"
	is_active_member_of_single_assosicated_org_id
}

#
# Vars
#

allow if {
	input.resource.kind in ["prj", "con"]
	input.action.name in ["set-var", "delete-var", "delete-all-vars"]
	is_active_member_of_resource_org
}

#
# Dispatcher
#
# everything is forbidden by default.
