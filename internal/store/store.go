package store

import (
	"sync"
)

type Store interface {
	// Key/Val
	Set(key, value string)
	Get(key string) (string, bool)
}

type store struct {
	mu     sync.RWMutex
	ky_val map[string]string
	list   map[string][]string
}

func GetStore() Store {
	return &store{
		ky_val: make(map[string]string),
		list:   make(map[string][]string),
	}
}

func (s *store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ky_val[key] = value
}

func (s *store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.ky_val[key]
	return value, ok
}
