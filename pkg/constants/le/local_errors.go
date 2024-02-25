package le

type LocalError string

func (l LocalError) Error() string {
	return string(l)
}

const (
	//===========================================================================
	//  auth errors
	//===========================================================================

	ErrUserNotFound               LocalError = "user not found"
	ErrInvalidCredentials         LocalError = "invalid credentials"
	ErrPasswordsDoNotMatch        LocalError = "passwords do not match"
	ErrUserHasNoPassword          LocalError = "user has no password"
	ErrUserDeviceNotFound         LocalError = "user device not found"
	ErrSessionNotFound            LocalError = "session not found"
	ErrFailedToLogin              LocalError = "failed to login"
	ErrFailedToCreateUser         LocalError = "failed to create user"
	ErrFailedToRegisterDevice     LocalError = "failed to register device"
	ErrFailedToCheckDevice        LocalError = "failed to check device"
	ErrFailedToDeleteRefreshToken LocalError = "failed to delete refresh token"
	ErrFailedToCreateSession      LocalError = "failed to create session"
	ErrFailedToRemoveSession      LocalError = "failed to remove session"
	ErrFailedToGetUserIDFromToken LocalError = "failed to get user id from token"
	ErrFailedToGetAccessToken     LocalError = "failed to get token from context"
	ErrFailedToGetRefreshToken    LocalError = "failed to get refresh token from context"
	ErrFailedToRefreshTokens      LocalError = "failed to refresh tokens"
	ErrFailedToLogout             LocalError = "failed to logout"
	ErrSessionExpired             LocalError = "session expired"

	//===========================================================================
	//  controller errors
	//===========================================================================

	ErrEmptyRequestBody         LocalError = "request body is empty"
	ErrInvalidJSON              LocalError = "failed to decode request body"
	ErrEmptyData                LocalError = "data is empty"
	ErrInvalidData              LocalError = "invalid data"
	ErrFailedToGetData          LocalError = "failed to get data"
	ErrFailedToValidateData     LocalError = "failed to validate data"
	ErrFailedToParseQueryParams LocalError = "failed to parse query params"

	//===========================================================================
	//  user errors
	//===========================================================================

	ErrNoUsersFound              LocalError = "no users found"
	ErrFailedToGetUsers          LocalError = "failed to get users"
	ErrUserAlreadyExists         LocalError = "user with this email already exists"
	ErrEmailAlreadyTaken         LocalError = "this email already taken"
	ErrNoChangesDetected         LocalError = "no changes detected"
	ErrNoPasswordChangesDetected LocalError = "no password changes detected"
	ErrFailedToUpdateUser        LocalError = "failed to update user"
	ErrFailedToDeleteUser        LocalError = "failed to delete user"

	//===========================================================================
	//  list errors
	//===========================================================================

	ErrNoListsFound       LocalError = "no lists found"
	ErrListNotFound       LocalError = "list not found"
	ErrFailedToCreateList LocalError = "failed to create list"
	ErrFailedToGetLists   LocalError = "failed to get lists"
	ErrFailedToUpdateList LocalError = "failed to update list"
	ErrFailedToDeleteList LocalError = "failed to delete list"
	ErrEmptyQueryListID   LocalError = "list ID is empty in query"

	//===========================================================================
	//  heading errors
	//===========================================================================

	ErrNoHeadingsFound             LocalError = "no headings found"
	ErrHeadingNotFound             LocalError = "heading not found"
	ErrFailedToCreateHeading       LocalError = "failed to create heading"
	ErrFailedToGetHeadingsByListID LocalError = "failed to get headings by list ID"
	ErrFailedToUpdateHeading       LocalError = "failed to update heading"
	ErrFailedToMoveHeading         LocalError = "failed to move heading"
	ErrFailedToDeleteHeading       LocalError = "failed to delete heading"
	ErrEmptyQueryHeadingID         LocalError = "heading ID is empty in query"

	//===========================================================================
	//  task errors
	//===========================================================================

	ErrNoTasksFound         LocalError = "no tasks found"
	ErrTaskNotFound         LocalError = "task not found"
	ErrFailedToCreateTask   LocalError = "failed to create task"
	ErrFailedToUpdateTask   LocalError = "failed to update task"
	ErrFailedToCompleteTask LocalError = "failed to complete task"
	ErrFailedToMoveTask     LocalError = "failed to move task"
	ErrFailedToDeleteTask   LocalError = "failed to delete task"
	ErrEmptyQueryTaskID     LocalError = "task ID is empty in query"
	ErrInvalidTaskTimeRange LocalError = "invalid task time range"

	//===========================================================================
	//  tag errors
	//===========================================================================

	ErrTagNotFound            LocalError = "tag not found"
	ErrNoTagsFound            LocalError = "no tags found"
	ErrFailedToCreateTag      LocalError = "failed to create tag"
	ErrFailedToDeleteTag      LocalError = "failed to delete tag"
	ErrFailedToLinkTagsToTask LocalError = "failed to link tags to task"

	//===========================================================================
	//  other errors
	//===========================================================================

	ErrFailedToWriteResponse LocalError = "failed to write response"
)
