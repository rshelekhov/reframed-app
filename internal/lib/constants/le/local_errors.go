package le

type LocalError string

func (l LocalError) Error() string {
	return string(l)
}

// TODO: add unwrapping

const (
	// ===========================================================================
	//   auth errors
	// ===========================================================================

	ErrAppIDDoesNotExists         LocalError = "app_id does not exists"
	ErrUserNotFound               LocalError = "user not found"
	ErrUserUnauthenticated        LocalError = "user is not authenticated"
	ErrFailedToLoginUser          LocalError = "failed to login user"
	ErrFailedToCreateUser         LocalError = "failed to create user"
	ErrFailedToGetTokenData       LocalError = "failed to get token data"
	ErrFailedToGetUserIDFromToken LocalError = "failed to get user id from token"
	ErrFailedToGetRefreshToken    LocalError = "failed to get refresh token from context"
	ErrFailedGoGetClaimsFromToken LocalError = "failed to get claims from token"
	ErrFailedToLogout             LocalError = "failed to logout"

	// ===========================================================================
	//   controller errors
	// ===========================================================================

	ErrEmptyRequestBody         LocalError = "request body is empty"
	ErrInvalidJSON              LocalError = "failed to decode request body"
	ErrEmptyData                LocalError = "data is empty"
	ErrInvalidData              LocalError = "invalid data"
	ErrFailedToGetData          LocalError = "failed to get data"
	ErrFailedToValidateData     LocalError = "failed to validate data"
	ErrFailedToParseQueryParams LocalError = "failed to parse query params"

	// ===========================================================================
	//   user errors
	// ===========================================================================

	ErrUserAlreadyExists         LocalError = "user with this email already exists"
	ErrEmailAlreadyTaken         LocalError = "this email already taken"
	ErrNoChangesDetected         LocalError = "no changes detected"
	ErrNoPasswordChangesDetected LocalError = "no password changes detected"
	ErrFailedToUpdateUser        LocalError = "failed to update user"
	ErrFailedToDeleteUser        LocalError = "failed to delete user"

	// ===========================================================================
	//   list errors
	// ===========================================================================

	ErrNoListsFound             LocalError = "no lists found"
	ErrListNotFound             LocalError = "list not found"
	ErrFailedToCreateList       LocalError = "failed to create list"
	ErrFailedToGetLists         LocalError = "failed to get lists"
	ErrFailedToGetDefaultListID LocalError = "failed to get default list ID"
	ErrFailedToUpdateList       LocalError = "failed to update list"
	ErrFailedToDeleteList       LocalError = "failed to delete list"
	ErrEmptyQueryListID         LocalError = "list ID is empty in query"

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
	ErrEmptyQueryHeadingID         LocalError = "heading ID is empty in query"

	// ===========================================================================
	//   task errors
	// ===========================================================================

	ErrNoTasksFound         LocalError = "no tasks found"
	ErrTaskNotFound         LocalError = "task not found"
	ErrTaskStatusNotFound   LocalError = "task status not found"
	ErrFailedToCreateTask   LocalError = "failed to create task"
	ErrFailedToUpdateTask   LocalError = "failed to update task"
	ErrFailedToCompleteTask LocalError = "failed to complete task"
	ErrFailedToMoveTask     LocalError = "failed to move task"
	ErrFailedToDeleteTask   LocalError = "failed to delete task"
	ErrEmptyQueryTaskID     LocalError = "task ID is empty in query"
	ErrInvalidTaskTimeRange LocalError = "invalid task time range"

	// ===========================================================================
	//   tag errors
	// ===========================================================================

	ErrTagNotFound            LocalError = "tag not found"
	ErrNoTagsFound            LocalError = "no tags found"
	ErrFailedToCreateTag      LocalError = "failed to create tag"
	ErrFailedToDeleteTag      LocalError = "failed to delete tag"
	ErrFailedToLinkTagsToTask LocalError = "failed to link tags to task"

	// ===========================================================================
	//   other errors
	// ===========================================================================

	ErrFailedToWriteResponse LocalError = "failed to write response"
)
