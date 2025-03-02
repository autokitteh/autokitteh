package connection

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	orgIDVar               = sdktypes.NewSymbol("org_id")
	privateClientIDVar     = sdktypes.NewSymbol("private_client_id")
	privateClientSecretVar = sdktypes.NewSymbol("private_client_secret")
	privateTenantIDVar     = sdktypes.NewSymbol("private_tenant_id")
)

// OrgInfo contains basic details about a Microsoft organization
// (based on: https://learn.microsoft.com/en-us/graph/api/organization-get).
// "VerifiedDomains" isn't included because it's an array, but it's available if needed.
type OrgInfo struct {
	ID          string `json:"id" var:"org_id"`
	DisplayName string `json:"displayName" var:"org_display_name"`
	TenantType  string `json:"tenantType" var:"org_tenant_type"`
}

// PrivateApp contains the user-provided details of
// a private Microsoft OAuth 2.0 app or daemon app.
type PrivateApp struct {
	ClientID     string `var:"private_client_id"`
	ClientSecret string `var:"private_client_secret,secret"`
	Certificate  string `var:"private_certificate,secret"`
	TenantID     string `var:"private_tenant_id"`
}

// UserInfo contains user profile details from Microsoft Graph
// (based on: https://learn.microsoft.com/en-us/graph/api/user-get).
type UserInfo struct {
	PrincipalName string `json:"userPrincipalName" var:"user_principal_name"`
	ID            string `json:"id" var:"user_id"`
	DisplayName   string `json:"displayName" var:"user_display_name"`
	Surname       string `json:"surname" var:"user_surname"`
	GivenName     string `json:"givenName" var:"user_given_name"`
	Language      string `json:"preferredLanguage" var:"user_language"`
	Mail          string `json:"mail" var:"user_mail"`
	MobilePhone   string `json:"mobilePhone" var:"user_mobile_phone"`
}
