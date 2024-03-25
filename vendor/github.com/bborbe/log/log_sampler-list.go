// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

type SamplerList []Sampler

func (s SamplerList) IsSample() bool {
	for _, sampler := range s {
		if sampler.IsSample() {
			return true
		}
	}
	return false
}
