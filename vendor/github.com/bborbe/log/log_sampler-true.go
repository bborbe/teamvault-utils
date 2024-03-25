// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

func NewSamplerTrue() Sampler {
	return SamplerFunc(func() bool {
		return true
	})
}
