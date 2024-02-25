package mock

import (
	"context"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/stretchr/testify/mock"
)

type ListStorage struct {
	mock.Mock
}

func (l *ListStorage) CreateList(ctx context.Context, list model.List) error {
	args := l.Called(ctx, list)
	return args.Error(0)
}

func (l *ListStorage) GetListByID(ctx context.Context, id string) (model.List, error) {
	//TODO implement me
	panic("implement me")
}

func (l *ListStorage) GetLists(ctx context.Context, pgn model.Pagination) ([]model.List, error) {
	//TODO implement me
	panic("implement me")
}

func (l *ListStorage) UpdateList(ctx context.Context, list model.List) error {
	//TODO implement me
	panic("implement me")
}

func (l *ListStorage) DeleteList(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
