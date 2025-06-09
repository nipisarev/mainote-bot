package validator

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	fields []Field
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{fields: []Field{{Field: field, Message: message}}}
}

func (e ValidationError) Error() string {
	messages := make([]string, 0, len(e.fields))
	for _, field := range e.fields {
		messages = append(messages, field.Message)
	}
	return fmt.Sprintf("Validation errors: %s ", strings.Join(messages, ", "))
}

func (e ValidationError) Fields() []Field {
	return e.fields
}

type Field struct {
	Field   string
	Message string
}
