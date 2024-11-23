package authz_test

import rego.v1

import data.authz

user_id := "usr_test"

other_user_id := "usr_other"

test_unauthn_allowed if {
	not authz.allow with input as {}
}

test_authn_same_user_allowed if {
	authz.allow with input as {"user_id": user_id, "owner_id": user_id}
}

test_authn_different_user_not_allowed if {
	not authz.allow with input as {"action": "whatever", "user_id": user_id, "owner_id": other_user_id}
}
