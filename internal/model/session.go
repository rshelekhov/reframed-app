package model

import (
	"time"
)

type Session struct {
	UserID       string    `db:"user_id"`
	DeviceID     string    `db:"device_id"`
	RefreshToken string    `db:"refresh_token"`
	LastVisitAt  time.Time `db:"last_visit_at"`
	ExpiresAt    time.Time `db:"expires_at"`
}

func (s Session) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}
