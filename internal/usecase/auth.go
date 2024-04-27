package usecase

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	ssogrpc "github.com/rshelekhov/reframed/internal/clients/sso/grpc"
	ssov1 "github.com/rshelekhov/sso-protos/gen/go/sso"
	"time"

	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type AuthUsecase struct {
	authStorage    port.AuthStorage
	ssoClient      *ssogrpc.Client
	jwt            *jwtoken.TokenService
	listUsecase    port.ListUsecase
	headingUsecase port.HeadingUsecase
}

func NewAuthUsecase(
	storage port.AuthStorage,
	ssoClient *ssogrpc.Client,
	jwt *jwtoken.TokenService,
	listUsecase port.ListUsecase,
	headingUsecase port.HeadingUsecase,
) *AuthUsecase {
	return &AuthUsecase{
		authStorage:    storage,
		ssoClient:      ssoClient,
		jwt:            jwt,
		listUsecase:    listUsecase,
		headingUsecase: headingUsecase,
	}
}

func (u *AuthUsecase) CreateUser(
	ctx context.Context,
	userData *model.UserRequestData,
	userDevice model.UserDeviceRequestData,
) (
	tokenData *ssov1.TokenData,
	userID string,
	err error,
) {
	const op = "usecase.AuthUsecase.CreateUser"

	resp, err := u.ssoClient.Api.Register(ctx, &ssov1.RegisterRequest{
		Email:    userData.Email,
		Password: userData.Password,
		AppId:    userData.AppID,
		UserDeviceData: &ssov1.UserDeviceData{
			UserAgent: userDevice.UserAgent,
			Ip:        userDevice.IP,
		},
	})
	if err != nil {
		return nil, "", err
	}

	tokenData = resp.GetTokenData()
	if tokenData == nil {
		return nil, "", le.ErrFailedToGetTokenData
	}

	tokenParsed, err := u.jwt.ParseToken(ctx, tokenData.GetAccessToken())
	if err != nil {
		return nil, "", err
	}

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, "", le.ErrFailedGoGetClaimsFromToken
	}

	userID = claims[key.UserID].(string)

	return tokenData, userID, nil
}

func (u *AuthUsecase) LoginUser(
	ctx context.Context,
	userData *model.UserRequestData,
	userDevice model.UserDeviceRequestData,
) (
	tokenData *ssov1.TokenData,
	userID string,
	err error,
) {
	resp, err := u.ssoClient.Api.Login(ctx, &ssov1.LoginRequest{
		Email:    userData.Email,
		Password: userData.Password,
		AppId:    userData.AppID,
		UserDeviceData: &ssov1.UserDeviceData{
			UserAgent: userDevice.UserAgent,
			Ip:        userDevice.IP,
		},
	})
	if err != nil {
		return nil, "", err
	}

	tokenData = resp.GetTokenData()
	if tokenData == nil {
		return nil, "", le.ErrFailedToGetTokenData
	}

	tokenParsed, err := u.jwt.ParseToken(ctx, tokenData.GetAccessToken())
	if err != nil {
		return nil, "", err
	}

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, "", le.ErrFailedGoGetClaimsFromToken
	}

	userID = claims[key.UserID].(string)

	return tokenData, userID, nil
}

func (u *AuthUsecase) Refresh(
	ctx context.Context,
	refreshToken string,
	data model.UserDeviceRequestData,
) (
	tokenData *ssov1.TokenData,
	userID string,
	err error,
) {

	resp, err := u.ssoClient.Api.Refresh(ctx, &ssov1.RefreshRequest{
		RefreshToken: refreshToken,
		AppId:        u.jwt.AppID,
		UserDeviceData: &ssov1.UserDeviceData{
			UserAgent: data.UserAgent,
			Ip:        data.IP,
		},
	})
	if err != nil {
		return nil, "", err
	}

	tokenData = resp.GetTokenData()
	if tokenData == nil {
		return nil, "", le.ErrFailedToGetTokenData
	}

	tokenParsed, err := u.jwt.ParseToken(ctx, tokenData.GetAccessToken())
	if err != nil {
		return nil, "", err
	}

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, "", le.ErrFailedGoGetClaimsFromToken
	}

	userID = claims[key.UserID].(string)

	return tokenData, userID, nil
}

func (u *AuthUsecase) LogoutUser(ctx context.Context, data model.UserDeviceRequestData) error {
	_, err := u.ssoClient.Api.Logout(ctx, &ssov1.LogoutRequest{
		UserDeviceData: &ssov1.UserDeviceData{
			UserAgent: data.UserAgent,
			Ip:        data.IP,
		},
	})
	if err != nil {
		return err
	}
	
	return nil
}

func (u *AuthUsecase) GetUserByID(ctx context.Context, id string) (model.UserResponseData, error) {
	user, err := u.authStorage.GetUserByID(ctx, id)
	if err != nil {
		return model.UserResponseData{}, err
	}

	userResponse := model.UserResponseData{
		ID:        user.ID,
		Email:     user.Email,
		UpdatedAt: user.UpdatedAt,
	}

	return userResponse, err
}

func (u *AuthUsecase) UpdateUser(ctx context.Context, jwt *jwtoken.TokenService, data *model.UserRequestData, userID string) error {
	const op = "usecase.UserUsecase.UpdateUser"

	currentUser, err := u.authStorage.GetUserData(ctx, userID)
	if err != nil {
		return err
	}

	hash, err := jwtoken.PasswordHashBcrypt(
		data.Password,
		jwt.PasswordHashCost,
		[]byte(jwt.PasswordHashSalt),
	)
	if err != nil {
		return fmt.Errorf("%s: failed to generate password hash: %w", op, err)
	}

	updatedUser := model.User{
		ID:           userID,
		Email:        data.Email,
		PasswordHash: hash,
		UpdatedAt:    time.Now(),
	}

	emailChanged := updatedUser.Email != "" && updatedUser.Email != currentUser.Email
	passwordChanged := updatedUser.PasswordHash != ""

	if !emailChanged && !passwordChanged {
		return le.ErrNoChangesDetected
	}

	if err = u.authStorage.CheckEmailUniqueness(ctx, updatedUser); err != nil {
		return err
	}

	if data.Password != "" {
		if err = u.checkPassword(jwt, currentUser.PasswordHash, data.Password); err != nil {
			return err
		}
	}

	return u.authStorage.UpdateUser(ctx, updatedUser)
}

func (u *AuthUsecase) DeleteUser(ctx context.Context, userID string, data model.UserDeviceRequestData) error {
	deviceID, err := u.authStorage.GetUserDeviceID(ctx, userID, data.UserAgent)
	if err != nil {
		return err
	}

	deletedUser := model.User{
		ID:        userID,
		DeletedAt: time.Now(),
	}

	err = u.authStorage.DeleteUser(ctx, deletedUser)
	if err != nil {
		return err
	}

	err = u.authStorage.DeleteSession(ctx, userID, deviceID)
	if err != nil {
		return err
	}

	return nil
}
