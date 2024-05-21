package usecase

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	ssogrpc "github.com/rshelekhov/reframed/internal/clients/sso/grpc"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	ssov1 "github.com/rshelekhov/sso-protos/gen/go/sso"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type AuthUsecase struct {
	ssoClient      *ssogrpc.Client
	jwt            *jwtoken.TokenService
	listUsecase    port.ListUsecase
	headingUsecase port.HeadingUsecase
}

func NewAuthUsecase(
	ssoClient *ssogrpc.Client,
	jwt *jwtoken.TokenService,
	listUsecase port.ListUsecase,
	headingUsecase port.HeadingUsecase,
) *AuthUsecase {
	return &AuthUsecase{
		ssoClient:      ssoClient,
		jwt:            jwt,
		listUsecase:    listUsecase,
		headingUsecase: headingUsecase,
	}
}

func (u *AuthUsecase) RegisterNewUser(
	ctx context.Context,
	userData *model.UserRequestData,
	userDevice model.UserDeviceRequestData,
) (
	tokenData *ssov1.TokenData,
	userID string,
	err error,
) {
	const op = "usecase.AuthUsecase.RegisterNewUser"

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
		st, ok := status.FromError(err)
		if !ok {
			return nil, "", err
		}

		switch st.Code() {
		case codes.AlreadyExists:
			return nil, "", le.ErrUserAlreadyExists
		case codes.Unauthenticated:
			return nil, "", le.ErrAppIDDoesNotExists
		default:
			return nil, "", err
		}
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
	ctx, err := jwtoken.AddAccessTokenToMetadata(ctx)
	if err != nil {
		return err
	}

	_, err = u.ssoClient.Api.Logout(ctx, &ssov1.LogoutRequest{
		AppId: u.jwt.AppID,
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

func (u *AuthUsecase) GetUserByID(ctx context.Context) (model.UserResponseData, error) {
	ctx, err := jwtoken.AddAccessTokenToMetadata(ctx)
	if err != nil {
		return model.UserResponseData{}, err
	}

	user, err := u.ssoClient.Api.GetUser(ctx, &ssov1.GetUserRequest{
		AppId: u.jwt.AppID,
	})
	if err != nil {
		return model.UserResponseData{}, err
	}

	userResponse := model.UserResponseData{
		Email:     user.GetEmail(),
		UpdatedAt: user.GetUpdatedAt().AsTime(),
	}

	return userResponse, err
}

func (u *AuthUsecase) UpdateUser(ctx context.Context, data *model.UserRequestData) error {
	if _, err := u.ssoClient.Api.UpdateUser(ctx, &ssov1.UpdateUserRequest{
		Email:           data.Email,
		CurrentPassword: data.Password,
		UpdatedPassword: data.UpdatedPassword,
		AppId:           u.jwt.AppID,
	}); err != nil {
		return err
	}

	return nil
}

func (u *AuthUsecase) DeleteUser(ctx context.Context, data model.UserDeviceRequestData) error {
	if _, err := u.ssoClient.Api.DeleteUser(ctx, &ssov1.DeleteUserRequest{
		AppId: u.jwt.AppID,
		UserDeviceData: &ssov1.UserDeviceData{
			UserAgent: data.UserAgent,
			Ip:        data.IP,
		},
	}); err != nil {
		return err
	}

	return nil
}
