package authz

// Authorization operation constants.
// These constants define all the permission actions used throughout the system.

const (
	// Store operations
	OpStoreWriteDo        = "write:do"
	OpStoreReadGet        = "read:get"
	OpStoreReadList       = "read:list"
	OpStoreWritePublish   = "write:publish"
	OpStoreWriteUnpublish = "write:unpublish"

	// Organization operations
	OpOrgCreateCreate       = "create:create"
	OpOrgReadGet            = "read:get"
	OpOrgDeleteDelete       = "delete:delete"
	OpOrgUpdateUpdate       = "update:update"
	OpOrgReadListMembers    = "read:list-members"
	OpOrgWriteAddMember     = "write:add-member"
	OpOrgDeleteRemoveMember = "delete:remove-member"
	OpOrgReadGetMember      = "read:get-member"
	OpOrgReadGetOrgs        = "read:get-orgs"
	OpOrgWriteUpdateMember  = "write:update-member"

	// Build operations
	OpBuildCreateSave   = "create:save"
	OpBuildReadGet      = "read:get"
	OpBuildReadList     = "read:list"
	OpBuildReadDownload = "read:download"
	OpBuildDeleteDelete = "delete:delete"
	OpBuildReadDescribe = "read:describe"

	// Deployment operations
	OpDeploymentWriteActivate   = "write:activate"
	OpDeploymentWriteTest       = "write:test"
	OpDeploymentWriteCreate     = "write:create"
	OpDeploymentWriteDeactivate = "write:deactivate"
	OpDeploymentDeleteDelete    = "delete:delete"
	OpDeploymentReadList        = "read:list"
	OpDeploymentReadGet         = "read:get"

	// Project operations
	OpProjectCreateCreate          = "create:create"
	OpProjectDeleteDelete          = "delete:delete"
	OpProjectUpdateUpdate          = "update:update"
	OpProjectReadGet               = "read:get"
	OpProjectReadList              = "read:list"
	OpProjectWriteBuild            = "write:build"
	OpProjectWriteSetResources     = "write:set-resources"
	OpProjectReadDownloadResources = "read:download-resources"
	OpProjectReadExport            = "read:export"
	OpProjectReadLint              = "read:lint"

	// Event operations
	OpEventReadGet    = "read:get"
	OpEventReadList   = "read:list"
	OpEventCreateSave = "create:save"

	// Trigger operations
	OpTriggerWriteCreate  = "write:create"
	OpTriggerUpdateUpdate = "update:update"
	OpTriggerWriteDelete  = "write:delete"
	OpTriggerReadGet      = "read:get"
	OpTriggerReadList     = "read:list"

	// Dispatcher operations
	OpDispatch   = "dispatch"
	OpRedispatch = "redispatch"

	// User operations
	OpUserCreateCreate = "create:create"
	OpUserReadGet      = "read:get"
	OpUserReadGetID    = "read:get-id"
	OpUserUpdateUpdate = "update:update"

	// Variable operations
	OpVarWriteSetVar              = "write:set-var"
	OpVarWriteDeleteAllVars       = "write:delete-all-vars"
	OpVarWriteDeleteVar           = "write:delete-var"
	OpVarReadGetAllVars           = "read:get-all-vars"
	OpVarReadGetVar               = "read:get-var"
	OpVarReadFindVarConnectionIDs = "read:find-var-connections-ids"

	// Integration operations
	OpIntegrationGet  = "get"
	OpIntegrationList = "list"

	// Connection operations
	OpConnectionWriteCreate  = "write:create"
	OpConnectionUpdateUpdate = "update:update"
	OpConnectionDeleteDelete = "delete:delete"
	OpConnectionReadList     = "read:list"
	OpConnectionTest         = "test"
	OpConnectionRefresh      = "refresh"
	OpConnectionReadGet      = "read:get"

	// Session operations
	OpSessionReadGetPrints   = "read:get-prints"
	OpSessionReadGetLog      = "read:get-log"
	OpSessionReadDownloadLog = "read:download-log"
	OpSessionReadGet         = "read:get"
	OpSessionWriteStop       = "write:stop"
	OpSessionReadList        = "read:list"
	OpSessionDeleteDelete    = "delete:delete"
	OpSessionCreateStart     = "create:start"
)
