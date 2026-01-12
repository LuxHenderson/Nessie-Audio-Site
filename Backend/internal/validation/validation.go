package validation

import (
	"regexp"
	"strings"

	apierrors "github.com/nessieaudio/ecommerce-backend/internal/errors"
)

var (
	// Email regex pattern
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Phone regex pattern (flexible, allows various formats)
	phoneRegex = regexp.MustCompile(`^[\d\s\-\(\)\+]+$`)
)

// Validator accumulates validation errors
type Validator struct {
	errors []apierrors.ValidationError
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		errors: make([]apierrors.ValidationError, 0),
	}
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, apierrors.ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors
func (v *Validator) Errors() []apierrors.ValidationError {
	return v.errors
}

// RequireString validates a string is not empty
func (v *Validator) RequireString(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, field+" is required")
	}
}

// RequireEmail validates an email address
func (v *Validator) RequireEmail(field, value string) {
	v.RequireString(field, value)
	if strings.TrimSpace(value) != "" && !emailRegex.MatchString(value) {
		v.AddError(field, field+" must be a valid email address")
	}
}

// RequirePhone validates a phone number (basic format check)
func (v *Validator) RequirePhone(field, value string) {
	if strings.TrimSpace(value) == "" {
		return // Phone is optional in most cases
	}
	if !phoneRegex.MatchString(value) {
		v.AddError(field, field+" must be a valid phone number")
	}
}

// RequireInt validates an integer is positive
func (v *Validator) RequirePositiveInt(field string, value int) {
	if value <= 0 {
		v.AddError(field, field+" must be a positive number")
	}
}

// RequireFloat validates a float is positive
func (v *Validator) RequirePositiveFloat(field string, value float64) {
	if value <= 0 {
		v.AddError(field, field+" must be a positive number")
	}
}

// RequireMinLength validates minimum string length
func (v *Validator) RequireMinLength(field, value string, minLength int) {
	if len(strings.TrimSpace(value)) < minLength {
		v.AddError(field, field+" must be at least "+string(rune(minLength))+" characters")
	}
}

// RequireMaxLength validates maximum string length
func (v *Validator) RequireMaxLength(field, value string, maxLength int) {
	if len(value) > maxLength {
		v.AddError(field, field+" must be no more than "+string(rune(maxLength))+" characters")
	}
}

// RequireOneOf validates value is in allowed list
func (v *Validator) RequireOneOf(field, value string, allowed []string) {
	found := false
	for _, a := range allowed {
		if value == a {
			found = true
			break
		}
	}
	if !found {
		v.AddError(field, field+" must be one of: "+strings.Join(allowed, ", "))
	}
}

// ValidateAddress validates a shipping address
func (v *Validator) ValidateAddress(prefix string, name, address1, city, state, zip, country string) {
	v.RequireString(prefix+"_name", name)
	v.RequireString(prefix+"_address1", address1)
	v.RequireString(prefix+"_city", city)
	v.RequireString(prefix+"_state", state)
	v.RequireString(prefix+"_zip", zip)
	v.RequireString(prefix+"_country", country)

	// Basic ZIP code format check (US or basic international)
	if zip != "" && len(zip) < 3 {
		v.AddError(prefix+"_zip", "Postal code is too short")
	}
}
