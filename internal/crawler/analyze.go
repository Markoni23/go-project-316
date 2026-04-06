package crawler

import (
	"context"
	"fmt"
	"net/http"
)

func Analyze(ctx context.Context, options Options) error {
	req, err := http.NewRequestWithContext(ctx, "GET", options.Url, nil)
	if err != nil {
		return err
	}

	resp, err := options.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rep, err := CreateReport(ctx, options.Url, options.Depths, resp)
	if err != nil {
		return err
	}

	formatted, err := ReportFormat(rep)
	if err != nil {
		return err
	}

	fmt.Println(string(formatted))

	return nil
}
