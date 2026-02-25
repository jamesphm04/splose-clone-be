package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidators registers custom validators for the validator package
func RegisterCustomValidators(validate *validator.Validate) {
	validate.RegisterValidation("phoneNumber", validatePhoneNumber)
	validate.RegisterValidation("dateOfBirth", validateDateOfBirth)
}

// validatePhoneNumber validates a phone number
func validatePhoneNumber(fl validator.FieldLevel) bool {
	phoneNumber := fl.Field().String()
	return regexp.MustCompile(`^(\+?[1-9]\d{1,14}|0\d{8,14})$`).MatchString(phoneNumber)
}

// validateDateOfBirth validates a date of birth
func validateDateOfBirth(fl validator.FieldLevel) bool {
	dateOfBirth := fl.Field().String()
	return regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(dateOfBirth)
}
