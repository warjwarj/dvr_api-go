package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// intStack := Stack[int]{}
// arbitraryTypeStack := Stack[ArbitraryType]{}

// Threadsafe Stack implementation
type Stack[T any] struct {
	items []T
	lock  sync.Mutex
}

func (s *Stack[T]) Init(startSize int) {
	s.items = make([]T, 0)
}

// Push adds an item to the top of the stack.
func (s *Stack[T]) Push(item T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.items = append(s.items, item)
}

// Pop removes and returns the top item from the stack.
func (s *Stack[T]) Pop() (T, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.items) == 0 {
		var zero T
		return zero, errors.New("stack is empty")
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, nil
}

// Peek returns the top item from the stack without removing it.
func (s *Stack[T]) Peek() (T, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.items) == 0 {
		var zero T
		return zero, errors.New("stack is empty")
	}
	return s.items[len(s.items)-1], nil
}

// IsEmpty checks if the stack is empty.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Size returns the number of items in the stack.
func (s *Stack[T]) Size() int {
	return len(s.items)
}

// could probably improve this
func getIdFromMessage(message *string, id *string) error {
	msgSlice := strings.Split(*message, ";")
	for _, v := range msgSlice {
		_, err := strconv.Atoi(v)
		if err == nil {
			*id = v
			return nil
		}
	}
	return fmt.Errorf("couldn't extract id from: %v", message)
}

type Message struct {
	message *string
	id      *string
}
