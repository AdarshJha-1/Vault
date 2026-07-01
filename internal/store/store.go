package store

import (
	"sync"
)

type node struct {
	key   string
	value string

	prev *node
	next *node
}

func createNode(key, value string) *node {
	return &node{
		key:   key,
		value: value,
		prev:  nil,
		next:  nil,
	}
}

type Store interface {
	Set(key, value string)
	Get(key string) (string, bool)
	Delete(key string) bool
}

type store struct {
	mu       sync.RWMutex
	ky_val   map[string]*node
	maxLimit int

	head *node
	tail *node
}

func GetStore(maxLimit int) Store {

	head := createNode("", "")
	tail := createNode("", "")

	head.next = tail
	tail.prev = head
	return &store{
		mu:       sync.RWMutex{},
		ky_val:   map[string]*node{},
		maxLimit: maxLimit,
		head:     head,
		tail:     tail,
	}
}

func (s *store) moveToFront(n *node) {
	s.removeNode(n)
	s.addInList(n)
}

func (s *store) removeNode(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev

	n.next = nil
	n.prev = nil
}

func (s *store) deleteFromStore(n *node) {
	s.removeNode(n)
	delete(s.ky_val, n.key)
}

func (s *store) addInList(newNode *node) {
	newNode.next = s.head.next
	newNode.prev = s.head
	s.head.next.prev = newNode
	s.head.next = newNode
}

func (s *store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingNode, ok := s.ky_val[key]; ok {
		existingNode.value = value
		s.moveToFront(existingNode)
		return
	}

	newNode := createNode(key, value)
	s.ky_val[key] = newNode
	s.addInList(newNode)

	if len(s.ky_val) > s.maxLimit {
		nodeToDelete := s.tail.prev
		s.deleteFromStore(nodeToDelete)
	}
}

func (s *store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingNode, ok := s.ky_val[key]; ok {
		s.moveToFront(existingNode)
		return existingNode.value, true
	}
	return "", false
}

func (s *store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingNode, ok := s.ky_val[key]; ok {
		s.deleteFromStore(existingNode)
		return true
	}
	return false
}
