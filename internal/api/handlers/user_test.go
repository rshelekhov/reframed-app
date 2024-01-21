package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rshelekhov/reframed/internal/api/handlers"
	"github.com/rshelekhov/reframed/internal/logger/slogdiscard"
	"github.com/rshelekhov/reframed/internal/models"
	"github.com/rshelekhov/reframed/internal/storage"
	"github.com/rshelekhov/reframed/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
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

			handler := &handlers.UserHandler{
				Storage: mockStorage,
				Logger:  mockLogger,
			}

			mockStorage.
				On("CreateUser", mock.Anything, mock.AnythingOfType("models.User")).
				Return(tc.expectedError).
				Once()

			reqBody, _ := json.Marshal(tc.user)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(reqBody))

			rr := httptest.NewRecorder()
			handler.CreateUser()(rr, req)

			require.Equal(t, tc.expectedCode, rr.Code)
			if tc.expectedError != nil {
				require.Contains(t, rr.Body.String(), tc.expectedError.Error())
			}
		})
	}
}

func TestUserHandler_GetUserByID(t *testing.T) {
	testCases := []struct {
		name          string
		userID        string
		user          models.User
		expectedCode  int
		expectedError error
	}{
		{
			name:   "success",
			userID: "123",
			user: models.User{
				ID:    "123",
				Email: "test@example.com",
			},
			expectedCode:  http.StatusOK,
			expectedError: nil,
		},
		{
			name:          "user not found",
			userID:        "123",
			user:          models.User{},
			expectedCode:  http.StatusNotFound,
			expectedError: storage.ErrUserNotFound,
		},
		{
			name:          "empty ID",
			userID:        "",
			user:          models.User{},
			expectedCode:  http.StatusBadRequest,
			expectedError: handlers.ErrEmptyID,
		},
		{
			name:          "failed to get user",
			userID:        "123",
			user:          models.User{},
			expectedCode:  http.StatusInternalServerError,
			expectedError: handlers.ErrFailedToGetData,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage := &mocks.UserStorage{}
			mockLogger := slogdiscard.NewDiscardLogger()

			handler := &handlers.UserHandler{
				Storage: mockStorage,
				Logger:  mockLogger,
			}

			mockStorage.On("GetUserByID", mock.Anything, mock.AnythingOfType("string")).
				Return(tc.user, tc.expectedError).
				Once()

			req := httptest.NewRequest("GET", "/user/{id}", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.userID)

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.GetUserByID()(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)

			if tc.name == "user not found" {
				body, err := io.ReadAll(rr.Body)

				assert.Nil(t, err)
				assert.Contains(t, string(body), storage.ErrUserNotFound.Error())
			}
		})
	}
}
