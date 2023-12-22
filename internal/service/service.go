package service

import "github.com/rshelekhov/remedi/internal/storage/postgres"

type Service struct {
	storage postgres.Storage
}

func NewService(s postgres.Storage) Service {
	return Service{
		storage: s,
	}
}
