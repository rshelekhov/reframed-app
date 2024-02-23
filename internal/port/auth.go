package port

import (
	"context"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/pkg/httpserver/middleware/jwtoken"
	"time"
)

type (
	AuthUsecase interface {
		CreateUser(ctx context.Context, jwt *jwtoken.TokenService, data *model.UserRequestData) (string, error)
		CreateUserSession(ctx context.Context, jwt *jwtoken.TokenService, userID string, device model.UserDeviceRequestData) (jwtoken.TokenData, error)
		LoginUser(ctx context.Context, jwt *jwtoken.TokenService, data *model.UserRequestData) (string, error)
		CheckSessionAndDevice(ctx context.Context, refreshToken string, data model.UserDeviceRequestData) (model.Session, error)
		LogoutUser(ctx context.Context, userID string, data model.UserDeviceRequestData) error
		GetUserByID(ctx context.Context, id string) (model.UserResponseData, error)
		UpdateUser(ctx context.Context, jwt *jwtoken.TokenService, data *model.UserRequestData, userID string) error
		DeleteUser(ctx context.Context, userUD string, data model.UserDeviceRequestData) error
	}

	AuthStorage interface {
		CreateUser(ctx context.Context, user model.User) error
		AddDevice(ctx context.Context, device model.UserDevice) error
		SaveSession(ctx context.Context, session model.Session) error
		GetUserDeviceID(ctx context.Context, userID, userAgent string) (string, error)
		UpdateLatestLoginAt(ctx context.Context, deviceID string, latestLoginAt time.Time) error
		GetUserByEmail(ctx context.Context, email string) (model.User, error)
		GetUserData(ctx context.Context, userID string) (model.User, error)
		GetSessionByRefreshToken(ctx context.Context, refreshToken string) (model.Session, error)
		DeleteSession(ctx context.Context, userID, deviceID string) error
		GetUserByID(ctx context.Context, userID string) (model.User, error)
		CheckEmailUniqueness(ctx context.Context, user model.User) error
		UpdateUser(ctx context.Context, user model.User) error
		DeleteUser(ctx context.Context, user model.User) error
	}
)
