package handlers

import "errors"

var (
	ErrEmptyID = errors.New("id is empty")

	ErrEmptyRequestBody = errors.New("request body is empty")
	ErrInvalidJSON      = errors.New("failed to decode request body")

	ErrEmptyData            = errors.New("data is empty")
	ErrInvalidData          = errors.New("invalid data")
	ErrFailedToValidateData = errors.New("failed to validate data")

	ErrFailedToGetData         = errors.New("failed to get data")
	ErrFailedToParsePagination = errors.New("failed to parse limit and offset")

	ErrFailedToCreateToken = errors.New("failed to create token")
	ErrInvalidCredentials  = errors.New("invalid credentials")
)
