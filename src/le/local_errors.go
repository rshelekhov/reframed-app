package le

type LocalError string

func (l LocalError) Error() string {
	return string(l)
}

const (
	//===========================================================================
	//  auth errors
	//===========================================================================

	ErrUserNotFound            LocalError = "user not found"
	ErrInvalidCredentials      LocalError = "invalid credentials"
	ErrUserDeviceNotFound      LocalError = "user device not found"
	ErrSessionNotFound         LocalError = "session not found"
	ErrFailedToRegisterDevice  LocalError = "failed to register device"
	ErrFailedToCheckDevice     LocalError = "failed to check device"
	ErrFailedToCreateSession   LocalError = "failed to create session"
	ErrFailedToGetAccessToken  LocalError = "failed to get token from context"
	ErrFailedToGetRefreshToken LocalError = "failed to get refresh token from context"
	ErrRefreshTokenExpired     LocalError = "refresh token expired"
	ErrFailedToCreateUser      LocalError = "failed to create user"

	//===========================================================================
	//  handlers errors
	//===========================================================================

	ErrEmptyRequestBody     LocalError = "request body is empty"
	ErrInvalidJSON          LocalError = "failed to decode request body"
	ErrEmptyData            LocalError = "data is empty"
	ErrInvalidData          LocalError = "invalid data"
	ErrFailedToGetData      LocalError = "failed to get data"
	ErrFailedToValidateData LocalError = "failed to validate data"

	//===========================================================================
	//  user errors
	//===========================================================================

	ErrNoUsersFound              LocalError = "no users found"
	ErrUserAlreadyExists         LocalError = "user with this email already exists"
	ErrEmailAlreadyTaken         LocalError = "this email already taken"
	ErrNoChangesDetected         LocalError = "no changes detected"
	ErrNoPasswordChangesDetected LocalError = "no password changes detected"
	ErrFailedToUpdateUser        LocalError = "failed to update user"

	//===========================================================================
	//  list errors
	//===========================================================================

	ErrNoListsFound       LocalError = "no lists found"
	ErrFailedToCreateList LocalError = "failed to create list"
	ErrFailedToGetLists   LocalError = "failed to get lists"

	//===========================================================================
	//  other errors
	//===========================================================================

	ErrFailedToParsePagination LocalError = "failed to parse limit and offset"
)
