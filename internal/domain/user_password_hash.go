package domain

type PasswordHashBcrypt struct {
	Cost int    `json:"cost,omitempty"`
	Salt string `json:"salt,omitempty"`
}
