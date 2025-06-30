package dto

import "fxserver/pkg/validator"

// ParseValidationErrors provides a standardized way to parse validation errors
func ParseValidationErrors(err error) map[string]string {
	details := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for field, message := range validationErrors.Errors {
			details[field] = message
		}
	} else {
		details["validation"] = err.Error()
	}
	return details
}

// NewValidationErrors creates validation errors for multiple fields
func NewValidationErrors(err error) ErrorResponse {
	details := ParseValidationErrors(err)
	
	// For multiple validation errors, we'll use the first one as primary
	// and include others in a general message
	var message string
	var param string
	
	if len(details) == 1 {
		for field, msg := range details {
			message = msg
			param = field
			break
		}
	} else {
		message = "Multiple validation errors occurred"
		// Could be enhanced to show all errors
	}
	
	return ErrorResponse{
		Error: ErrorDetail{
			Type:    "validation_error",
			Code:    "invalid_parameters",
			Message: message,
			Param:   param,
		},
	}
}