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

	ErrFailedToRegisterDevice = errors.New("failed to register device")
	ErrFailedToCheckDevice    = errors.New("failed to check device")
	ErrDeviceNotFound         = errors.New("device not found")

	ErrFailedToCreateSession   = errors.New("failed to create session")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrFailedToGetRefreshToken = errors.New("failed to get refresh token from context")
)
