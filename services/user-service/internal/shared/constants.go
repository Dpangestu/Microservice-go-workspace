package shared

// Pagination constants
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
	MinPageSize     = 1
)

// Error codes
const (
	ErrInvalidRequest  = "INVALID_REQUEST"
	ErrNotFound        = "NOT_FOUND"
	ErrUnauthorized    = "UNAUTHORIZED"
	ErrConflict        = "CONFLICT"
	ErrInternalError   = "INTERNAL_ERROR"
	ErrValidationError = "VALIDATION_ERROR"
)

// Log actions
const (
	LogRoleCreate = "ROLE_CREATE"
	LogRoleUpdate = "ROLE_UPDATE"
	LogRoleDelete = "ROLE_DELETE"
	LogPermAssign = "PERM_ASSIGN"
	LogPermRevoke = "PERM_REVOKE"
	LogUserCreate = "USER_CREATE"
	LogUserUpdate = "USER_UPDATE"
	LogUserDelete = "USER_DELETE"
)
