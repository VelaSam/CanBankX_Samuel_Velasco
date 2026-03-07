package errors

//this represent the standard error format selon RFC 7807 and ADR-0003
type APIError struct {
	ErrorCode string   `json:"error_code"`
	Message   string   `json:"message"`
	Details   []string `json:"details,omitempty"`
}

var (
	ErrEmailAlreadyInUse = &APIError{ErrorCode: "EMAIL_EXISTS", Message: "Email is already in use"}
	ErrInvalidInput      = &APIError{ErrorCode: "BAD_REQUEST", Message: "Invalid input data"}
	ErrInternalServer    = &APIError{ErrorCode: "INTERNAL_ERROR", Message: "An unexpected error occurred"}
	ErrUserNotFound      = &APIError{ErrorCode: "USER_NOT_FOUND", Message: "User not found"}
)
