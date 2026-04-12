package models

import "sync"

type PagesMap struct {
	pages map[string]*Page
	mu    sync.Mutex
}

func (pm *PagesMap) Add(page *Page) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.pages[page.Url] = page
}

func (pm *PagesMap) Get(url string) (*Page, bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if p, ok := pm.pages[url]; ok {
		return p, true
	}
	return nil, false
}

func (pm *PagesMap) GetURLS() []string {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	urls := make([]string, 0, len(pm.pages))
	for url, _ := range pm.pages {
		urls = append(urls, url)
	}
	return urls
}

func NewPagesMap() *PagesMap {
	return &PagesMap{
		pages: make(map[string]*Page),
	}
}
