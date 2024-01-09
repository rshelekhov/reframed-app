package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rshelekhov/reframed/internal/entity"
	"github.com/rshelekhov/reframed/pkg/storage"
)

type UserStorage struct {
	*pgxpool.Pool
}

func NewUserStorage(pg *pgxpool.Pool) *UserStorage {
	return &UserStorage{pg}
}

// CreateUser creates a new user
func (s *UserStorage) CreateUser(ctx context.Context, user entity.User) error {

	const (
		op = "user.storage.CreateUser"

		querySelectRoleID = `SELECT id FROM roles WHERE id = $1`

		queryCheckUserExists = `SELECT CASE
										WHEN EXISTS(
											SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL
										) THEN 'active' 
										WHEN EXISTS(
											SELECT 1 FROM users WHERE email = $1 and deleted_at IS NOT NULL
										) THEN 'soft_deleted'
										ELSE 'not_found' END AS status`

		queryReplaceSoftDeletedUser = `WITH update_deleted AS (
												UPDATE users SET deleted_at = NULL WHERE email = $1 RETURNING *
											)
											INSERT INTO users
												(id, email, password, role_id, first_name, last_name, phone, updated_at)
												VALUES ($2, $3, $4, $5, $6, $7, $8, $9)`

		queryInsertUser = `INSERT INTO users
    						(id, email, password, role_id, first_name, last_name, phone, updated_at)
							VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	)

	// Begin transaction
	tx, err := s.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			err = fmt.Errorf("%s: failed to rollback transaction: %w", op, rollbackErr)
		}
	}()

	// Check if role exists
	var roleID int

	err = tx.QueryRow(ctx, querySelectRoleID, user.RoleID).Scan(&roleID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: role not found: %w", op, storage.ErrRoleNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: failed to check if role exists: %w", op, err)
	}

	var status string
	err = tx.QueryRow(ctx, queryCheckUserExists, user.Email).Scan(&status)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			return fmt.Errorf("%s: failed to rollback transaction: %w", op, errRollback)
		}
		return fmt.Errorf("%s: failed to check if user exists: %w", op, err)
	}

	if status == "soft_deleted" {
		_, err = tx.Exec(
			ctx,
			queryReplaceSoftDeletedUser,
			user.Email,
			user.ID,
			user.Password,
			roleID,
			user.FirstName,
			user.LastName,
			user.Phone,
			user.UpdatedAt)
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				return fmt.Errorf("%s: failed to rollback transaction: %w", op, errRollback)
			}
			return fmt.Errorf("%s: failed to replace soft deleted user: %w", op, err)
		}
	} else if status == "not_found" {
		_, err = tx.Exec(
			ctx,
			queryInsertUser,
			user.ID,
			user.Email,
			user.Password,
			roleID,
			user.FirstName,
			user.LastName,
			user.Phone,
			user.UpdatedAt,
		)
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				return fmt.Errorf("%s: failed to rollback transaction: %w", op, errRollback)
			}
			return fmt.Errorf("%s: failed to insert new user: %w", op, err)
		}
	} else {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			return fmt.Errorf("%s: failed to rollback transaction: %w", op, errRollback)
		}
		return fmt.Errorf("%s: user with this email already exists %w", op, storage.ErrUserAlreadyExists)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return nil
}

// GetUser returns a user by ID
func (s *UserStorage) GetUser(ctx context.Context, id string) (entity.GetUser, error) {
	const op = "user.storage.GetUser"

	var user entity.GetUser
	query := `SELECT id, email, role_id, first_name, last_name, phone, updated_at
							FROM users WHERE id = $1 AND deleted_at IS NULL`

	err := s.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.RoleID,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
	}
	if err != nil {
		return user, fmt.Errorf("%s: failed to get user: %w", op, err)
	}

	return user, nil
}

// GetUsers returns a list of users
func (s *UserStorage) GetUsers(ctx context.Context, pgn entity.Pagination) ([]*entity.GetUser, error) {
	const op = "user.storage.GetUsers"

	query := `SELECT id, email, role_id, first_name, last_name, phone, updated_at
							FROM users WHERE deleted_at IS NULL ORDER BY id DESC LIMIT $1 OFFSET $2`

	rows, err := s.Query(ctx, query, pgn.Limit, pgn.Offset)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	defer rows.Close()

	var users []*entity.GetUser
	users, err = pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[entity.GetUser])
	if err != nil {
		return nil, fmt.Errorf("%s: failed to collect rows: %w", op, err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("%s: no users found: %w", op, storage.ErrNoUsersFound)
	}

	return users, nil
}

// UpdateUser updates a user by ID
func (s *UserStorage) UpdateUser(ctx context.Context, user entity.User) error {
	const (
		op = "user.storage.UpdateUser"

		queryCheckUserExists = `SELECT CASE
										WHEN EXISTS(
											SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL
										) THEN 'active' 
										WHEN EXISTS(
											SELECT 1 FROM users WHERE email = $1 and deleted_at IS NOT NULL
										) THEN 'soft_deleted'
										ELSE 'not_found' END AS status`

		queryReplaceSoftDeletedUser = `WITH update_deleted AS (
												UPDATE users SET deleted_at = NULL WHERE email = $1 RETURNING *
											)
											INSERT INTO users
												(id, email, password, role_id, first_name, last_name, phone, updated_at)
												VALUES ($2, $3, $4, $5, $6, $7, $8, $9)`

		queryUpdateUser = `UPDATE users
								SET email = $1,
									password = $2,
									first_name = $3,
									last_name = $4,
									phone = $5,
									updated_at = $6
								WHERE id = $7 AND deleted_at IS NULL`
	)

	// TODO check if there are no changes

	// Begin transaction
	tx, err := s.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			err = fmt.Errorf("%s: failed to rollback transaction: %w", op, rollbackErr)
		}
	}()

	// Check if the updating email already exists
	var status string
	err = tx.QueryRow(ctx, queryCheckUserExists, user.Email).Scan(&status)
	if err != nil {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			return fmt.Errorf("%s: failed to rollback transaction: %w", op, errRollback)
		}
		return fmt.Errorf("%s: failed to check if user exists: %w", op, err)
	}

	if status == "soft_deleted" {
		_, err = tx.Exec(ctx,
			queryReplaceSoftDeletedUser,
			user.Email,
			user.ID,
			user.Email,
			user.Password,
			user.RoleID,
			user.FirstName,
			user.LastName,
			user.Phone,
			user.UpdatedAt,
		)
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				return fmt.Errorf("%s: failed to rollback transaction: %w", op, errRollback)
			}

			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == storage.UniqueConstraintViolation {
				return fmt.Errorf("%s: email already exists: %w", op, storage.ErrUserAlreadyExists)
			}
			return fmt.Errorf("%s: failed to replace soft deleted user: %w", op, err)
		}
	} else if status == "not_found" {
		result, err := tx.Exec(ctx,
			queryUpdateUser,
			user.Email,
			user.Password,
			user.FirstName,
			user.LastName,
			user.Phone,
			user.UpdatedAt,
			user.ID,
		)
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				return fmt.Errorf("%s: failed to rollback transaction: %w", op, errRollback)
			}
			return fmt.Errorf("%s: failed to update user: %w", op, err)
		}

		rowsAffected := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("%s: user with ID %s not found: %w", op, user.ID, storage.ErrUserNotFound)
		}

	} else {
		errRollback := tx.Rollback(ctx)
		if errRollback != nil {
			return fmt.Errorf("%s: failed to rollback transaction: %w", op, errRollback)
		}
		return fmt.Errorf("%s: user with this email already exists %w", op, storage.ErrUserAlreadyExists)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (s *UserStorage) DeleteUser(ctx context.Context, id string) error {
	const op = "user.storage.DeleteUser"

	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := s.Exec(ctx, query, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: user with this id not found %w", op, storage.ErrUserNotFound)
	}
	if err != nil {
		return fmt.Errorf("%s: failed to delete user: %w", op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s: user with ID %s not found: %w", op, id, storage.ErrUserNotFound)
	}

	return nil
}

// GetUserRoles returns a list of roles
func (s *UserStorage) GetUserRoles(ctx context.Context) ([]*entity.GetRole, error) {
	const op = "user.storage.GetUserRoles"

	query := `SELECT id, title FROM roles`

	rows, err := s.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute query: %w", op, err)
	}
	defer rows.Close()

	var roles []*entity.GetRole

	roles, err = pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[entity.GetRole])
	if err != nil {
		return nil, fmt.Errorf("%s: failed to collect rows: %w", op, err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("%s: no roles found: %w", op, storage.ErrNoRolesFound)
	}

	return roles, nil
}
