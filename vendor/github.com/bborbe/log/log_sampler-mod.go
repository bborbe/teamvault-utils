// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

import "sync"

func NewSampleMod(mod uint64) Sampler {
	var counter uint64
	var mux sync.Mutex
	return SamplerFunc(func() bool {
		mux.Lock()
		defer mux.Unlock()
		return counter%mod == 0
	})
}
