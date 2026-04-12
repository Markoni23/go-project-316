package parser

import (
	"bytes"
	"code/internal/crawler/models"
	"fmt"
	"html"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ParseLinksFromBody(body []byte, url string) []string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil
	}

	var links []string
	seen := map[string]bool{}
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		href = strings.TrimSpace(href)
		if href == "" {
			return
		}
		if seen[href] {
			return
		}

		if !strings.HasPrefix(href, "http") {
			return
		}

		sameDomain, err := isSameDomain(href, url)
		if err != nil {
			log.Println(err)
			return
		}
		if !sameDomain {
			return
		}

		seen[href] = true
		links = append(links, href)
	})

	return links
}

func ParseSEOFromBody(body []byte) models.SEO {
	seo := models.SEO{}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		fmt.Println(err)
		return seo
	}

	title := strings.TrimSpace(doc.Find("title").Text())

	if title != "" {
		seo.Title = html.UnescapeString(title)
		seo.HasTitle = true
	}

	doc.Find("meta[name]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		name, ok := s.Attr("name")
		if !ok {
			return true
		}
		if strings.EqualFold(strings.TrimSpace(name), "description") {
			content, ok := s.Attr("content")
			if ok {
				content = strings.TrimSpace(html.UnescapeString(content))
				if content != "" {
					seo.HasDescription = true
					seo.Description = content
				}
			}
			return false
		}
		return true
	})

	if doc.Find("h1").Length() > 0 {
		seo.HasH1 = true
	}

	return seo
}

func isSameDomain(first, second string) (bool, error) {
	f, err := normalizedHostname(first)
	if err != nil {
		return false, err
	}

	s, err := normalizedHostname(second)
	if err != nil {
		return false, err
	}

	return f == s, nil
}

func normalizedHostname(link string) (string, error) {
	f, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	normalized := func(host string) string {
		host = strings.ToLower(host)
		return strings.TrimPrefix(host, "www.")
	}
	return normalized(f.Hostname()), nil
}
