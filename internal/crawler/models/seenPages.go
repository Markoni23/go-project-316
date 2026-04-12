package models

import "sync"

type SeenPages struct {
	seenPages map[string]bool
	mu        sync.Mutex
}

func (s *SeenPages) Add(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seenPages[url] = true
}

func (s *SeenPages) Seen(url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seenPages[url]
}

func NewSeenPages() *SeenPages {
	return &SeenPages{
		seenPages: make(map[string]bool),
	}
}
