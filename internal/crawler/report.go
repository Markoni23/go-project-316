package crawler

import (
	"context"
	"net/http"
	"time"
)

type Report struct {
	RootUrl     string    `json:"root_url"`
	Depth       int       `json:"depth"`
	GeneratedAt time.Time `json:"generated_at"`
	Pages       []Page    `json:"pages"`
}

func CreateReport(ctx context.Context, root string, depth int, response *http.Response) (Report, error) {
	rep := Report{
		RootUrl:     root,
		Depth:       depth,
		GeneratedAt: time.Now(),
	}

	page := Page{
		Url:        root,
		Depth:      depth - 1,
		HttpStatus: response.StatusCode,
		Status:     response.Status,
		Error:      "",
	}

	rep.Pages = append(rep.Pages, page)

	return rep, nil
}
