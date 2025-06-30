package dto

// ErrorDetail represents detailed error information in Stripe style
type ErrorDetail struct {
	Type    string `json:"type"`              // e.g., "validation_error", "authentication_error"
	Code    string `json:"code,omitempty"`    // e.g., "missing_field", "invalid_format"
	Message string `json:"message"`           // Human-readable error message
	Param   string `json:"param,omitempty"`   // Parameter that caused the error
}

// ErrorResponse follows Stripe API style
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// NewError creates a Stripe-style error response
func NewError(message string, errorType ...string) ErrorResponse {
	errType := "api_error"
	if len(errorType) > 0 {
		errType = errorType[0]
	}
	return ErrorResponse{
		Error: ErrorDetail{
			Type:    errType,
			Message: message,
		},
	}
}

// NewValidationError creates a validation error with specific field
func NewValidationError(message, param string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Type:    "validation_error",
			Code:    "invalid_parameter",
			Message: message,
			Param:   param,
		},
	}
}

// NewAuthError creates an authentication error
func NewAuthError(message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Type:    "authentication_error",
			Message: message,
		},
	}
}

// NewNotFoundError creates a resource not found error
func NewNotFoundError(resource string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Type:    "invalid_request_error",
			Code:    "resource_missing",
			Message: resource + " not found",
		},
	}
}