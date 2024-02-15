package v1_test

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	http2 "github.com/rshelekhov/reframed/internal/controller/http/v1"
	"github.com/rshelekhov/reframed/pkg/logger/slogdiscard"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetID(t *testing.T) {
	mockLogger := slogdiscard.NewDiscardLogger()

	t.Run("valid ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/path/123", nil)
		rr := httptest.NewRecorder()

		router := chi.NewRouter()
		router.Get("/path/{id}", func(w http.ResponseWriter, r *http.Request) {
			id, statusCode, err := GetIDFromURL(r, mockLogger)
			assert.NoError(t, err)
			assert.Equal(t, "123", id)
			assert.Equal(t, http.StatusOK, statusCode)
		})

		router.ServeHTTP(rr, req)
	})

	t.Run("empty ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/path/", nil)

		_, statusCode, err := GetIDFromURL(req, mockLogger)

		assert.Equal(t, ErrEmptyID, err)
		assert.Equal(t, http.StatusBadRequest, statusCode)

	})
}

func TestDecodeJSON(t *testing.T) {
	type TestData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	testCases := []struct {
		name          string
		body          string
		expectedCode  int
		expectedError error
	}{
		{
			name:          "valid JSON",
			body:          `{"email": "<EMAIL>", "password": "<PASSWORD>"}`,
			expectedError: nil,
		},
		{
			name:          "invalid JSON",
			body:          `{invalid}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrInvalidJSON,
		},
		{
			name:          "empty body",
			body:          "",
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrEmptyRequestBody,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockLogger := slogdiscard.NewDiscardLogger()

			reqBody := bytes.NewBufferString(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/", reqBody)
			rr := httptest.NewRecorder()

			err := http2.decodeJSON(rr, req, mockLogger, &TestData{})

			if err != nil {
				assert.Equal(t, tc.expectedError, err)
				assert.Equal(t, tc.expectedCode, rr.Code)
			}
		})
	}
}

func TestValidateData(t *testing.T) {
	type TestData struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	testCases := []struct {
		name          string
		data          interface{}
		expectedCode  int
		expectedError error
	}{
		{
			name:          "valid data",
			data:          TestData{Email: "john@example.com", Password: "password123"},
			expectedCode:  http.StatusOK,
			expectedError: nil,
		},
		{
			name:          "invalid data",
			data:          TestData{Email: "alice.example.com", Password: "pass"},
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrInvalidData,
		},
		{
			name:          "empty data",
			data:          nil,
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrEmptyData,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger := slogdiscard.NewDiscardLogger()

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			rr := httptest.NewRecorder()

			err := http2.validateData(rr, req, mockLogger, tc.data)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err)
				assert.Equal(t, tc.expectedCode, rr.Code)
				assert.Contains(t, rr.Body.String(), tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCode, rr.Code)
			}
		})
	}
}
