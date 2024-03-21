package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	jwtoken2 "github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	"github.com/segmentio/ksuid"

	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type AuthUsecase struct {
	authStorage    port.AuthStorage
	listUsecase    port.ListUsecase
	headingUsecase port.HeadingUsecase
}

func NewAuthUsecase(
	storage port.AuthStorage,
	listUsecase port.ListUsecase,
	headingUsecase port.HeadingUsecase,
) *AuthUsecase {
	return &AuthUsecase{
		authStorage:    storage,
		listUsecase:    listUsecase,
		headingUsecase: headingUsecase,
	}
}

func (u *AuthUsecase) CreateUser(ctx context.Context, jwt *jwtoken2.TokenService, data *model.UserRequestData) (string, error) {
	const op = "usecase.AuthUsecase.CreateUser"

	hash, err := jwtoken2.PasswordHashBcrypt(
		data.Password,
		jwt.PasswordHashCost,
		[]byte(jwt.PasswordHashSalt),
	)
	if err != nil {
		return "", fmt.Errorf("%s: failed to generate password hash: %w", op, err)
	}

	user := model.User{
		ID:           ksuid.New().String(),
		Email:        data.Email,
		PasswordHash: hash,
		UpdatedAt:    time.Now(),
	}

	if err = u.authStorage.Transaction(ctx, func(_ port.AuthStorage) error {
		if err = u.authStorage.CreateUser(ctx, user); err != nil {
			return err
		}

		if err = u.listUsecase.CreateDefaultList(ctx, user.ID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	return user.ID, nil
}

// TODO: Move sessions from Postgres to Redis
func (u *AuthUsecase) CreateUserSession(
	ctx context.Context,
	jwt *jwtoken2.TokenService,
	userID string,
	data model.UserDeviceRequestData,
) (
	jwtoken2.TokenData,
	error,
) {
	additionalClaims := map[string]interface{}{
		jwtoken2.ContextUserID: userID,
	}

	deviceID, err := u.getDeviceID(ctx, userID, data)
	if err != nil {
		return jwtoken2.TokenData{}, err
	}

	accessToken, err := jwt.NewAccessToken(additionalClaims)
	if err != nil {
		return jwtoken2.TokenData{}, err
	}

	refreshToken, err := jwt.NewRefreshToken()
	if err != nil {
		return jwtoken2.TokenData{}, err
	}

	expiresAt := time.Now().Add(jwt.RefreshTokenTTL)

	session := model.Session{
		UserID:       userID,
		DeviceID:     deviceID,
		RefreshToken: refreshToken,
		LastVisitAt:  time.Now(),
		ExpiresAt:    expiresAt,
	}

	if err = u.authStorage.SaveSession(ctx, session); err != nil {
		return jwtoken2.TokenData{}, err
	}

	additionalFields := map[string]string{key.UserID: userID}
	tokenData := jwtoken2.TokenData{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		Domain:           jwt.RefreshTokenCookieDomain,
		Path:             jwt.RefreshTokenCookiePath,
		ExpiresAt:        expiresAt,
		HTTPOnly:         true,
		AdditionalFields: additionalFields,
	}

	return tokenData, nil
}

func (u *AuthUsecase) getDeviceID(ctx context.Context, userID string, data model.UserDeviceRequestData) (string, error) {
	deviceID, err := u.authStorage.GetUserDeviceID(ctx, userID, data.UserAgent)
	if errors.Is(err, le.ErrUserDeviceNotFound) {
		return u.registerDevice(ctx, userID, data)
	}
	if err != nil {
		return "", err
	}

	err = u.updateLatestLoginAt(ctx, deviceID)
	if err != nil {
		return "", err
	}

	return deviceID, nil
}

func (u *AuthUsecase) registerDevice(ctx context.Context, userID string, data model.UserDeviceRequestData) (string, error) {
	userDevice := model.UserDevice{
		ID:            ksuid.New().String(),
		UserID:        userID,
		UserAgent:     data.UserAgent,
		IP:            data.IP,
		Detached:      false,
		LatestLoginAt: time.Now(),
	}

	if err := u.authStorage.AddDevice(ctx, userDevice); err != nil {
		return "", err
	}

	return userDevice.ID, nil
}

func (u *AuthUsecase) updateLatestLoginAt(ctx context.Context, deviceID string) error {
	latestLoginAt := time.Now().UTC()
	return u.authStorage.UpdateLatestLoginAt(ctx, deviceID, latestLoginAt)
}

func (u *AuthUsecase) LoginUser(ctx context.Context, jwt *jwtoken2.TokenService, data *model.UserRequestData) (string, error) {
	user, err := u.authStorage.GetUserByEmail(ctx, data.Email)
	if err != nil {
		return "", err
	}

	if err = u.VerifyPassword(ctx, jwt, user, data.Password); err != nil {
		return "", err
	}

	return user.ID, nil
}

func (u *AuthUsecase) VerifyPassword(ctx context.Context, jwt *jwtoken2.TokenService, user model.User, password string) error {
	const op = "user.AuthUsecase.VerifyPassword"

	user, err := u.authStorage.GetUserData(ctx, user.ID)
	if err != nil {
		return err
	}

	if len(user.PasswordHash) == 0 {
		return le.ErrUserHasNoPassword
	}

	matched, err := jwtoken2.PasswordMatch(user.PasswordHash, password, []byte(jwt.PasswordHashSalt))
	if err != nil {
		return fmt.Errorf("%s: failed to check if password match: %w", op, err)
	}

	if !matched {
		return le.ErrInvalidCredentials
	}

	return nil
}

func (u *AuthUsecase) CheckSessionAndDevice(ctx context.Context, refreshToken string, data model.UserDeviceRequestData) (model.Session, error) {
	// Get the session by refresh token
	session, err := u.authStorage.GetSessionByRefreshToken(ctx, refreshToken)
	if errors.Is(err, le.ErrSessionNotFound) {
		return model.Session{}, le.ErrSessionNotFound
	}
	if err != nil {
		return model.Session{}, err
	}

	// Check if the session is expired
	if session.IsExpired() {
		return model.Session{}, le.ErrSessionExpired
	}

	// Check if the device exists
	_, err = u.authStorage.GetUserDeviceID(ctx, session.UserID, data.UserAgent)
	if errors.Is(err, le.ErrUserDeviceNotFound) {
		return model.Session{}, le.ErrUserDeviceNotFound
	}
	if err != nil {
		return model.Session{}, err
	}

	return session, nil
}

func (u *AuthUsecase) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	return u.authStorage.DeleteRefreshToken(ctx, refreshToken)
}

func (u *AuthUsecase) LogoutUser(ctx context.Context, userID string, data model.UserDeviceRequestData) error {
	// Check if the device exists
	deviceID, err := u.authStorage.GetUserDeviceID(ctx, userID, data.UserAgent)
	if err != nil {
		return err
	}

	return u.authStorage.DeleteSession(ctx, userID, deviceID)
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

func (u *AuthUsecase) UpdateUser(ctx context.Context, jwt *jwtoken2.TokenService, data *model.UserRequestData, userID string) error {
	const op = "usecase.UserUsecase.UpdateUser"

	currentUser, err := u.authStorage.GetUserData(ctx, userID)
	if err != nil {
		return err
	}

	hash, err := jwtoken2.PasswordHashBcrypt(
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

func (u *AuthUsecase) checkPassword(jwt *jwtoken2.TokenService, currentPasswordHash, passwordFromRequest string) error {
	const op = "usecase.UserUsecase.checkPassword"

	updatedPasswordHash, err := jwtoken2.PasswordHashBcrypt(
		passwordFromRequest,
		jwt.PasswordHashCost,
		[]byte(jwt.PasswordHashSalt),
	)
	if err != nil {
		return fmt.Errorf("%s: failed to generate hash for the updated password: %w", op, err)
	}

	if updatedPasswordHash == currentPasswordHash {
		return le.ErrNoPasswordChangesDetected
	}

	return nil
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
