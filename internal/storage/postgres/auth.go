package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
	"github.com/rshelekhov/reframed/internal/storage/postgres/sqlc"
)

type AuthStorage struct {
	*pgxpool.Pool
	*sqlc.Queries
}

func NewAuthStorage(pool *pgxpool.Pool) port.AuthStorage {
	return &AuthStorage{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}

func (s *AuthStorage) Transaction(ctx context.Context, fn func(storage port.AuthStorage) error) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				err = fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
			}
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = fn(s)

	return err
}

// CreateUser creates a new user
func (s *AuthStorage) CreateUser(ctx context.Context, user model.User) error {
	const op = "user.storage.CreateUser"

	userStatus, err := s.getUserStatus(ctx, user.Email)
	if err != nil {
		return err
	}

	switch userStatus {
	case "active":
		return le.ErrUserAlreadyExists
	case "soft_deleted":
		if err = s.replaceSoftDeletedUser(ctx, user); err != nil {
			return err
		}
	case "not_found":
		if err = s.insertUser(ctx, user); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: unknown user status: %s", op, userStatus)
	}

	return nil
}

// getUserStatus returns the status of the user with the given email
func (s *AuthStorage) getUserStatus(ctx context.Context, email string) (string, error) {
	const op = "user.storage.getUserStatus"

	status, err := s.Queries.GetUserStatus(ctx, email)
	if err != nil {
		return "", fmt.Errorf("%s: failed to check if user exists: %w", op, err)
	}

	return status, nil
}

// replaceSoftDeletedUser replaces a soft deleted user with the given user
func (s *AuthStorage) replaceSoftDeletedUser(ctx context.Context, user model.User) error {
	const op = "user.storage.replaceSoftDeletedUser"

	if err := s.Queries.SetDeletedUserAtNull(ctx, user.Email); err != nil {
		return fmt.Errorf("%s: failed to set deleted_at to NULL: %w", op, err)
	}

	if err := s.Queries.InsertUser(ctx, sqlc.InsertUserParams{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		UpdatedAt:    user.UpdatedAt,
	}); err != nil {
		return fmt.Errorf("%s: failed to replace soft deleted user: %w", op, err)
	}
	return nil
}

// insertUser inserts a new user
func (s *AuthStorage) insertUser(ctx context.Context, user model.User) error {
	const op = "user.storage.insertNewUser"

	if err := s.Queries.InsertUser(ctx, sqlc.InsertUserParams{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		UpdatedAt:    user.UpdatedAt,
	}); err != nil {
		return fmt.Errorf("%s: failed to insert new user: %w", op, err)
	}
	return nil
}

func (s *AuthStorage) UpdateLatestLoginAt(ctx context.Context, deviceID string, latestLoginAt time.Time) error {
	const op = "user.storage.UpdateLatestLoginAt"

	if err := s.Queries.UpdateLatestLoginAt(ctx, sqlc.UpdateLatestLoginAtParams{
		ID:            deviceID,
		LatestLoginAt: latestLoginAt,
	}); err != nil {
		return fmt.Errorf("%s: failed to update latest login at: %w", op, err)
	}
	return nil
}

func (s *AuthStorage) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	const op = "user.storage.GetUserCredentials"

	user, err := s.Queries.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, le.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("%s: failed to get user credentials: %w", op, err)
	}

	return model.User{
		ID:        user.ID,
		Email:     user.Email,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *AuthStorage) GetUserData(ctx context.Context, userID string) (model.User, error) {
	const op = "user.storage.GetUserData"

	user, err := s.Queries.GetUserData(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, le.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("%s: failed to get user credentials: %w", op, err)
	}

	return model.User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		UpdatedAt:    user.UpdatedAt,
	}, nil
}

func (s *AuthStorage) GetUserByID(ctx context.Context, userID string) (model.User, error) {
	const op = "user.storage.GetUserData"

	user, err := s.Queries.GetUserByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, le.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("%s: failed to get user: %w", op, err)
	}

	return model.User{
		ID:        user.ID,
		Email:     user.Email,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// UpdateUser updates a user by ID
func (s *AuthStorage) UpdateUser(ctx context.Context, user model.User) error {
	const op = "UpdateUser.storage.UpdateUser"

	// Prepare the dynamic update query based on the provided fields
	queryUpdate := "UPDATE users SET updated_at = $1"
	queryParams := []interface{}{user.UpdatedAt}

	if user.Email != "" {
		queryUpdate += ", email = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, user.Email)
	}

	if user.PasswordHash != "" {
		queryUpdate += ", password_hash = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, user.PasswordHash)
	}

	// Add condition for the specific user ID
	queryUpdate += " WHERE id = $" + strconv.Itoa(len(queryParams)+1)
	queryParams = append(queryParams, user.ID)

	// Execute the update query
	_, err := s.Exec(ctx, queryUpdate, queryParams...)
	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to execute update query: %w", op, err)
	}

	return nil
}

// CheckEmailUniqueness checks if the provided email already exists in the database for another user
func (s *AuthStorage) CheckEmailUniqueness(ctx context.Context, user model.User) error {
	const op = "user.storage.checkEmailUniqueness"

	existingUserID, err := s.Queries.GetUserID(ctx, user.Email)
	if !errors.Is(err, pgx.ErrNoRows) && existingUserID != user.ID {
		return le.ErrEmailAlreadyTaken
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: failed to check email uniqueness: %w", op, err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (s *AuthStorage) DeleteUser(ctx context.Context, user model.User) error {
	const op = "user.storage.DeleteUser"

	err := s.Queries.DeleteUser(ctx, sqlc.DeleteUserParams{
		ID: user.ID,
		DeletedAt: pgtype.Timestamptz{
			Time: user.DeletedAt,
		},
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to delete user: %w", op, err)
	}
	return nil
}

//
// Storage methods for user sessions
//

func (s *AuthStorage) AddDevice(ctx context.Context, device model.UserDevice) error {
	const op = "user.storage.AddDevice"

	if err := s.Queries.AddDevice(ctx, sqlc.AddDeviceParams{
		ID:            device.ID,
		UserID:        device.UserID,
		UserAgent:     device.UserAgent,
		Ip:            device.IP,
		Detached:      device.Detached,
		LatestLoginAt: device.LatestLoginAt,
	}); err != nil {
		return fmt.Errorf("%s: failed to add device: %w", op, err)
	}
	return nil
}

func (s *AuthStorage) GetUserDeviceID(ctx context.Context, userID, userAgent string) (string, error) {
	const op = "user.storage.GetUserDeviceID"

	deviceID, err := s.Queries.GetUserDeviceID(ctx, sqlc.GetUserDeviceIDParams{
		UserID:    userID,
		UserAgent: userAgent,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return "", le.ErrUserDeviceNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: failed to get id of user device: %w", op, err)
	}

	return deviceID, nil
}

func (s *AuthStorage) SaveSession(ctx context.Context, session model.Session) error {
	// TODO: add constraint that user can have only active sessions for 5 devices
	const op = "user.storage.SaveSession"

	if err := s.Queries.SaveSession(ctx, sqlc.SaveSessionParams{
		RefreshToken: session.RefreshToken,
		UserID:       session.UserID,
		DeviceID:     session.DeviceID,
		LastVisitAt:  session.LastVisitAt,
		ExpiresAt:    session.ExpiresAt,
	}); err != nil {
		return fmt.Errorf("%s: failed to save session: %w", op, err)
	}

	return nil
}

func (s *AuthStorage) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (model.Session, error) {
	const op = "user.storage.GetSessionByRefreshToken"

	session, err := s.Queries.GetSessionByRefreshToken(ctx, refreshToken)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Session{}, le.ErrSessionNotFound
	}
	if err != nil {
		return model.Session{}, fmt.Errorf("%s: failed to get session: %w", op, err)
	}

	return model.Session{
		UserID:       session.UserID,
		DeviceID:     session.DeviceID,
		RefreshToken: refreshToken,
		LastVisitAt:  session.LastVisitAt,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

func (s *AuthStorage) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	const op = "user.storage.DeleteRefreshToken"

	if err := s.Queries.DeleteRefreshTokenFromSession(ctx, refreshToken); err != nil {
		return fmt.Errorf("%s: failed to delete expired session: %w", op, err)
	}

	return nil
}

func (s *AuthStorage) DeleteSession(ctx context.Context, userID, deviceID string) error {
	const op = "user.storage.DeleteSession"

	err := s.Queries.DeleteSession(ctx, sqlc.DeleteSessionParams{
		UserID:   userID,
		DeviceID: deviceID,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return le.ErrSessionNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to remove session: %w", op, err)
	}

	return nil
}
