package crawler

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Report struct {
	RootUrl     string    `json:"root_url"`
	Depth       int       `json:"depth"`
	GeneratedAt time.Time `json:"generated_at"`
	Pages       []Page    `json:"pages"`
}

type Page struct {
	Url          string       `json:"url"`
	Depth        int          `json:"depth"`
	HttpStatus   int          `json:"http_status"`
	Status       string       `json:"status"`
	Error        *string      `json:"error,omitempty"`
	BrokenLinks  []BrokenLink `json:"broken_links,omitempty"`
	DiscoveredAt time.Time    `json:"discovered_at"`
}

type BrokenLink struct {
	Url        string `json:"url"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
}

func CreatePage(ctx context.Context, options Options) (Page, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", options.Url, nil)
	if err != nil {
		return Page{}, err
	}

	resp, err := options.Client.Do(req)
	if err != nil {
		return Page{}, err
	}

	defer resp.Body.Close()
	page := Page{
		Url:          options.Url,
		Depth:        options.Depth - 1,
		HttpStatus:   resp.StatusCode,
		DiscoveredAt: time.Time{},
	}

	if resp.StatusCode >= http.StatusBadRequest {
		page.Error = &resp.Status
		return page, nil
	}

	page.Status = resp.Status

	links := parseLinksFromBody(resp.Body)

	brokenLinks := make([]BrokenLink, 0)

	for link := range checkBrokenLinks(links, options.Client, options.Delay) {
		brokenLinks = append(brokenLinks, link)
	}

	if len(brokenLinks) > 0 {
		page.BrokenLinks = brokenLinks
	}

	return page, nil
}

func parseLinksFromBody(body io.Reader) chan string {
	z := html.NewTokenizer(body)
	links := make(chan string)
	go func() {
		defer close(links)
		for {
			tt := z.Next()
			switch tt {
			case html.ErrorToken:
				return
			case html.StartTagToken:
				token := z.Token()
				if token.Data != "a" {
					continue
				}
				url, ok := getHref(token)
				if !ok {
					continue
				}
				if strings.HasPrefix(url, "http") {
					links <- url
				}
			}
		}
	}()
	return links
}

func checkBrokenLinks(links <-chan string, client *http.Client, delay time.Duration) chan BrokenLink {
	brokenLinks := make(chan BrokenLink)
	go func() {
		defer close(brokenLinks)
		for link := range links {
			resp, err := client.Head(link)
			if err != nil {
				brokenLinks <- BrokenLink{
					Url:   link,
					Error: err.Error(),
				}
			} else if resp.StatusCode != http.StatusOK {
				brokenLinks <- BrokenLink{
					Url:        link,
					StatusCode: resp.StatusCode,
				}
			}
		}
	}()
	return brokenLinks
}

func getHref(t html.Token) (href string, ok bool) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	return href, ok
}
