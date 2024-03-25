// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

import (
	"sync"
	stdtime "time"

	libtime "github.com/bborbe/time"
)

func NewSampleTime(duration stdtime.Duration) Sampler {
	var mux sync.Mutex
	var lastlog stdtime.Time
	return SamplerFunc(func() bool {
		mux.Lock()
		defer mux.Unlock()
		if libtime.Now().Sub(lastlog) <= duration {
			return false
		}
		lastlog = libtime.Now()
		return true
	})
}
