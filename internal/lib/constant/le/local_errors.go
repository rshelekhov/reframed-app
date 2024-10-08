package le

type LocalError string

func (l LocalError) Error() string {
	return string(l)
}

const (
	// ===========================================================================
	//   auth errors
	// ===========================================================================

	ErrAppIDDoesNotExists           LocalError = "app_id does not exists"
	ErrUserNotFound                 LocalError = "user not found"
	ErrUserUnauthenticated          LocalError = "user is not authenticated"
	ErrFailedToLoginUser            LocalError = "failed to login user"
	ErrFailedToCreateUser           LocalError = "failed to create user"
	ErrFailedToRequestResetPassword LocalError = "failed to request reset password"
	ErrFailedToGetTokenData         LocalError = "failed to get token data"
	ErrFailedToGetUserIDFromToken   LocalError = "failed to get user id from token"
	ErrFailedToGetRefreshToken      LocalError = "failed to get refresh token from context"
	ErrFailedToRefreshTokens        LocalError = "failed to refresh tokens"
	ErrFailedGoGetClaimsFromToken   LocalError = "failed to get claims from token"
	ErrFailedToLogout               LocalError = "failed to logout"

	ErrEmailVerificationTokenExpiredWithEmailResent LocalError = "verification token expired, a new email with a new token has been sent to the user"
	ErrEmailVerificationTokenNotFound               LocalError = "email verification token not found"
	ErrEmailVerificationTokenNotFoundInQuery        LocalError = "email verification token not found in query"
	ErrFailedToVerifyEmail                          LocalError = "failed to verify email"

	ErrResetPasswordTokenExpiredWithEmailResent LocalError = "reset password token expired, a new email with a new token has been sent to the user"
	ErrResetPasswordTokenNotFound               LocalError = "reset password token not found"
	ErrResetPasswordTokenNotFoundInQuery        LocalError = "reset password token not found in query"
	ErrFailedToChangePassword                   LocalError = "failed to change password"
	ErrUpdatedPasswordMustNotMatchTheCurrent    LocalError = "updated password must not match the current password"

	// ===========================================================================
	//   handler errors
	// ===========================================================================

	ErrEmptyRequestBody         LocalError = "request body is empty"
	ErrInvalidJSON              LocalError = "failed to decode request body"
	ErrEmptyData                LocalError = "data is empty"
	ErrInvalidData              LocalError = "invalid data"
	ErrFailedToGetData          LocalError = "failed to get data"
	ErrFailedToValidateData     LocalError = "failed to validate data"
	ErrFailedToParseQueryParams LocalError = "failed to parse query params"
	ErrInvalidCursor            LocalError = "invalid format for cursor, expected object id type string or YYYY-MM-DD"

	// ===========================================================================
	//   user errors
	// ===========================================================================

	ErrUserAlreadyExists  LocalError = "user with this email already exists"
	ErrEmailAlreadyTaken  LocalError = "this email already taken"
	ErrNoChangesDetected  LocalError = "no changes detected"
	ErrFailedToUpdateUser LocalError = "failed to update user"
	ErrFailedToDeleteUser LocalError = "failed to delete user"

	// ===========================================================================
	//   list errors
	// ===========================================================================

	ErrNoListsFound            LocalError = "no lists found"
	ErrListNotFound            LocalError = "list not found"
	ErrDefaultListNotFound     LocalError = "default list not found"
	ErrFailedToCreateList      LocalError = "failed to create list"
	ErrFailedToGetLists        LocalError = "failed to get lists"
	ErrFailedToUpdateList      LocalError = "failed to update list"
	ErrFailedToDeleteList      LocalError = "failed to delete list"
	ErrCannotDeleteDefaultList LocalError = "cannot delete default list"
	ErrEmptyQueryListID        LocalError = "list_id is empty in query"

	// ===========================================================================
	//   heading errors
	// ===========================================================================

	ErrNoHeadingsFound             LocalError = "no headings found"
	ErrHeadingNotFound             LocalError = "heading not found"
	ErrDefaultHeadingNotFound      LocalError = "default heading not found"
	ErrFailedToCreateHeading       LocalError = "failed to create heading"
	ErrFailedToGetHeadingsByListID LocalError = "failed to get headings by list ID"
	ErrFailedToUpdateHeading       LocalError = "failed to update heading"
	ErrFailedToMoveHeading         LocalError = "failed to move heading"
	ErrFailedToDeleteHeading       LocalError = "failed to delete heading"
	ErrEmptyQueryHeadingID         LocalError = "heading_id is empty in query"

	// ===========================================================================
	//   task errors
	// ===========================================================================

	ErrNoTasksFound         LocalError = "no tasks found"
	ErrTaskNotFound         LocalError = "task not found"
	ErrTaskStatusIDNotFound LocalError = "task status_id not found"
	ErrFailedToCreateTask   LocalError = "failed to create task"
	ErrFailedToUpdateTask   LocalError = "failed to update task"
	ErrFailedToCompleteTask LocalError = "failed to complete task"
	ErrFailedToMoveTask     LocalError = "failed to move task"
	ErrFailedToArchiveTask  LocalError = "failed to archive task"
	ErrInvalidTaskTimeRange LocalError = "invalid task time range"

	// ===========================================================================
	//   tag errors
	// ===========================================================================

	ErrTagNotFound LocalError = "tag not found"
	ErrNoTagsFound LocalError = "no tags found"

	// ===========================================================================
	//   status errors
	// ===========================================================================

	ErrNoStatusesFound              LocalError = "no statuses found"
	ErrStatusNotFound               LocalError = "task status not found"
	ErrFailedToConvertStatusIDtoInt LocalError = "failed to convert status_id to int"
	ErrInvalidStatusID              LocalError = "invalid status_id"

	// ===========================================================================
	//   other errors
	// ===========================================================================

	ErrFailedToWriteResponse LocalError = "failed to write response"
	ErrBadRequest            LocalError = "bad request"
)
