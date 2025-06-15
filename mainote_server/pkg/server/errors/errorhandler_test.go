package errors_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"

	api "mainote-backend/pkg/generated/api"
	srvErrs "mainote-backend/pkg/server/errors"
)

func TestErrorHandler(t *testing.T) {
	cases := []struct {
		name       string
		req        *http.Request
		err        error
		response   *api.ImplResponse
		wantStatus int
		wantBody   string
	}{
		{
			name:       "NilError",
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
			err:        nil,
			response:   &api.ImplResponse{},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"errorMessage":"Internal error."}`,
		},
		{
			name:       "ParsingError",
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
			err:        &api.ParsingError{Err: errors.New("parsing error")},
			response:   &api.ImplResponse{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"errorMessage":"Invalid payload, expected valid json."}`,
		},
		{
			name:       "RequiredError",
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
			err:        &api.RequiredError{Field: "nil.field"},
			response:   &api.ImplResponse{},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"errorMessage":"Data is invalid.","errorFields":[{"field":"nil.field","error":"required field is missing."}]}`,
		},
		{
			name:       "GenericError",
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
			err:        errors.New("generic error"),
			response:   &api.ImplResponse{},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"errorMessage":"Internal error."}`,
		},
		{
			name: "ResponseWithErrorCode",
			req:  httptest.NewRequest(http.MethodGet, "/", nil),
			err:  errors.New("generic error"),
			response: &api.ImplResponse{
				Code: http.StatusBadRequest,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"errorMessage":"Internal error."}`,
		},
		{
			name: "ResponseWithErrorBody",
			req:  httptest.NewRequest(http.MethodGet, "/", nil),
			err:  errors.New("generic error"),
			response: &api.ImplResponse{
				Body: map[string]string{"error": "error message"},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"error message"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			srvErrs.ErrorHandler[*api.ParsingError, *api.RequiredError, api.ImplResponse](w, tc.req, tc.err, tc.response)

			resp := w.Result()
			require.Equal(t, tc.wantStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			require.Equal(t, tc.wantBody+"\n", string(body))
		})
	}
}
