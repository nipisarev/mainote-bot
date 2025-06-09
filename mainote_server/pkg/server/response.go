package server

import (
	"encoding/json"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
)

type ImplResponse struct {
	Code int
	Body any
}

type ErrorResponse struct {
	Type         string `json:"type,omitempty"`
	ErrorMessage string `json:"errorMessage"`
}

func NewErrorResponseBody(errorMessage string) ErrorResponse {
	return ErrorResponse{
		ErrorMessage: errorMessage,
	}
}

func NewErrorResponseBodyWithType(errType, errorMessage string) ErrorResponse {
	return ErrorResponse{
		Type:         errType,
		ErrorMessage: errorMessage,
	}
}

func SendErrorMessageResponse(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	SendErrorResponse(w, r, statusCode, NewErrorResponseBody(message))
}

func SendErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errResp any) {
	if err := EncodeJSONResponse(errResp, statusCode, w); err != nil {
		log.Error().Ctx(r.Context()).Err(err).Msg("Failed to send error response")
	}
}

func EncodeJSONResponse(i any, status int, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if i != nil {
		return errors.WithStack(json.NewEncoder(w).Encode(i))
	}

	return nil
}
