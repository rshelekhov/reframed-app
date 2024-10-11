package usecase

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rshelekhov/jwtauth"
	ssogrpc "github.com/rshelekhov/reframed/internal/clients/sso/grpc"
	"github.com/rshelekhov/reframed/internal/config"
	ssov1 "github.com/rshelekhov/sso-protos/gen/go/sso"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type AuthUsecase struct {
	cfg            *config.ServerSettings
	ssoClient      *ssogrpc.Client
	jwt            *jwtauth.TokenService
	UserUsecase    port.UserUsecase
	ListUsecase    port.ListUsecase
	HeadingUsecase port.HeadingUsecase
}

func NewAuthUsecase(
	cfg *config.ServerSettings,
	ssoClient *ssogrpc.Client,
	jwt *jwtauth.TokenService,
) *AuthUsecase {
	return &AuthUsecase{
		cfg:       cfg,
		ssoClient: ssoClient,
		jwt:       jwt,
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
	verificationURL := u.cfg.AppData.BaseURL + "/verify-email/?token="

	resp, err := u.ssoClient.Api.RegisterUser(ctx, &ssov1.RegisterUserRequest{
		Email:           userData.Email,
		Password:        userData.Password,
		AppID:           u.jwt.AppID,
		VerificationURL: verificationURL,
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

	if err = u.ListUsecase.CreateDefaultList(ctx, userID); err != nil {
		return nil, "", err
	}

	return tokenData, userID, nil
}

func (u *AuthUsecase) VerifyEmail(ctx context.Context, verificationToken string) error {
	_, err := u.ssoClient.Api.VerifyEmail(ctx, &ssov1.VerifyEmailRequest{
		VerificationToken: verificationToken,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return err
		}

		switch st.Code() {
		case codes.FailedPrecondition:
			return le.ErrEmailVerificationTokenExpiredWithEmailResent
		case codes.NotFound:
			return le.ErrEmailVerificationTokenNotFound
		default:
			return err
		}
	}
	return nil
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
		AppID:    u.jwt.AppID,
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
		case codes.NotFound:
			return nil, "", le.ErrUserNotFound
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

func (u *AuthUsecase) RequestResetPassword(ctx context.Context, email string) error {
	confirmChangePasswordURL := u.cfg.AppData.BaseURL + "/password/change?token="

	_, err := u.ssoClient.Api.ResetPassword(ctx, &ssov1.ResetPasswordRequest{
		Email:                    email,
		AppID:                    u.jwt.AppID,
		ConfirmChangePasswordURL: confirmChangePasswordURL,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return err
		}

		switch st.Code() {
		case codes.Unauthenticated:
			return le.ErrAppIDDoesNotExists
		case codes.NotFound:
			return le.ErrUserNotFound
		default:
			return err
		}
	}
	return nil
}

func (u *AuthUsecase) ChangePassword(ctx context.Context, password, resetPasswordToken string) error {
	_, err := u.ssoClient.Api.ChangePassword(ctx, &ssov1.ChangePasswordRequest{
		ResetPasswordToken: resetPasswordToken,
		AppID:              u.jwt.AppID,
		UpdatedPassword:    password,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return err
		}

		switch st.Code() {
		case codes.FailedPrecondition:
			return le.ErrResetPasswordTokenExpiredWithEmailResent
		case codes.InvalidArgument:
			return le.ErrUpdatedPasswordMustNotMatchTheCurrent
		default:
			return err
		}
	}
	return nil
}

func (u *AuthUsecase) RefreshTokens(
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
		AppID:        u.jwt.AppID,
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
	ctx, err := jwtauth.AddAccessTokenToMetadata(ctx)
	if err != nil {
		return err
	}

	_, err = u.ssoClient.Api.Logout(ctx, &ssov1.LogoutRequest{
		AppID: u.jwt.AppID,
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
	ctx, err := jwtauth.AddAccessTokenToMetadata(ctx)
	if err != nil {
		return model.UserResponseData{}, err
	}

	user, err := u.ssoClient.Api.GetUser(ctx, &ssov1.GetUserRequest{
		AppID: u.jwt.AppID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return model.UserResponseData{}, err
		}

		if st.Code() == codes.NotFound {
			return model.UserResponseData{}, le.ErrUserNotFound
		}
	}

	userResponse := model.UserResponseData{
		Email:     user.GetEmail(),
		UpdatedAt: user.GetUpdatedAt().AsTime(),
	}

	return userResponse, err
}

func (u *AuthUsecase) UpdateUser(ctx context.Context, data *model.UpdateUserRequestData) error {
	ctx, err := jwtauth.AddAccessTokenToMetadata(ctx)
	if err != nil {
		return err
	}

	_, err = u.ssoClient.Api.UpdateUser(ctx, &ssov1.UpdateUserRequest{
		Email:           data.Email,
		CurrentPassword: data.Password,
		UpdatedPassword: data.UpdatedPassword,
		AppID:           u.jwt.AppID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return err
		}

		switch st.Code() {
		case codes.NotFound:
			return le.ErrUserNotFound
		case codes.AlreadyExists:
			return le.ErrEmailAlreadyTaken
		case codes.InvalidArgument:
			return fmt.Errorf("%w: %s", le.ErrBadRequest, st.Message())
		default:
			return err
		}
	}

	return nil
}

func (u *AuthUsecase) DeleteUser(ctx context.Context, userID string) error {
	ctx, err := jwtauth.AddAccessTokenToMetadata(ctx)
	if err != nil {
		return err
	}

	_, err = u.ssoClient.Api.DeleteUser(ctx, &ssov1.DeleteUserRequest{
		AppID: u.jwt.AppID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return err
		}

		switch st.Code() {
		case codes.NotFound:
			return le.ErrUserNotFound
		default:
			return err
		}
	}

	if err = u.UserUsecase.DeleteUserRelatedData(ctx, userID); err != nil {
		return err
	}

	return nil
}
