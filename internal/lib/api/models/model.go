package models

// Pagination represents pagination parameters
type Pagination struct {
	Limit  string `json:"limit"`
	Offset string `json:"offset"`
}
