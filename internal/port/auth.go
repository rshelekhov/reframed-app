package port

import (
	"context"
	"github.com/rshelekhov/reframed/internal/model"
	ssov1 "github.com/rshelekhov/sso-protos/gen/go/sso"
)

type AuthUsecase interface {
	CreateUser(ctx context.Context, userData *model.UserRequestData, userDevice model.UserDeviceRequestData) (tokenData *ssov1.TokenData, userID string, err error)
	LoginUser(ctx context.Context, userData *model.UserRequestData, userDevice model.UserDeviceRequestData) (tokenData *ssov1.TokenData, userID string, err error)
	Refresh(ctx context.Context, refreshToken string, data model.UserDeviceRequestData) (tokenData *ssov1.TokenData, userID string, err error)
	LogoutUser(ctx context.Context, data model.UserDeviceRequestData) error
	GetUserByID(ctx context.Context) (model.UserResponseData, error)
	UpdateUser(ctx context.Context, data *model.UserRequestData) error
	DeleteUser(ctx context.Context, data model.UserDeviceRequestData) error
}
