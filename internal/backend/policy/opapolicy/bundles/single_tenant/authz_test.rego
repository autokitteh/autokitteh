package authz_test

import rego.v1

import data.authz

test_if_unauthn_allowed if {
	not authz.allow with input as {}
}

test_if_authn_allowed if {
	authz.allow with input as {"user_id": "meow"}
}
