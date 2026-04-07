package crawler

import (
	"net/http"
	"time"
)

type Options struct {
	Client  *http.Client
	Retries int
	Depth   int
	Delay   time.Duration
	Url     string
}
