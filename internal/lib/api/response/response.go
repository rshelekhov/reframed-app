package response

const (
	StatusOK    = "OK"
	StatusError = "ERROR"
)

type Response struct {
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	Success string `json:"success,omitempty"`
	ID      string `json:"id,omitempty"`
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func Success(msg string, id string) Response {
	return Response{
		Status:  StatusOK,
		Success: msg,
		ID:      id,
	}
}

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}
