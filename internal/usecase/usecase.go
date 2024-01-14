package usecase

import "github.com/rshelekhov/reframed/internal/api/handler"

// UserUsecases defines the user use-cases
type UserUsecases interface {
	handler.UserCreater
	handler.UserIDGetter
	handler.UsersGetter
	handler.UserUpdater
	handler.UserDeleter
}
