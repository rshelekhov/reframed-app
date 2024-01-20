package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/rshelekhov/reframed/internal/api/handlers"
	"github.com/rshelekhov/reframed/internal/logger/slogdiscard"
	"github.com/rshelekhov/reframed/internal/models"
	"github.com/rshelekhov/reframed/internal/storage"
	"github.com/rshelekhov/reframed/internal/storage/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserHandler_CreateUser(t *testing.T) {
	testCases := []struct {
		name          string
		user          models.User
		expectedCode  int
		expectedError error
	}{
		{
			name: "success",
			user: models.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode:  http.StatusCreated,
			expectedError: nil,
		},
		{
			name: "invalid email",
			user: models.User{
				Email:    "testexample.com",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: errors.New("field Email must be a valid email address"),
		},
		{
			name: "invalid password",
			user: models.User{
				Email:    "test@example.com",
				Password: "pass",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: errors.New("field Password must be greater than or equal to 8"),
		},
		{
			name: "user already exists",
			user: models.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: storage.ErrUserAlreadyExists,
		},
		{
			name: "email is required",
			user: models.User{
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: errors.New("field Email is required"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage := &mocks.UserStorage{}
			mockLogger := slogdiscard.NewDiscardLogger()

			// Create handler
			handler := &handlers.UserHandler{
				Storage: mockStorage,
				Logger:  mockLogger,
			}

			testCase := tc

			mockStorage.
				On("CreateUser", mock.Anything, mock.AnythingOfType("models.User")).
				Return(testCase.expectedError).
				Once()

			// Create request
			reqBody, _ := json.Marshal(testCase.user)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(reqBody))

			// Call handler
			rr := httptest.NewRecorder()
			handler.CreateUser()(rr, req)

			// Assert
			require.Equal(t, testCase.expectedCode, rr.Code)
			if testCase.expectedError != nil {
				require.Contains(t, rr.Body.String(), testCase.expectedError.Error())
			}
		})
	}
}
