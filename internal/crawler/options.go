package crawler

import (
	"net/http"
	"time"
)

type Options struct {
	Client  *http.Client
	Retries int
	Depths  int
	Delay   time.Duration
	Url     string
}
