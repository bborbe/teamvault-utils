// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"sync"
)

// AddFire allow add a new trigger
type AddFire interface {
	// Add returns a new fire for trigger
	Add() Fire
}

// MultiTrigger combines Done and AddFire
type MultiTrigger interface {
	Done
	AddFire
}

// NewMultiTrigger returns a MultiTrigger
func NewMultiTrigger() MultiTrigger {
	return &multiTrigger{}
}

type multiTrigger struct {
	mux      sync.Mutex
	triggers []Trigger
}

func (m *multiTrigger) Add() Fire {
	m.mux.Lock()
	defer m.mux.Unlock()

	result := NewTrigger()
	m.triggers = append(m.triggers, result)
	return result
}

func (m *multiTrigger) Done() <-chan struct{} {
	m.mux.Lock()
	defer m.mux.Unlock()

	result := NewTrigger()
	var wg sync.WaitGroup
	counter := 0
	for _, trigger := range m.triggers {
		v := trigger
		select {
		case <-v.Done():
			counter++
		default:
			wg.Add(1)
			go func(done Done) {
				select {
				case <-done.Done():
					wg.Done()
				}
			}(v)
		}
	}
	if counter == len(m.triggers) {
		result.Fire()
	} else {
		go func() {
			wg.Wait()
			result.Fire()
		}()
	}
	return result.Done()
}
