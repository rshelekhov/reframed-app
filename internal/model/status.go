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
		ID    int32  `db:"id"`
		Title string `db:"title"`
	}

	StatusResponseData struct {
		ID    int    `json:"status_id"`
		Title string `json:"title"`
	}
)
