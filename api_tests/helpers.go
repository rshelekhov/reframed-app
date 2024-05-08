package api_tests

import "github.com/brianvoe/gofakeit/v6"

const (
	host              = "localhost:8082"
	appID             = int32(1)
	emptyAppID        = int32(0)
	invalidAppID      = int32(-1)
	passDefaultLength = 10
)

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, true, passDefaultLength)
}
