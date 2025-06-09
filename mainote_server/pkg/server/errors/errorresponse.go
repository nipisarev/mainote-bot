package errors

type ErrorResponse struct {
	ErrorMessage string `json:"errorMessage"`

	ErrorFields []ErrorField `json:"errorFields,omitempty"`
}

type ErrorField struct {
	Field string `json:"field,omitempty"`

	Error string `json:"error,omitempty"`
}
