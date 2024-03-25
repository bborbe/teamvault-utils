// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"sync"
	"time"
)

type CurrentTimeGetter interface {
	Now() time.Time
}

type CurrentTimeSetter interface {
	SetNow(now time.Time)
}

//counterfeiter:generate -o mocks/time-current-time.go --fake-name TimeCurrentTime . CurrentTime
type CurrentTime interface {
	CurrentTimeGetter
	CurrentTimeSetter
}

func NewCurrentTime() CurrentTime {
	return &currentTime{}
}

type currentTime struct {
	mux sync.Mutex
	now *time.Time
}

func (n *currentTime) Now() time.Time {
	n.mux.Lock()
	defer n.mux.Unlock()
	if n.now != nil {
		return *n.now
	}
	return Now()
}

func (n *currentTime) SetNow(now time.Time) {
	n.mux.Lock()
	defer n.mux.Unlock()
	n.now = &now
}
