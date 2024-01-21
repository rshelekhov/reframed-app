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

			// Create handler
			handler := &handlers.UserHandler{
				Storage: mockStorage,
				Logger:  mockLogger,
			}

			mockStorage.
				On("CreateUser", mock.Anything, mock.AnythingOfType("models.User")).
				Return(tc.expectedError).
				Once()

			// Create request
			reqBody, _ := json.Marshal(tc.user)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(reqBody))

			// Call handler
			rr := httptest.NewRecorder()
			handler.CreateUser()(rr, req)

			// Assert
			require.Equal(t, tc.expectedCode, rr.Code)
			if tc.expectedError != nil {
				require.Contains(t, rr.Body.String(), tc.expectedError.Error())
			}
		})
	}
}

func TestUserHandler_GetUserByID(t *testing.T) {
	// Setup
	mockStorage := &mocks.UserStorage{}
	mockLogger := slogdiscard.NewDiscardLogger()

	handler := &handlers.UserHandler{
		Storage: mockStorage,
		Logger:  mockLogger,
	}

	router := chi.NewRouter()
	router.HandleFunc("/user/{id}", handler.GetUserByID())

	t.Run("Get User By ID", func(t *testing.T) {
		// Create a request to get user by ID
		req := httptest.NewRequest("GET", "/user/123", nil)
		rr := httptest.NewRecorder()

		mockStorage.On("GetUserByID", mock.Anything, mock.AnythingOfType("string")).
			Return(models.User{
				ID:    "123",
				Email: "test@example.com",
			}, nil).
			Once()

		// Serve the request using the router
		router.ServeHTTP(rr, req)

		resp := rr.Result()

		// Verify the response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		// Add more assertions for response content, if necessary
	})

	t.Run("User Not Found", func(t *testing.T) {
		// Create a request with invalid user ID
		req := httptest.NewRequest("GET", "/user/456", nil)
		rr := httptest.NewRecorder()

		mockStorage.On("GetUserByID", mock.Anything, mock.AnythingOfType("string")).
			Return(models.User{}, storage.ErrUserNotFound).
			Once()

		router.ServeHTTP(rr, req)

		resp := rr.Result()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		assert.Nil(t, err)
		assert.Contains(t, string(body), storage.ErrUserNotFound.Error())
	})

	t.Run("Empty ID", func(t *testing.T) {
		// Create a request with empty user ID
		req := httptest.NewRequest("GET", "/user/{id}", nil)
		rr := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "")

		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.GetUserByID()(rr, req)

		resp := rr.Result()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
