package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/rshelekhov/reframed/internal/domain"
	"github.com/rshelekhov/reframed/pkg/constants/key"
	"github.com/rshelekhov/reframed/pkg/httpserver/middleware/jwtoken"
	"github.com/segmentio/ksuid"
	"time"
)

type AuthUsecase struct {
	authStorage domain.AuthStorage
	listUsecase domain.ListUsecase
}

func NewAuthUsecase(storage domain.AuthStorage, listUsecase domain.ListUsecase) *AuthUsecase {
	return &AuthUsecase{
		authStorage: storage,
		listUsecase: listUsecase,
	}
}

func (u *AuthUsecase) CreateUser(ctx context.Context, jwt *jwtoken.TokenService, data *domain.UserRequestData) (string, error) {
	const op = "usecase.AuthUsecase.CreateUser"

	hash, err := jwtoken.PasswordHashBcrypt(
		data.Password,
		jwt.PasswordHashCost,
		[]byte(jwt.PasswordHashSalt),
	)
	if err != nil {
		return "", fmt.Errorf("%s: failed to generate password hash: %w", op, err)
	}

	updatedAt := time.Now().UTC()

	user := domain.User{
		ID:           ksuid.New().String(),
		Email:        data.Email,
		PasswordHash: hash,
		UpdatedAt:    &updatedAt,
	}

	// Create the user
	err = u.authStorage.CreateUser(ctx, user)
	if err != nil {
		return "", err
	}

	defaultList := domain.List{
		ID:        ksuid.New().String(),
		Title:     domain.DefaultInboxList.String(),
		UserID:    user.ID,
		UpdatedAt: &updatedAt,
	}

	err = u.listUsecase.CreateDefaultList(ctx, defaultList)
	if err != nil {
		return "", err
	}

	return user.ID, nil
}

// TODO: Move sessions from Postgres to Redis
func (u *AuthUsecase) CreateUserSession(ctx context.Context, jwt *jwtoken.TokenService, userID string, data domain.UserDeviceRequestData) (jwtoken.TokenData, error) {
	additionalClaims := map[string]interface{}{
		jwtoken.ContextUserID: userID,
	}

	deviceID, err := u.getDeviceID(ctx, userID, data)
	if err != nil {
		return jwtoken.TokenData{}, err
	}

	accessToken, err := jwt.NewAccessToken(additionalClaims)
	if err != nil {
		return jwtoken.TokenData{}, err
	}

	refreshToken, err := jwt.NewRefreshToken()
	if err != nil {
		return jwtoken.TokenData{}, err
	}

	lastVisitAt := time.Now().UTC()

	expiresAt := time.Now().Add(jwt.RefreshTokenTTL)

	session := domain.Session{
		UserID:       userID,
		DeviceID:     deviceID,
		RefreshToken: refreshToken,
		LastVisitAt:  &lastVisitAt,
		ExpiresAt:    &expiresAt,
	}

	if err = u.authStorage.SaveSession(ctx, session); err != nil {
		return jwtoken.TokenData{}, err
	}

	additionalFields := map[string]string{key.UserID: userID}
	tokenData := jwtoken.TokenData{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		Domain:           jwt.RefreshTokenCookieDomain,
		Path:             jwt.RefreshTokenCookiePath,
		ExpiresAt:        expiresAt,
		HttpOnly:         true,
		AdditionalFields: additionalFields,
	}

	return tokenData, nil
}

func (u *AuthUsecase) getDeviceID(ctx context.Context, userID string, data domain.UserDeviceRequestData) (string, error) {
	deviceID, err := u.authStorage.GetUserDeviceID(ctx, userID, data.UserAgent)
	if errors.Is(err, domain.ErrUserDeviceNotFound) {
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

func (u *AuthUsecase) registerDevice(ctx context.Context, userID string, data domain.UserDeviceRequestData) (string, error) {
	latestLoginAt := time.Now().UTC()

	userDevice := domain.UserDevice{
		ID:            ksuid.New().String(),
		UserID:        userID,
		UserAgent:     data.UserAgent,
		IP:            data.IP,
		Detached:      false,
		DetachedAt:    nil,
		LatestLoginAt: &latestLoginAt,
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

func (u *AuthUsecase) LoginUser(ctx context.Context, jwt *jwtoken.TokenService, data *domain.UserRequestData) (string, error) {
	user, err := u.authStorage.GetUserByEmail(ctx, data.Email)
	if err != nil {
		return "", err
	}

	if err = u.VerifyPassword(ctx, jwt, user, data.Password); err != nil {
		return "", err
	}

	return user.ID, nil
}

func (u *AuthUsecase) VerifyPassword(ctx context.Context, jwt *jwtoken.TokenService, user domain.User, password string) error {
	const op = "user.AuthUsecase.VerifyPassword"

	user, err := u.authStorage.GetUserData(ctx, user.ID)
	if err != nil {
		return err
	}
	if len(user.PasswordHash) == 0 {
		return domain.ErrUserHasNoPassword
	}

	matched, err := jwtoken.PasswordMatch(user.PasswordHash, password, []byte(jwt.PasswordHashSalt))
	if err != nil {
		return fmt.Errorf("%s: failed to check if password match: %w", op, err)
	}
	if !matched {
		return domain.ErrInvalidCredentials
	}
	return nil
}

func (u *AuthUsecase) CheckSessionAndDevice(ctx context.Context, refreshToken string, data domain.UserDeviceRequestData) (domain.Session, error) {
	// Get the session by refresh token
	session, err := u.authStorage.GetSessionByRefreshToken(ctx, refreshToken)
	if errors.Is(err, domain.ErrSessionNotFound) {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.Session{}, err
	}

	// Check if the session is expired
	if session.IsExpired() {
		return domain.Session{}, domain.ErrSessionExpired
	}

	// Check if the device exists
	_, err = u.authStorage.GetUserDeviceID(ctx, session.UserID, data.UserAgent)
	if errors.Is(err, domain.ErrUserDeviceNotFound) {
		return domain.Session{}, domain.ErrUserDeviceNotFound
	}
	if err != nil {
		return domain.Session{}, err
	}
	return session, nil
}

func (u *AuthUsecase) LogoutUser(ctx context.Context, userID string, data domain.UserDeviceRequestData) error {
	// Check if the device exists
	deviceID, err := u.authStorage.GetUserDeviceID(ctx, userID, data.UserAgent)
	if err != nil {
		return err
	}
	return u.authStorage.RemoveSession(ctx, userID, deviceID)
}

func (u *AuthUsecase) GetUserByID(ctx context.Context, id string) (domain.UserResponseData, error) {
	user, err := u.authStorage.GetUserByID(ctx, id)
	if err != nil {
		return domain.UserResponseData{}, err
	}

	userResponse := domain.UserResponseData{
		ID:        user.ID,
		Email:     user.Email,
		UpdatedAt: user.UpdatedAt,
	}

	return userResponse, err
}

func (u *AuthUsecase) UpdateUser(ctx context.Context, jwt *jwtoken.TokenService, data *domain.UserRequestData, userID string) error {

	user, err := u.authStorage.GetUserData(ctx, userID)
	if err != nil {
		return err
	}

	updatedAt := time.Now().UTC()

	updatedUser := domain.User{
		ID:        user.ID,
		Email:     data.Email,
		UpdatedAt: &updatedAt,
	}

	if data.Password != "" {
		if err = u.checkPassword(jwt, user.PasswordHash, data.Password); err != nil {
			return err
		}
	}

	return u.authStorage.UpdateUser(ctx, updatedUser)
}

func (u *AuthUsecase) checkPassword(jwt *jwtoken.TokenService, currentPasswordHash, passwordFromRequest string) error {
	const op = "usecase.UserUsecase.checkPassword"

	updatedPasswordHash, err := jwtoken.PasswordHashBcrypt(
		passwordFromRequest,
		jwt.PasswordHashCost,
		[]byte(jwt.PasswordHashSalt),
	)
	if err != nil {
		return fmt.Errorf("%s: failed to generate hash for the updated password: %w", op, err)
	}

	if updatedPasswordHash == currentPasswordHash {
		return domain.ErrNoPasswordChangesDetected
	}

	return nil
}

func (u *AuthUsecase) DeleteUser(ctx context.Context, userID string, data domain.UserDeviceRequestData) error {
	deviceID, err := u.authStorage.GetUserDeviceID(ctx, userID, data.UserAgent)
	if err != nil {
		return err
	}

	deletedAt := time.Now().UTC()

	deletedUser := domain.User{
		ID:        userID,
		DeletedAt: &deletedAt,
	}

	err = u.authStorage.DeleteUser(ctx, deletedUser)
	if err != nil {
		return err
	}

	err = u.authStorage.RemoveSession(ctx, userID, deviceID)
	if err != nil {
		return err
	}

	return nil
}
