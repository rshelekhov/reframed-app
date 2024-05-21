package api_tests

import "github.com/brianvoe/gofakeit/v6"

const (
	host                  = "localhost:8082"
	passwordDefaultLength = 10
)

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, true, passwordDefaultLength)
}
