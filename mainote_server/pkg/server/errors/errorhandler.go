package errors

import (
	"encoding/json"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"

	"mainote-backend/pkg/server"
	"mainote-backend/pkg/validator"
)

type FieldProvider interface {
	FieldName() string
}

func ErrorHandler[ParsingError error, RequiredError FieldProvider, ImplResponse server.ImplResponse](w http.ResponseWriter, r *http.Request, err error, result *ImplResponse) {
	var parsingErr ParsingError
	var requiredErr RequiredError
	var validationErr *validator.ValidationError

	switch {
	case errors.As(err, &parsingErr):
		sendErrorResponse(w, r, http.StatusBadRequest, ErrorResponse{
			ErrorMessage: "Invalid payload, expected valid json.",
		})
	case errors.As(err, &requiredErr):
		sendErrorResponse(w, r, http.StatusUnprocessableEntity, ErrorResponse{
			ErrorMessage: "Data is invalid.",
			ErrorFields: []ErrorField{
				{Field: requiredErr.FieldName(), Error: "required field is missing."},
			},
		})
	case errors.As(err, &validationErr):
		errorFields := make([]ErrorField, 0, len(validationErr.Fields()))
		for _, field := range validationErr.Fields() {
			errorFields = append(errorFields, ErrorField{Field: field.Field, Error: field.Message})
		}
		sendErrorResponse(w, r, http.StatusUnprocessableEntity, ErrorResponse{
			ErrorMessage: "Data is invalid.",
			ErrorFields:  errorFields,
		})
	default:
		errorCode := http.StatusInternalServerError
		var errorBody any = ErrorResponse{
			ErrorMessage: "Internal error.",
		}

		// Use response code and body if provided
		if result != nil {
			// Since ImplResponse is constrained to server.ImplResponse, we can use type assertion
			if implResp, ok := any(result).(*server.ImplResponse); ok {
				if implResp.Code != 0 {
					errorCode = implResp.Code
				}
				if implResp.Body != nil {
					errorBody = implResp.Body
				}
			}
		}

		log.Error().Ctx(r.Context()).Err(err).Msg("Error response")
		sendErrorResponse(w, r, errorCode, errorBody)
	}
}

func sendErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errResp any) {
	if err := encodeJSONResponse(errResp, &statusCode, w); err != nil {
		log.Error().Ctx(r.Context()).Err(err).Msg("Failed to send error response")
	}
}

func encodeJSONResponse(i any, status *int, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if status != nil {
		w.WriteHeader(*status)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if i != nil {
		return errors.WithStack(json.NewEncoder(w).Encode(i))
	}

	return nil
}
