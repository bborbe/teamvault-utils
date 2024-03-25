// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import "sync"

// Fire a trigger
type Fire interface {
	// Fire trigger als Dons ch to get a element
	Fire()
}

// Done check for a trigger
type Done interface {
	// Done chan gets a element if trigger was fired
	Done() <-chan struct{}
}

// Trigger combines fire and done
type Trigger interface {
	Fire
	Done
}

// NewTrigger create a new Trigger
func NewTrigger() Trigger {
	return &trigger{}
}

type trigger struct {
	mux sync.Mutex
	ch  chan struct{}
}

func (t *trigger) Fire() {
	t.mux.Lock()
	defer t.mux.Unlock()
	if t.ch == nil {
		t.ch = make(chan struct{})
		close(t.ch)
		return
	}
	select {
	case <-t.ch:
	default:
		close(t.ch)
	}
}

func (t *trigger) Done() <-chan struct{} {
	t.mux.Lock()
	defer t.mux.Unlock()
	if t.ch == nil {
		t.ch = make(chan struct{})
	}
	return t.ch
}
