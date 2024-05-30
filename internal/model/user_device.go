package model

type UserDeviceRequestData struct {
	UserAgent string `json:"user_agent"`
	IP        string `json:"ip"`
}
