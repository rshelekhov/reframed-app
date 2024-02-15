package mocks

import (
	"context"
	"github.com/rshelekhov/reframed/internal/domain"
	"github.com/stretchr/testify/mock"
)

type UserStorage struct {
	mock.Mock
}

func (u *UserStorage) CreateUser(ctx context.Context, user domain.User) error {
	args := u.Called(ctx, user)
	return args.Error(0)
}

func (u *UserStorage) GetUserByID(ctx context.Context, id string) (domain.User, error) {
	args := u.Called(ctx, id)
	result := args.Get(0)
	return result.(domain.User), args.Error(1)
}

func (u *UserStorage) GetUsers(ctx context.Context, pgn domain.Pagination) ([]domain.User, error) {
	args := u.Called(ctx, pgn)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (u *UserStorage) UpdateUser(ctx context.Context, user domain.User) error {
	args := u.Called(ctx, user)
	return args.Error(0)
}

func (u *UserStorage) DeleteUser(ctx context.Context, id string) error {
	args := u.Called(ctx, id)
	return args.Error(0)
}
