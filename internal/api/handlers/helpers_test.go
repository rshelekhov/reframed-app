package handlers_test

import (
	"bytes"
	"fmt"
	"github.com/magiconair/properties/assert"
	"github.com/rshelekhov/reframed/internal/api/handlers"
	"github.com/rshelekhov/reframed/internal/logger/slogdiscard"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
			expectedError: fmt.Errorf(handlers.ErrInvalidJSON),
		},
		{
			name:          "empty body",
			body:          "",
			expectedCode:  http.StatusBadRequest,
			expectedError: fmt.Errorf(handlers.ErrEmptyRequestBody),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			loggerMock := slogdiscard.NewDiscardLogger()

			reqBody := bytes.NewBufferString(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/", reqBody)
			rr := httptest.NewRecorder()

			err := handlers.DecodeJSON(rr, req, loggerMock, &TestData{})

			if err != nil {
				response := rr.Result()

				assert.Equal(t, tc.expectedError, err)
				assert.Equal(t, tc.expectedCode, response.StatusCode)
			}
		})
	}
}

/*
func TestValidateData(t *testing.T) {
	type TestData struct {
		Name string `validate:"required"`
	}

	testCases := []struct {
		name          string
		data          TestData
		expectedCode  int
		expectedError error
	}{
		{},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			loggerMock := slogdiscard.NewDiscardLogger()

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			rr := httptest.NewRecorder()

			data := struct {
				Name string `validate:"required"`
			}{}

			err := handlers.ValidateData(rr, req, loggerMock, data)*/
/*
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		response := rr.Result()
		assert.Equal(t, tc.expectedError, err)
		assert.Equal(t, tc.expectedCode, response.StatusCode)
	}

	if err != nil {
		response := rr.Result()
		assert.Equal(t, tc.expectedError, err)
		assert.Equal(t, tc.expectedCode, response.StatusCode)
	}*/

/*if err := handlers.ValidateData(tt.args.w, tt.args.r, tt.args.log, tt.args.data); (err != nil) != tt.wantErr {
	t.Errorf("ValidateData() error = %v, wantErr %v", err, tt.wantErr)
}*/
/*})
	}
}*/

func TestValidateData2(t *testing.T) {
	type TestData struct {
		Name  string `json:"name" validate:"required,email"`
		Age   int    `json:"age" validate:"required,min=18"`
		Email string `json:"email" validate:"required,email"`
	}
	// var mockData TestData

	testCases := []struct {
		name       string
		data       interface{}
		wantErrMsg string
	}{
		{
			name:       "Valid Data",
			data:       TestData{Name: "John", Age: 25, Email: "john@example.com"},
			wantErrMsg: "",
		},
		{
			name:       "Invalid Data",
			data:       TestData{Name: "Alice", Email: "alice.example.com"},
			wantErrMsg: handlers.ErrInvalidData,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loggerMock := slogdiscard.NewDiscardLogger()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", nil)

			err := handlers.ValidateData(rec, req, loggerMock, tc.data)

			if err != nil {
				assert.Equal(t, tc.wantErrMsg, err.Error())
				response := rec.Result()
				assert.Equal(t, http.StatusBadRequest, response.StatusCode)
			} else {
				assert.Equal(t, "", tc.wantErrMsg)
				response := rec.Result()
				assert.Equal(t, http.StatusOK, response.StatusCode)
			}
		})
	}
}
