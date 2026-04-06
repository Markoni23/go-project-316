package crawler

import (
	"context"
	"net/http"
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

	rep, err := CreateReport(ctx, options.Url, options.Depths, resp)
	if err != nil {
		return "", err
	}

	formatted, err := ReportFormat(rep)
	if err != nil {
		return "", err
	}

	return string(formatted), nil
}
