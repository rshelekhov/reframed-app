package model

// Pagination represents pagination parameters
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
