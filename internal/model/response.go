package model

type Response struct {
	Code        int    `json:"code"`
	StatusText  string `json:"status_text"`
	Description string `json:"description"`
	Data        any    `json:"data"`
}
