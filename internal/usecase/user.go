package usecase

import (
	"context"

	"github.com/rshelekhov/reframed/internal/port"
)

type UserUsecase struct {
	storage port.UserStorage
}

func NewUserUsecase(storage port.UserStorage) *UserUsecase {
	return &UserUsecase{storage: storage}
}

func (u *UserUsecase) DeleteUserRelatedData(ctx context.Context, userID string) error {
	if err := u.storage.DeleteUserData(ctx, userID); err != nil {
		return err
	}

	return nil
}
