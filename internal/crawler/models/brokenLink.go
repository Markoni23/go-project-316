package models

type BrokenLink struct {
	Url        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}
