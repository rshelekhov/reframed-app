package models

type PasswordHashBcrypt struct {
	Cost int    `json:"cost,omitempty" mapstructure:"PASSWORD_HASH_BCRYPT_COST"`
	Salt string `json:"salt,omitempty" mapstructure:"PASSWORD_HASH_BCRYPT_SALT"`
}
