// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

import (
	"sync"
)

type Set[T comparable] interface {
	Add(element T)
	Remove(element T)
	Contains(element T) bool
	Slice() []T
	Length() int
}

func NewSet[T comparable]() Set[T] {
	return &set[T]{
		data: make(map[T]struct{}),
	}
}

type set[T comparable] struct {
	mux  sync.Mutex
	data map[T]struct{}
}

func (s *set[T]) Add(element T) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.data[element] = struct{}{}
}

func (s *set[T]) Remove(element T) {
	s.mux.Lock()
	defer s.mux.Unlock()

	delete(s.data, element)
}

func (s *set[T]) Contains(element T) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	_, found := s.data[element]
	return found
}

func (s *set[T]) Slice() []T {
	s.mux.Lock()
	defer s.mux.Unlock()

	result := make([]T, 0, len(s.data))
	for k := range s.data {
		result = append(result, k)
	}
	return result
}

func (s *set[T]) Length() int {
	s.mux.Lock()
	defer s.mux.Unlock()

	return len(s.data)
}
