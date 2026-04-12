package models

import "time"

type Page struct {
	Url          string       `json:"url"`
	Depth        int          `json:"depth"`
	HttpStatus   int          `json:"http_status"`
	Status       string       `json:"status"`
	Error        *string      `json:"error,omitempty"`
	SEO          SEO          `json:"seo,omitempty"`
	Pages        []Page       `json:"pages"`
	BrokenLinks  []BrokenLink `json:"broken_links,omitempty"`
	DiscoveredAt time.Time    `json:"discovered_at"`
}

func (p *Page) AddChildPage(child *Page) {
	p.Pages = append(p.Pages, *child)
}
