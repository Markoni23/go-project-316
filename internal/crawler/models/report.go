package models

import "time"

type Report struct {
	RootUrl     string    `json:"root_url"`
	Depth       int       `json:"depth"`
	GeneratedAt time.Time `json:"generated_at"`
	Pages       []Page    `json:"pages"`
}
