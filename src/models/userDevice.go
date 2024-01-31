package models

import "time"

type UserDevice struct {
	ID            string     `db:"id" json:"id,omitempty"`
	UserID        string     `db:"user_id" json:"user_id,omitempty"`
	UserAgent     string     `db:"user_agent" json:"user_agent,omitempty"`
	IP            string     `db:"ip" json:"ip,omitempty"`
	Detached      bool       `db:"detached" json:"detached,omitempty"`
	LatestLoginAt *time.Time `db:"latest_login_at" json:"latest_login_at,omitempty"`
	DetachedAt    *time.Time `db:"detached_at" json:"detached_at,omitempty"`
}
