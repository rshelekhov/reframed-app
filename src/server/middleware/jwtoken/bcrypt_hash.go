package jwtoken

import (
	"github.com/rshelekhov/reframed/src/models"
	"golang.org/x/crypto/bcrypt"
)

func PasswordHash(password string, params models.PasswordHashBcrypt, salt []byte) (string, error) {
	cost := params.Cost
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}

	hash, err := passwordHashBcrypt(password, cost, salt)
	if err != nil {
		return "", err
	}

	return hash, nil
}
