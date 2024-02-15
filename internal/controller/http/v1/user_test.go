package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	http2 "github.com/rshelekhov/reframed/internal/controller/http/v1"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/storage"
	"github.com/rshelekhov/reframed/internal/storage/mocks"
	"github.com/rshelekhov/reframed/pkg/logger/slogdiscard"
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
		name                     string
		user                     model.User
		expectedCode             int
		expectedUserStorageError error
		expectedListStorageError error
	}{
		{
			name: "success",
			user: model.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode:             http.StatusCreated,
			expectedUserStorageError: nil,
			expectedListStorageError: nil,
		},
		{
			name: "invalid email",
			user: model.User{
				Email:    "testexample.com",
				Password: "password123",
			},
			expectedCode:             http.StatusBadRequest,
			expectedUserStorageError: errors.New("field Email must be a valid email address"),
			expectedListStorageError: nil,
		},
		{
			name: "invalid password",
			user: model.User{
				Email:    "test@example.com",
				Password: "pass",
			},
			expectedCode:             http.StatusBadRequest,
			expectedUserStorageError: errors.New("field Password must be greater than or equal to 8"),
			expectedListStorageError: nil,
		},
		{
			name: "user already exists",
			user: model.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode:             http.StatusBadRequest,
			expectedUserStorageError: storage.ErrUserAlreadyExists,
			expectedListStorageError: nil,
		},
		{
			name: "email is required",
			user: model.User{
				Password: "password123",
			},
			expectedCode:             http.StatusBadRequest,
			expectedUserStorageError: errors.New("field Email is required"),
			expectedListStorageError: nil,
		},
		{
			name: "failed to create Inbox list",
			user: model.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode:             http.StatusInternalServerError,
			expectedUserStorageError: nil,
			expectedListStorageError: errors.New("failed to create list"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockUserStorage := &mocks.UserStorage{}
			mockListStorage := &mocks.ListStorage{}
			mockLogger := slogdiscard.NewDiscardLogger()

			handler := &http2.UserHandler{
				userStorage: mockUserStorage,
				listStorage: mockListStorage,
				logger:      mockLogger,
			}

			mockUserStorage.
				On("CreateUser", mock.Anything, mock.AnythingOfType("model.User")).
				Return(tc.expectedUserStorageError).
				Once()

			mockListStorage.
				On("CreateList", mock.Anything, mock.AnythingOfType("model.List")).
				Return(tc.expectedListStorageError).
				Once()

			reqBody, _ := json.Marshal(tc.user)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(reqBody))

			rr := httptest.NewRecorder()
			handler.CreateUser()(rr, req)

			if tc.expectedUserStorageError != nil {
				assert.Equal(t, tc.expectedCode, rr.Code)
				require.Contains(t, rr.Body.String(), tc.expectedUserStorageError.Error())
			} else if tc.expectedListStorageError != nil {
				assert.Equal(t, tc.expectedCode, rr.Code)
				require.Contains(t, rr.Body.String(), tc.expectedListStorageError.Error())
			} else {
				require.Equal(t, tc.expectedCode, rr.Code)
			}
		})
	}
}

func TestUserHandler_GetUserByID(t *testing.T) {
	testCases := []struct {
		name          string
		userID        string
		user          model.User
		expectedCode  int
		expectedError error
	}{
		{
			name:   "success",
			userID: "123",
			user: model.User{
				ID:    "123",
				Email: "test@example.com",
			},
			expectedCode:  http.StatusOK,
			expectedError: nil,
		},
		{
			name:          "user not found",
			userID:        "123",
			user:          model.User{},
			expectedCode:  http.StatusNotFound,
			expectedError: storage.ErrUserNotFound,
		},
		{
			name:          "empty ID",
			userID:        "",
			user:          model.User{},
			expectedCode:  http.StatusBadRequest,
			expectedError: ErrEmptyID,
		},
		{
			name:          "failed to get user",
			userID:        "123",
			user:          model.User{},
			expectedCode:  http.StatusInternalServerError,
			expectedError: ErrFailedToGetData,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockStorage := &mocks.UserStorage{}
			mockLogger := slogdiscard.NewDiscardLogger()

			handler := &http2.UserHandler{
				userStorage: mockStorage,
				logger:      mockLogger,
			}

			mockStorage.On("GetUserData", mock.Anything, mock.AnythingOfType("string")).
				Return(tc.user, tc.expectedError).
				Once()

			req := httptest.NewRequest(http.MethodGet, "/user/{id}", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.userID)

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.GetUser()(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)

			if tc.name == "user not found" {
				body, err := io.ReadAll(rr.Body)

				assert.Nil(t, err)
				assert.Contains(t, string(body), storage.ErrUserNotFound.Error())
			}
		})
	}
}

func TestUserHandler_GetUsers(t *testing.T) {
	testCases := []struct {
		name          string
		url           string
		users         []model.User
		expectedCode  int
		expectedError error
	}{
		{
			name: "success",
			url:  "/users?limit=100&offset=0",
			users: []model.User{
				{
					ID:    "123",
					Email: "test@example.com",
				},
				{
					ID:    "456",
					Email: "test2@example.com",
				},
			},
			expectedCode:  http.StatusOK,
			expectedError: nil,
		},
		{
			name:          "no users found",
			url:           "/users?limit=100&offset=0",
			users:         []model.User{},
			expectedCode:  http.StatusNotFound,
			expectedError: storage.ErrNoUsersFound,
		},
		{
			name:          "failed to get users",
			url:           "/users?limit=100&offset=0",
			users:         []model.User{},
			expectedCode:  http.StatusInternalServerError,
			expectedError: errors.New("failed to get users"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockStorage := &mocks.UserStorage{}
			mockLogger := slogdiscard.NewDiscardLogger()

			handler := &http2.UserHandler{
				userStorage: mockStorage,
				logger:      mockLogger,
			}

			mockStorage.On("GetUsers", mock.Anything, mock.AnythingOfType("model.Pagination")).
				Return(tc.users, tc.expectedError).
				Once()

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.GetUsers()(w, r)
			}))
			defer ts.Close()

			resp, err := http.Get(ts.URL + tc.url)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode)

			if tc.expectedError != nil {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				require.Contains(t, string(body), tc.expectedError.Error())
			}

		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	testCases := []struct {
		name          string
		userID        string
		user          model.UpdateUser
		expectedCode  int
		expectedError error
	}{
		{
			name:   "success",
			userID: "123",
			user: model.UpdateUser{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedCode:  http.StatusOK,
			expectedError: nil,
		},
		{
			name:          "user not found",
			userID:        "123",
			user:          model.UpdateUser{},
			expectedCode:  http.StatusNotFound,
			expectedError: storage.ErrUserNotFound,
		},
		{
			name:          "email already taken",
			userID:        "123",
			user:          model.UpdateUser{},
			expectedCode:  http.StatusBadRequest,
			expectedError: storage.ErrEmailAlreadyTaken,
		},
		{
			name:          "no changes detected",
			userID:        "123",
			user:          model.UpdateUser{},
			expectedCode:  http.StatusBadRequest,
			expectedError: storage.ErrNoChangesDetected,
		},
		{
			name:          "no password changes detected (the same password)",
			userID:        "123",
			user:          model.UpdateUser{},
			expectedCode:  http.StatusBadRequest,
			expectedError: storage.ErrNoPasswordChangesDetected,
		},
		{
			name:          "failed to update user",
			userID:        "123",
			user:          model.UpdateUser{},
			expectedCode:  http.StatusInternalServerError,
			expectedError: errors.New("failed to update user"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockStorage := &mocks.UserStorage{}
			mockLogger := slogdiscard.NewDiscardLogger()

			handler := &http2.UserHandler{
				userStorage: mockStorage,
				logger:      mockLogger,
			}

			mockStorage.
				On("UpdateUser", mock.Anything, mock.AnythingOfType("model.User")).
				Return(tc.expectedError).
				Once()

			reqBody, _ := json.Marshal(tc.user)

			req := httptest.NewRequest(http.MethodPut, "/user/{id}", bytes.NewReader(reqBody))

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.userID)

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.UpdateUser()(rr, req)

			require.Equal(t, tc.expectedCode, rr.Code)
			if tc.expectedError != nil {
				require.Contains(t, rr.Body.String(), tc.expectedError.Error())
			}

			if tc.name == "user not found" {
				body, err := io.ReadAll(rr.Body)

				assert.Nil(t, err)
				assert.Contains(t, string(body), storage.ErrUserNotFound.Error())
			}
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	testCases := []struct {
		name          string
		userID        string
		expectedCode  int
		expectedError error
	}{
		{
			name:          "success",
			userID:        "123",
			expectedCode:  http.StatusOK,
			expectedError: nil,
		},
		{
			name:          "user not fount",
			userID:        "123",
			expectedCode:  http.StatusNotFound,
			expectedError: storage.ErrUserNotFound,
		},
		{
			name:          "failed to delete user",
			userID:        "123",
			expectedCode:  http.StatusInternalServerError,
			expectedError: errors.New("failed to delete user"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockStorage := &mocks.UserStorage{}
			mockLogger := slogdiscard.NewDiscardLogger()

			handler := &http2.UserHandler{
				userStorage: mockStorage,
				logger:      mockLogger,
			}

			mockStorage.
				On("DeleteUser", mock.Anything, mock.AnythingOfType("string")).
				Return(tc.expectedError).
				Once()

			req := httptest.NewRequest(http.MethodDelete, "/user/{id}", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.userID)

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.DeleteUser()(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)

			if tc.name == "user not found" {
				body, err := io.ReadAll(rr.Body)

				assert.Nil(t, err)
				assert.Contains(t, string(body), storage.ErrUserNotFound.Error())
			}
		})
	}
}
