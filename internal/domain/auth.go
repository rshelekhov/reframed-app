package domain

import (
	"context"
	"github.com/rshelekhov/reframed/pkg/httpserver/middleware/jwtoken"
	"time"
)

type (
	AuthUsecase interface {
		CreateUser(ctx context.Context, jwt *jwtoken.TokenService, data *UserRequestData) (string, error)
		CreateUserSession(ctx context.Context, jwt *jwtoken.TokenService, userID string, device UserDeviceRequestData) (jwtoken.TokenData, error)
		LoginUser(ctx context.Context, jwt *jwtoken.TokenService, data *UserRequestData) (string, error)
		CheckSessionAndDevice(ctx context.Context, refreshToken string, data UserDeviceRequestData) (Session, error)
		LogoutUser(ctx context.Context, userID string, data UserDeviceRequestData) error
		GetUserByID(ctx context.Context, id string) (UserResponseData, error)
		UpdateUser(ctx context.Context, jwt *jwtoken.TokenService, data *UserRequestData, userID string) error
		DeleteUser(ctx context.Context, userUD string, data UserDeviceRequestData) error
	}

	AuthStorage interface {
		CreateUser(ctx context.Context, user User) error
		AddDevice(ctx context.Context, device UserDevice) error
		SaveSession(ctx context.Context, session Session) error
		GetUserDeviceID(ctx context.Context, userID, userAgent string) (string, error)
		UpdateLatestLoginAt(ctx context.Context, deviceID string, latestLoginAt time.Time) error
		GetUserByEmail(ctx context.Context, email string) (User, error)
		GetUserData(ctx context.Context, userID string) (User, error)
		GetSessionByRefreshToken(ctx context.Context, refreshToken string) (Session, error)
		RemoveSession(ctx context.Context, userID, deviceID string) error
		GetUserByID(ctx context.Context, userID string) (User, error)
		UpdateUser(ctx context.Context, user User) error
		DeleteUser(ctx context.Context, user User) error
	}
)
