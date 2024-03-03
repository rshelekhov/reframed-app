package mock

import (
	"context"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/stretchr/testify/mock"
)

type UserStorage struct {
	mock.Mock
}

func (u *UserStorage) CreateUser(ctx context.Context, user model.User) error {
	args := u.Called(ctx, user)
	return args.Error(0)
}

func (u *UserStorage) GetUserByID(ctx context.Context, id string) (model.User, error) {
	args := u.Called(ctx, id)
	result := args.Get(0)
	return result.(model.User), args.Error(1)
}

func (u *UserStorage) GetUsers(ctx context.Context, pgn model.Pagination) ([]model.User, error) {
	args := u.Called(ctx, pgn)
	return args.Get(0).([]model.User), args.Error(1)
}

func (u *UserStorage) UpdateUser(ctx context.Context, user model.User) error {
	args := u.Called(ctx, user)
	return args.Error(0)
}

func (u *UserStorage) DeleteUser(ctx context.Context, id string) error {
	args := u.Called(ctx, id)
	return args.Error(0)
}
