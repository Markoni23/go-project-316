package crawler

import (
	"context"
	"net/http"
	"time"
)

func Analyze(ctx context.Context, options Options) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", options.Url, nil)
	if err != nil {
		return "", err
	}

	resp, err := options.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	rep := Report{
		RootUrl:     options.Url,
		Depth:       options.Depth,
		GeneratedAt: time.Now(),
		Pages:       make([]Page, 0),
	}

	page, err := CreatePage(ctx, options)

	if err != nil {
		return "", err
	}

	rep.Pages = append(rep.Pages, page)

	formatted, err := ReportFormat(rep)
	if err != nil {
		return "", err
	}

	return string(formatted), nil
}
