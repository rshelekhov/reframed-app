package model

type StatusName string

const (
	StatusNotStarted StatusName = "Not started"
	StatusPlanned    StatusName = "Planned"
	StatusCompleted  StatusName = "Completed"
	StatusArchived   StatusName = "Archived"
)

func (s StatusName) String() string {
	return string(s)
}

type (
	Status struct {
		ID    string `db:"id"`
		Title string `db:"title"`
	}

	StatusRequestData struct {
		ID    string `json:"status_id"`
		Title string `json:"title"`
	}

	StatusResponseData struct {
		ID    string `json:"status_id"`
		Title string `json:"title"`
	}
)
