package mocks

import (
	"context"
	"github.com/rshelekhov/reframed/internal/models"
	"github.com/stretchr/testify/mock"
)

type ListStorage struct {
	mock.Mock
}

func (l *ListStorage) CreateList(ctx context.Context, list models.List) error {
	args := l.Called(ctx, list)
	return args.Error(0)
}

func (l *ListStorage) GetListByID(ctx context.Context, id string) (models.List, error) {
	//TODO implement me
	panic("implement me")
}

func (l *ListStorage) GetLists(ctx context.Context, pgn models.Pagination) ([]models.List, error) {
	//TODO implement me
	panic("implement me")
}

func (l *ListStorage) UpdateList(ctx context.Context, list models.List) error {
	//TODO implement me
	panic("implement me")
}

func (l *ListStorage) DeleteList(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
