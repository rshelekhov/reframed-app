package models

import "time"

type Session struct {
	UserID       string     `db:"user_id" json:"user_id,omitempty"`
	DeviceID     string     `db:"device_id" json:"device_id,omitempty"`
	RefreshToken string     `db:"refresh_token" json:"refresh_token"`
	LastVisitAt  *time.Time `db:"last_visit_at" json:"last_visit_at,omitempty"`
	ExpiresAt    *time.Time `db:"expires_at" json:"expires_at"`
}
