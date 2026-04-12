package models

import (
	"net/http"
	"time"
)

type Options struct {
	Client       *http.Client
	Retries      int
	Depth        int
	Delay        time.Duration
	Url          string
	WorkersCount int
	Timeout      time.Duration
	UserAgent    string
}
