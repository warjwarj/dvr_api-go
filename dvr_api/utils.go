package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
~~~~~~~~~~~~~~~
STACK
Threadsafe Stack implementation
~~~~~~~~~~~~~~~
*/

type Stack[T any] struct {
	items []T
	lock  sync.Mutex
}

// init
func (s *Stack[T]) Init() {
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
		var zVal T
		return zVal, errors.New("stack is empty")
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

/*
~~~~~~~~~~~~~~~
DICTIONARY
Threadsafe dictionary implementation
~~~~~~~~~~~~~~~
*/

// Define a dictionary type
type Dictionary[T any] struct {
	internal map[string]*T
	lock     sync.Mutex
}

func (d *Dictionary[T]) Init() {
	d.internal = make(map[string]*T)
}

// Function to add a key-value pair to the dictionary
func (d *Dictionary[T]) Add(key string, value T) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.internal[key] = &value
}

// Function to get a value from the dictionary by key
func (d *Dictionary[T]) Get(key string) (*T, bool) {
	d.lock.Lock()
	defer d.lock.Unlock()
	value, exists := d.internal[key]
	return value, exists
}

// Function to delete a key-value pair from the dictionary
func (d *Dictionary[T]) Delete(key string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.internal, key)
}

/*
~~~~~~~~~~~~~~~
MDVR PROTOCOL HELPERS
~~~~~~~~~~~~~~~
*/

// assuming the datetime is always at the fourth index
// Assumptions: date is at the fith index of the message, and that the date is formatted like: 20060102-150405
func getDateFromMessage(message string) (time.Time, error) {
	dtStr := strings.Split(message, ";")[4]
	if len(dtStr) != 15 {
		return time.Time{}, fmt.Errorf("Input string length should be 15 characters")
	}
	// Parse the input string as a time.Time object
	parsedTime, err := time.Parse("20060102-150405", dtStr)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}

// get id from msg format: IDENTIFIER;1234;<-That's the id;xXxXxXxXxX;<CR>
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
