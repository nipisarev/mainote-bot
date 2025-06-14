package errors_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"

	srvErrs "mainote-backend/pkg/server/errors"
)

func TestErrorHandler(t *testing.T) {
	cases := []struct {
		name       string
		req        *http.Request
		err        error
		response   *generated.ImplResponse
		wantStatus int
		wantBody   string
	}{
		{
			name:       "NilError",
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
			err:        nil,
			response:   &public.ImplResponse{},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"errorMessage":"Internal error."}`,
		},
		{
			name:       "ParsingError",
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
			err:        &public.ParsingError{Err: errors.New("parsing error")},
			response:   &public.ImplResponse{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"errorMessage":"Invalid payload, expected valid json."}`,
		},
		{
			name:       "RequiredError",
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
			err:        &public.RequiredError{Field: "nil.field"},
			response:   &public.ImplResponse{},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"errorMessage":"Data is invalid.","errorFields":[{"field":"nil.field","error":"required field is missing."}]}`,
		},
		{
			name:       "GenericError",
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
			err:        errors.New("generic error"),
			response:   &public.ImplResponse{},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"errorMessage":"Internal error."}`,
		},
		{
			name: "ResponseWithErrorCode",
			req:  httptest.NewRequest(http.MethodGet, "/", nil),
			err:  errors.New("generic error"),
			response: &public.ImplResponse{
				Code: http.StatusBadRequest,
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"errorMessage":"Internal error."}`,
		},
		{
			name: "ResponseWithErrorBody",
			req:  httptest.NewRequest(http.MethodGet, "/", nil),
			err:  errors.New("generic error"),
			response: &public.ImplResponse{
				Body: map[string]string{"error": "error message"},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"error message"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			srvErrs.ErrorHandler[*public.ParsingError, *public.RequiredError, public.ImplResponse](w, tc.req, tc.err, tc.response)

			resp := w.Result()
			require.Equal(t, tc.wantStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			require.Equal(t, tc.wantBody+"\n", string(body))
		})
	}
}
