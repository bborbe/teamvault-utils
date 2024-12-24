// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"sync"
)

type CurrentDateTimeGetter interface {
	Now() DateTime
}

type CurrentDateTimeSetter interface {
	SetNow(now DateTime)
}

//counterfeiter:generate -o mocks/time-current-time.go --fake-name TimeCurrentDateTime . CurrentDateTime
type CurrentDateTime interface {
	CurrentDateTimeGetter
	CurrentDateTimeSetter
}

func NewCurrentDateTime() CurrentDateTime {
	return &currentDateTime{}
}

type currentDateTime struct {
	mux sync.Mutex
	now *DateTime
}

func (n *currentDateTime) Now() DateTime {
	n.mux.Lock()
	defer n.mux.Unlock()
	if n.now != nil {
		return *n.now
	}
	return DateTime(Now())
}

func (n *currentDateTime) SetNow(now DateTime) {
	n.mux.Lock()
	defer n.mux.Unlock()
	n.now = &now
}
