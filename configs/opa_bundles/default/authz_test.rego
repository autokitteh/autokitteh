package authz_test

import rego.v1

# for now just verify it compiles.
import data.authz

test_has_active_role_in_org if {
	user := {"org_memberships": {
		"o1": {"status": "ACTIVE", "roles": ["admin"]},
		"o2": {"status": "INVITED", "roles": ["admin"]},
	}}

	authz.has_active_role_in_org(user, "o1", "admin")
	not authz.has_active_role_in_org(user, "o1", "meow")
	authz.has_active_role_in_org(user, "o1", "")

	not authz.has_active_role_in_org(user, "o2", "")
	not authz.has_active_role_in_org(user, "o2", "admin")

	not authz.has_active_role_in_org(user, "o3", "")
	not authz.has_active_role_in_org(user, "o3", "admin")
}

test_is_active_member_of_single_assosicated_org_id if {
	# single org.
	authz.is_active_member_of_single_assosicated_org_id with input as {
		"authn_user": {"org_memberships": {"o1": {"status": "ACTIVE"}}},
		"associations": {
			"project": {"org_id": "o1"},
			"org": {"org_id": "o1"},
		},
	}

	# two different orgs.
	not authz.is_active_member_of_single_assosicated_org_id with input as {
		"authn_user": {"org_memberships": {"o1": {"status": "ACTIVE"}, "o2": {"status": "ACTIVE"}}},
		"associations": {
			"project": {"org_id": "o2"},
			"org": {"org_id": "o1"},
		},
	}

	# on association without an org, should be ignored.
	authz.is_active_member_of_single_assosicated_org_id with input as {
		"authn_user": {"org_memberships": {"o1": {"status": "ACTIVE"}}},
		"associations": {
			"project": {"org_id": "o1"},
			"org": {"org_id": "o1"},
			"something": {},
		},
	}

	# not a member of an org.
	not authz.is_active_member_of_single_assosicated_org_id with input as {
		"authn_user": {"org_memberships": {"o1": {"status": "ACTIVE"}}},
		"associations": {
			"project": {"org_id": "o1"},
			"org": {"org_id": "o1"},
			"something": {"org_id": "o2"},
		},
	}

	# member of an inactive org.
	not authz.is_active_member_of_single_assosicated_org_id with input as {
		"authn_user": {"org_memberships": {"o1": {"status": "INVITED"}}},
		"associations": {"project": {"org_id": "o1"}},
	}

	# no associations.
	not authz.is_active_member_of_single_assosicated_org_id with input as {
		"authn_user": {"org_memberships": {"o1": {"status": "INVITED"}}},
		"associations": {},
	}

	# only empty associations.
	not authz.is_active_member_of_single_assosicated_org_id with input as {
		"authn_user": {"org_memberships": {"o1": {"status": "INVITED"}}},
		"associations": {"empty": {}},
	}
}
