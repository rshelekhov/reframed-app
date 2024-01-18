package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/rshelekhov/reframed/internal/api/handlers"
	"github.com/rshelekhov/reframed/internal/api/handlers/mocks"
	"github.com/rshelekhov/reframed/internal/logger/slogdiscard"
	"github.com/rshelekhov/reframed/internal/models"
	"github.com/rshelekhov/reframed/internal/storage"
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
			name: "valid",
			user: models.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode:  http.StatusCreated,
			expectedError: nil,
		},
		{
			name: "Invalid email",
			user: models.User{
				Email:    "invalid",
				Password: "password123",
			},
			expectedCode:  http.StatusUnprocessableEntity,
			expectedError: errors.New("field Email must be a valid email address"),
		},
		{
			name: "User already exists",
			user: models.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: storage.ErrUserAlreadyExists,
		},
		// TODO: Add more test cases
	}

	// Create mocks
	mockStorage := &mocks.UserStorage{}
	mockLogger := slogdiscard.NewDiscardLogger()

	// Create handler
	handler := &handlers.UserHandler{
		Storage: mockStorage,
		Logger:  mockLogger,
	}

	for _, tc := range testCases {
		tc := tc // Create a local copy for parallel tests

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Mock storage create user
			/*mockStorage.
			On("CreateUser", mock.Anything, mock.Anything).
			Return(tc.expectedError).
			Once()*/
			if tc.name == "User already exists" {
				mockStorage.
					On("CreateUser", mock.Anything, mock.Anything).
					Return(storage.ErrUserAlreadyExists).
					Once()
			} else if tc.name == "Invalid email" {
				mockStorage.
					On("CreateUser", mock.Anything, mock.Anything).
					Return(errors.New("field Email must be a valid email address")).
					Once()
			} else if tc.name == "valid" {
				mockStorage.
					On("CreateUser", mock.Anything, mock.Anything).
					Return(nil).
					Once()
			}

			// Create request
			reqBody, _ := json.Marshal(tc.user)
			req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(reqBody))
			require.NoError(t, err)

			// Call handler
			rr := httptest.NewRecorder()
			handler.CreateUser()(rr, req)

			// Assert
			require.Equal(t, tc.expectedCode, rr.Code)

			// TODO: Assert other expectations

		})
	}

}
