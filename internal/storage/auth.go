package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/pkg/constants/le"
	"strconv"
	"time"
)

type AuthStorage struct {
	*pgxpool.Pool
}

func NewAuthStorage(pool *pgxpool.Pool) *AuthStorage {
	return &AuthStorage{Pool: pool}
}

// CreateUser creates a new user
func (s *AuthStorage) CreateUser(ctx context.Context, user model.User) error {
	const op = "user.storage.CreateUser"

	tx, err := BeginTransaction(s.Pool, ctx, op)
	defer func() {
		RollbackOnError(&err, tx, ctx, op)
	}()

	userStatus, err := getUserStatus(ctx, tx, user.Email)
	if err != nil {
		return err
	}

	switch userStatus {
	case "active":
		return le.ErrUserAlreadyExists
	case "soft_deleted":
		if err = replaceSoftDeletedUser(ctx, tx, user); err != nil {
			return err
		}
	case "not_found":
		if err = insertUser(ctx, tx, user); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: unknown user status: %s", op, userStatus)
	}

	CommitTransaction(&err, tx, ctx, op)

	return nil
}

// getUserStatus returns the status of the user with the given email
func getUserStatus(ctx context.Context, tx pgx.Tx, email string) (string, error) {
	const (
		op = "user.storage.getUserStatus"

		query = `
			SELECT CASE
			WHEN EXISTS(
				SELECT 1
				FROM users
				WHERE email = $1
				  AND deleted_at IS NULL FOR UPDATE
			) THEN 'active'
			WHEN EXISTS(
				SELECT 1
				FROM users
				WHERE email = $1
				  AND deleted_at IS NOT NULL FOR UPDATE
			) THEN 'soft_deleted'
			ELSE 'not_found' END AS status`
	)

	var status string

	err := tx.QueryRow(ctx, query, email).Scan(&status)
	if err != nil {
		RollbackOnError(&err, tx, ctx, op)
		return "", fmt.Errorf("%s: failed to check if user exists: %w", op, err)
	}

	return status, nil
}

// replaceSoftDeletedUser replaces a soft deleted user with the given user
func replaceSoftDeletedUser(ctx context.Context, tx pgx.Tx, user model.User) error {
	const (
		op = "user.storage.replaceSoftDeletedUser"

		querySetDeletedAtNull = `
			UPDATE users
			SET deleted_at = NULL
			WHERE email = $1`

		queryInsertUser = `
			INSERT INTO users (id, email, password_hash, updated_at)
			VALUES ($1, $2, $3, $4)`
	)

	_, err := tx.Exec(ctx, querySetDeletedAtNull, user.Email)
	if err != nil {
		RollbackOnError(&err, tx, ctx, op)
		return fmt.Errorf("%s: failed to set deleted_at to NULL: %w", op, err)
	}

	_, err = tx.Exec(
		ctx,
		queryInsertUser,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.UpdatedAt,
	)
	if err != nil {
		RollbackOnError(&err, tx, ctx, op)
		return fmt.Errorf("%s: failed to replace soft deleted user: %w", op, err)
	}

	return nil
}

// insertUser inserts a new user
func insertUser(ctx context.Context, tx pgx.Tx, user model.User) error {
	const (
		op = "user.storage.insertNewUser"

		query = `
			INSERT INTO users (id, email, password_hash, updated_at)
			VALUES ($1, $2, $3, $4)`
	)

	_, err := tx.Exec(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.UpdatedAt,
	)
	if err != nil {
		RollbackOnError(&err, tx, ctx, op)
		return fmt.Errorf("%s: failed to insert new user: %w", op, err)
	}

	return nil
}

func (s *AuthStorage) AddDevice(ctx context.Context, device model.UserDevice) error {
	const (
		op = "user.storage.AddDevice"

		query = `
			INSERT INTO user_devices (id, user_id, user_agent, ip, detached, latest_login_at)
			VALUES ($1, $2, $3, $4, $5, $6)`
	)

	_, err := s.Exec(
		ctx,
		query,
		device.ID,
		device.UserID,
		device.UserAgent,
		device.IP,
		device.LatestLoginAt,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to add device: %w", op, err)
	}

	return nil
}

func (s *AuthStorage) SaveSession(ctx context.Context, session model.Session) error {
	// TODO: add constraint that user can have only active sessions for 5 devices
	const (
		op = "user.storage.SaveSession"

		query = `
			INSERT INTO refresh_sessions (user_id, device_id, refresh_token, last_visit_at, expires_at)
			VALUES ($1, $2, $3, $4, $5)`
	)

	_, err := s.Exec(
		ctx,
		query,
		session.UserID,
		session.DeviceID,
		session.RefreshToken,
		session.LastVisitAt,
		session.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to create session: %w", op, err)
	}
	return nil
}

func (s *AuthStorage) GetUserDeviceID(ctx context.Context, userID, userAgent string) (string, error) {
	const (
		op    = "user.storage.GetUserDeviceID"
		query = `
			SELECT id
			FROM user_devices
			WHERE user_id = $1
			  AND user_agent = $2
			  AND detached = false`
	)

	var deviceID string

	err := s.QueryRow(ctx, query, userID, userAgent).Scan(&deviceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", le.ErrUserDeviceNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: failed to get id of user device: %w", op, err)
	}

	return deviceID, nil

}

func (s *AuthStorage) UpdateLatestLoginAt(ctx context.Context, deviceID string, latestLoginAt time.Time) error {
	const (
		op = "user.storage.UpdateLatestLoginAt"

		query = `
			UPDATE user_devices
			SET latest_login_at = $1
			WHERE id = $2`
	)

	_, err := s.Exec(ctx, query, latestLoginAt, deviceID)
	if err != nil {
		return fmt.Errorf("%s: failed to update latest login at: %w", op, err)
	}

	return nil
}

func (s *AuthStorage) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	const (
		op = "user.storage.GetUserCredentials"

		query = `
			SELECT id, email, updated_at
			FROM users
			WHERE email = $1
			  AND deleted_at IS NULL`
	)

	var userDB model.User
	err := s.QueryRow(ctx, query, email).Scan(
		&userDB.ID,
		&userDB.Email,
		&userDB.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, le.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("%s: failed to get user credentials: %w", op, err)
	}

	return userDB, nil
}

func (s *AuthStorage) GetUserData(ctx context.Context, userID string) (model.User, error) {
	const (
		op = "user.storage.GetUserData"

		query = `
			SELECT id, email, password_hash, updated_at
			FROM users
			WHERE id = $1
			  AND deleted_at IS NULL`
	)

	var userDB model.User
	err := s.QueryRow(ctx, query, userID).Scan(
		&userDB.ID,
		&userDB.Email,
		&userDB.PasswordHash,
		&userDB.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, le.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("%s: failed to get user credentials: %w", op, err)
	}

	return userDB, nil
}

func (s *AuthStorage) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (model.Session, error) {
	const (
		op = "user.storage.GetSessionByRefreshToken"

		querySelect = `
			SELECT user_id, device_id, last_visit_at, expires_at
			FROM refresh_sessions
            WHERE refresh_token = $1`

		queryDelete = `
			DELETE FROM refresh_sessions
			WHERE refresh_token = $1`
	)

	var session model.Session
	session.RefreshToken = refreshToken

	err := s.QueryRow(ctx, querySelect, refreshToken).Scan(
		&session.UserID,
		&session.DeviceID,
		&session.LastVisitAt,
		&session.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return session, le.ErrSessionNotFound
	}
	if err != nil {
		return model.Session{}, fmt.Errorf("%s: failed to get session: %w", op, err)
	}

	_, err = s.Exec(ctx, queryDelete, refreshToken)
	if err != nil {
		return model.Session{}, fmt.Errorf("%s: failed to delete expired session: %w", op, err)
	}

	return session, nil
}

func (s *AuthStorage) RemoveSession(ctx context.Context, userID, deviceID string) error {
	const (
		op = "user.storage.RemoveSession"

		query = `
			DELETE FROM refresh_sessions
			WHERE user_id = $1
			  AND device_id = $2`
	)

	_, err := s.Exec(ctx, query, userID, deviceID)
	if err != nil {
		return fmt.Errorf("%s: failed to remove session: %w", op, err)
	}

	return nil
}

func (s *AuthStorage) GetUserByID(ctx context.Context, userID string) (model.User, error) {
	const (
		op = "user.storage.GetUserData"

		query = `
			SELECT id, email, updated_at
			FROM users
			WHERE id = $1
			  AND deleted_at IS NULL`
	)

	var user model.User

	err := s.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, le.ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("%s: failed to get user: %w", op, err)
	}

	return user, nil
}

// UpdateUser updates a user by ID
func (s *AuthStorage) UpdateUser(ctx context.Context, updatedUserData model.User) error {
	const op = "UpdateUser.storage.UpdateUser"

	// Begin transaction
	tx, err := BeginTransaction(s.Pool, ctx, op)
	defer func() {
		RollbackOnError(&err, tx, ctx, op)
	}()

	// Check if the updatedUserData exists
	currentUserData, err := s.GetUserByID(ctx, updatedUserData.ID)
	if err != nil {
		return err
	}

	emailChanged := updatedUserData.Email != "" && updatedUserData.Email != currentUserData.Email
	passwordChanged := updatedUserData.PasswordHash != ""

	if !emailChanged && !passwordChanged {
		return le.ErrNoChangesDetected
	}

	// Check if the updatedUserData email exists for a different updatedUserData
	if err = checkEmailUniqueness(ctx, tx, updatedUserData.Email, updatedUserData.ID); err != nil {
		return err
	}

	// Prepare the dynamic update query based on the provided fields
	queryUpdate := "UPDATE users SET updated_at = $1"
	queryParams := []interface{}{updatedUserData.UpdatedAt}

	if updatedUserData.Email != "" {
		queryUpdate += ", email = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, updatedUserData.Email)
	}
	if updatedUserData.PasswordHash != "" {
		queryUpdate += ", password_hash = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, updatedUserData.PasswordHash)
	}

	// Add condition for the specific updatedUserData ID
	queryUpdate += " WHERE id = $" + strconv.Itoa(len(queryParams)+1)
	queryParams = append(queryParams, updatedUserData.ID)

	// Execute the update query
	result, err := tx.Exec(ctx, queryUpdate, queryParams...)
	if err != nil {
		return fmt.Errorf("%s: failed to execute update query: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return le.ErrUserNotFound
	}

	CommitTransaction(&err, tx, ctx, op)

	return nil
}

// checkEmailUniqueness checks if the provided email already exists in the database for another user
func checkEmailUniqueness(ctx context.Context, tx pgx.Tx, email, id string) error {
	const (
		op = "user.storage.checkEmailUniqueness"

		query = `
			SELECT id
			FROM users
			WHERE email = $1
			  AND deleted_at IS NULL`
	)

	var existingUserID string
	err := tx.QueryRow(ctx, query, email).Scan(&existingUserID)

	if !errors.Is(err, pgx.ErrNoRows) && existingUserID != id {
		return le.ErrEmailAlreadyTaken
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: failed to check email uniqueness: %w", op, err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (s *AuthStorage) DeleteUser(ctx context.Context, user model.User) error {
	const (
		op = "user.storage.DeleteUser"

		query = `
			UPDATE users
			SET deleted_at = $1
			WHERE id = $2
			  AND deleted_at IS NULL`
	)

	result, err := s.Exec(ctx, query, user.DeletedAt, user.ID)
	if err != nil {
		return fmt.Errorf("%s: failed to delete user: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return le.ErrUserNotFound
	}

	return nil
}
