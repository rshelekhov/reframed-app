package postgres

import (
	"github.com/google/uuid"
	"time"
)

// User DB model
type User struct {
	ID        uuid.UUID
	Email     string
	Password  string
	RoleID    int
	FirstName string
	LastName  string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
