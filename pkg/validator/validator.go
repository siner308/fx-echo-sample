package validator

import (
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validator *validator.Validate
}

type ValidationErrors struct {
	Errors map[string]string
}

func (ve ValidationErrors) Error() string {
	return "validation failed"
}

func New() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

func (v *Validator) Validate(i interface{}) error {
	err := v.validator.Struct(i)
	if err != nil {
		// Convert validator errors to our custom format
		errors := make(map[string]string)
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, validationError := range validationErrors {
				errors[validationError.Field()] = validationError.Tag()
			}
		} else {
			errors["validation"] = err.Error()
		}
		return ValidationErrors{Errors: errors}
	}
	return nil
}
