package user

import (
	"github.com/google/uuid"
	"time"
)

// User DB model
type User struct {
	ID        uuid.UUID `db:"id"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	RoleID    int       `db:"role_id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Phone     string    `db:"phone"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}

// CreateUser uses in the request body and service layer
type CreateUser struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	RoleID    int    `json:"role_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

// UpdateUser uses in the request body and service layer
type UpdateUser struct {
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	RoleID    int       `json:"role_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Users used in the response body and service layer
type Users []*User
