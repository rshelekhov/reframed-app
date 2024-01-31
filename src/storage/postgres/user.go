package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/models"
	"strconv"
	"time"
)

type UserStorage struct {
	*pgxpool.Pool
}

func NewUserStorage(pool *pgxpool.Pool) *UserStorage {
	return &UserStorage{Pool: pool}
}

// CreateUser creates a new user
func (s *UserStorage) CreateUser(ctx context.Context, user models.User) error {
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
		return c.ErrUserAlreadyExists
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

		query = `SELECT CASE
						WHEN EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL FOR UPDATE) THEN 'active'
						WHEN EXISTS(SELECT 1 FROM users WHERE email = $1 and deleted_at IS NOT NULL FOR UPDATE) THEN 'soft_deleted'
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

// TODO: use select for update
// replaceSoftDeletedUser replaces a soft deleted user with the given user
func replaceSoftDeletedUser(ctx context.Context, tx pgx.Tx, user models.User) error {
	const (
		op                    = "user.storage.replaceSoftDeletedUser"
		querySetDeletedAtNull = `UPDATE users SET deleted_at = NULL WHERE email = $1`
		queryInsertUser       = `INSERT INTO users (id, email, password, updated_at) VALUES ($1, $2, $3, $4)`
	)

	_, err := tx.Exec(ctx, querySetDeletedAtNull, user.Email)
	if err != nil {
		RollbackOnError(&err, tx, ctx, op)
		return fmt.Errorf("%s: failed to set deleted_at to NULL: %w", op, err)
	}

	_, err = tx.Exec(ctx, queryInsertUser, user.ID, user.Email, user.Password, user.UpdatedAt)
	if err != nil {
		RollbackOnError(&err, tx, ctx, op)
		return fmt.Errorf("%s: failed to replace soft deleted user: %w", op, err)
	}

	return nil
}

// insertUser inserts a new user
func insertUser(ctx context.Context, tx pgx.Tx, user models.User) error {
	const (
		op = "user.storage.insertNewUser"

		query = `INSERT INTO users
    							(id, email, password, updated_at)
								VALUES ($1, $2, $3, $4)`
	)

	_, err := tx.Exec(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Password,
		user.UpdatedAt,
	)
	if err != nil {
		RollbackOnError(&err, tx, ctx, op)
		return fmt.Errorf("%s: failed to insert new user: %w", op, err)
	}

	return nil
}

func (s *UserStorage) GetUserCredentials(ctx context.Context, user *models.User) (models.User, error) {
	const (
		op = "user.storage.GetUserCredentials"

		query = `SELECT id, email, password, updated_at
					FROM users WHERE email = $1 AND password = $2 AND deleted_at IS NULL`
	)

	var userDB models.User
	err := s.QueryRow(ctx, query, user.Email, user.Password).Scan(
		&userDB.ID,
		&userDB.Email,
		&userDB.Password,
		&userDB.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return userDB, c.ErrUserNotFound
	}
	if err != nil {
		return userDB, fmt.Errorf("%s: failed to get user credentials: %w", op, err)
	}

	return userDB, nil
}

func (s *UserStorage) SaveSession(ctx context.Context, userID, deviceID string, session models.Session) error {
	// TODO: add constraint that user can have only active sessions for 5 devices
	const (
		op = "user.storage.SaveSession"

		query = `INSERT INTO refresh_sessions (user_id, device_id, refresh_token, last_visit_at, expires_at)
					VALUES ($1, $2, $3, $4, $5)`
	)

	_, err := s.Exec(ctx, query, userID, deviceID, session.RefreshToken, time.Now(), session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("%s: failed to create session: %w", op, err)
	}
	return nil
}

func (s *UserStorage) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (models.Session, error) {
	const (
		op = "user.storage.GetSessionByRefreshToken"

		querySelect = `SELECT user_id, device_id, last_visit_at, expires_at FROM refresh_sessions
               		WHERE refresh_token = $1`

		queryDelete = `DELETE FROM refresh_sessions WHERE refresh_token = $1`
	)

	var session models.Session
	session.RefreshToken = refreshToken

	err := s.QueryRow(ctx, querySelect, refreshToken).
		Scan(&session.UserID, &session.DeviceID, &session.LastVisitAt, &session.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return session, c.ErrSessionNotFound
	}
	if err != nil {
		return session, fmt.Errorf("%s: failed to get session: %w", op, err)
	}

	_, err = s.Exec(ctx, queryDelete, refreshToken)
	if err != nil {
		return session, fmt.Errorf("%s: failed to delete expired session: %w", op, err)
	}

	return session, nil
}

func (s *UserStorage) AddDevice(ctx context.Context, device models.UserDevice) error {
	const (
		op = "user.storage.AddDevice"

		query = `INSERT INTO user_devices (id, user_id, user_agent, ip, detached, latest_login_at)
					VALUES ($1, $2, $3, $4, $5, $6)`
	)

	_, err := s.Exec(
		ctx,
		query,
		device.ID,
		device.UserID,
		device.UserAgent,
		device.IP,
		device.Detached,
		device.LatestLoginAt,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to add device: %w", op, err)
	}

	return nil
}

func (s *UserStorage) GetUserDevice(ctx context.Context, userID, userAgent string) (models.UserDevice, error) {
	const (
		op = "user.storage.GetUserDevice"

		query = `SELECT id, user_id, user_agent, ip, detached, latest_login_at
							FROM user_devices
							WHERE user_id = $1 AND user_agent = $2 AND detached = false`
	)

	var device models.UserDevice
	err := s.QueryRow(ctx, query, userID, userAgent).Scan(
		&device.ID,
		&device.UserID,
		&device.UserAgent,
		&device.IP,
		&device.Detached,
		&device.LatestLoginAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return device, c.ErrUserDeviceNotFound
	}
	if err != nil {
		return device, fmt.Errorf("%s: failed to get user device: %w", op, err)
	}

	return device, nil
}

// GetUser returns a user by ID
func (s *UserStorage) GetUser(ctx context.Context, id string) (models.User, error) {
	const (
		op = "user.storage.GetUser"

		query = `SELECT id, email, updated_at
							FROM users WHERE id = $1 AND deleted_at IS NULL`
	)

	var user models.User

	err := s.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, c.ErrUserNotFound
	}
	if err != nil {
		return user, fmt.Errorf("%s: failed to get user: %w", op, err)
	}

	return user, nil
}

// GetUsers returns a list of users
func (s *UserStorage) GetUsers(ctx context.Context, pgn models.Pagination) ([]models.User, error) {
	const (
		op = "user.storage.GetUsers"

		query = `SELECT id, email, updated_at
							FROM users WHERE deleted_at IS NULL ORDER BY id DESC LIMIT $1 OFFSET $2`
	)

	rows, err := s.Query(ctx, query, pgn.Limit, pgn.Offset)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		user := models.User{}

		err = rows.Scan(&user.ID, &user.Email, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan row: %w", op, err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(users) == 0 {
		return nil, c.ErrNoUsersFound
	}

	return users, nil
}

// UpdateUser updates a user by ID
func (s *UserStorage) UpdateUser(ctx context.Context, user models.User) error {
	const op = "user.storage.UpdateUser"

	// Begin transaction
	tx, err := BeginTransaction(s.Pool, ctx, op)
	defer func() {
		RollbackOnError(&err, tx, ctx, op)
	}()

	// Check if the user exists
	currentUser, err := s.GetUser(ctx, user.ID)
	if err != nil {
		return err
	}

	// Get user password
	currentUserPassword, err := getUserPassword(ctx, tx, currentUser.ID)
	if err != nil {
		return err
	}

	emailChanged := user.Email != "" && user.Email != currentUser.Email
	passwordChanged := user.Password != ""

	if !emailChanged && !passwordChanged {
		return c.ErrNoChangesDetected
	}

	// Check if the user email exists for a different user
	if err = checkEmailUniqueness(ctx, tx, user.Email, user.ID); err != nil {
		return err
	}

	if passwordChanged && user.Password == currentUserPassword {
		return c.ErrNoPasswordChangesDetected
	}

	// Prepare the dynamic update query based on the provided fields
	queryUpdate := "UPDATE users SET updated_at = $1"
	queryParams := []interface{}{user.UpdatedAt}

	if user.Email != "" {
		queryUpdate += ", email = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, user.Email)
	}
	if user.Password != "" {
		queryUpdate += ", password = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, user.Password)
	}

	// Add condition for the specific user ID
	queryUpdate += " WHERE id = $" + strconv.Itoa(len(queryParams)+1)
	queryParams = append(queryParams, user.ID)

	// Execute the update query
	_, err = tx.Exec(ctx, queryUpdate, queryParams...)
	if err != nil {
		return fmt.Errorf("%s: failed to execute update query: %w", op, err)
	}

	CommitTransaction(&err, tx, ctx, op)

	return nil
}

// getUserPassword returns the password of the user with the given id
func getUserPassword(ctx context.Context, tx pgx.Tx, id string) (string, error) {
	const (
		op = "user.storage.getUserPassword"

		query = `SELECT password FROM users WHERE id = $1 AND deleted_at IS NULL`
	)

	var password string
	err := tx.QueryRow(ctx, query, id).Scan(&password)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", c.ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: failed to get user password: %w", op, err)
	}

	return password, nil
}

// checkEmailUniqueness checks if the provided email already exists in the database for another user
func checkEmailUniqueness(ctx context.Context, tx pgx.Tx, email, id string) error {
	const (
		op = "user.storage.checkEmailUniqueness"

		query = `SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`
	)

	var existingUserID string

	err := tx.QueryRow(ctx, query, email).Scan(&existingUserID)
	if !errors.Is(err, pgx.ErrNoRows) && existingUserID != id {
		return c.ErrEmailAlreadyTaken
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: failed to check email uniqueness: %w", op, err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (s *UserStorage) DeleteUser(ctx context.Context, id string) error {
	const (
		op = "user.storage.DeleteUser"

		query = `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
		// TODO: add deleting session
	)

	_, err := s.Exec(ctx, query, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return c.ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to delete user: %w", op, err)
	}

	return nil
}