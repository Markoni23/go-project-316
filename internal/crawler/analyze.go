package crawler

import (
	"bytes"
	. "code/internal/crawler/models"
	"code/internal/crawler/parser"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type ParseTask struct {
	URL        string
	Depth      int
	ParentPage *Page
}

func Analyze(ctx context.Context, options Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	rep := Report{
		RootUrl:     options.Url,
		Depth:       options.Depth,
		GeneratedAt: time.Now(),
		Pages:       make([]Page, 0),
	}

	seen := NewSeenPages()
	tasks := make(chan ParseTask)

	var wg sync.WaitGroup
	var rootPage *Page

	for i := 0; i < options.WorkersCount; i++ {
		go func() {
			for task := range tasks {
				go func(task ParseTask) {
					defer wg.Done()
					if task.ParentPage != nil {
						fmt.Println("STARTED ", task.URL, "PARENT ", task.ParentPage.Url)
					} else {
						fmt.Println("STARTED ", task.URL, "PARENT root")
					}
					if task.Depth >= options.Depth {
						return
					}

					resp, err := options.Client.Get(task.URL)
					seen.Add(task.URL)
					if err != nil {
						fmt.Println(err)
						return
					}

					page := Page{
						Url:          task.URL,
						Depth:        options.Depth - task.Depth,
						HttpStatus:   resp.StatusCode,
						DiscoveredAt: time.Now(),
					}

					if resp.StatusCode >= http.StatusBadRequest {
						page.Error = &resp.Status
					}

					body, err := readBody(resp.Body)
					if err != nil {
						fmt.Println(err)
					}

					page.Status = resp.Status

					page.SEO = parser.ParseSEOFromBody(body)

					links := parser.ParseLinksFromBody(body, task.URL)

					brokenLinks := make([]BrokenLink, 0)

					for _, link := range checkLinks(links, options.Client, options.Delay) {
						if link.IsBroken {
							brokenLinks = append(brokenLinks, BrokenLink{Url: link.Url, Error: link.Error, StatusCode: link.StatusCode})
							continue
						}

						if seen.Seen(link.Url) {
							continue
						}

						wg.Add(1)
						fmt.Println("ADDED ", link.Url, "PARENT ", page.Url)
						tasks <- ParseTask{URL: link.Url, Depth: task.Depth + 1, ParentPage: &page}
					}

					if len(brokenLinks) > 0 {
						page.BrokenLinks = brokenLinks
					}

					fmt.Println(page.Url, page.Depth)
					if task.ParentPage != nil {
						fmt.Println("CHILD", (&page).Url, " PARENT ", task.ParentPage.Url)
						task.ParentPage.AddChildPage(&page)
					} else {
						rootPage = &page
					}
				}(task)
			}
		}()
	}

	wg.Add(1)
	tasks <- ParseTask{options.Url, 1, nil}

	wg.Wait()
	close(tasks)
	rep.Pages = append(rep.Pages, *rootPage)

	formatted, err := ReportFormat(rep)
	if err != nil {
		return "", err
	}

	return string(formatted), nil
}

type CheckLinkResult struct {
	IsBroken   bool
	Url        string
	StatusCode int
	Error      string
}

func checkLinks(
	links []string,
	client *http.Client,
	delay time.Duration,
) []CheckLinkResult {
	checkedLinks := make([]CheckLinkResult, len(links))
	var wg sync.WaitGroup
	wg.Add(len(links))
	for i, link := range links {
		go func(link string, index int) {
			defer wg.Done()
			resp, err := client.Head(link)
			if err != nil {
				checkedLinks[index] = CheckLinkResult{
					IsBroken: true,
					Url:      link,
					Error:    err.Error(),
				}
			} else if resp.StatusCode != http.StatusOK {
				checkedLinks[index] = CheckLinkResult{
					IsBroken:   true,
					Url:        link,
					StatusCode: resp.StatusCode,
				}
			} else {
				checkedLinks[index] = CheckLinkResult{
					IsBroken: false,
					Url:      link,
				}
			}
		}(link, i)
	}
	wg.Wait()
	return checkedLinks
}

func readBody(r io.ReadCloser) (data []byte, err error) {
	defer func() {
		if closeErr := r.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
