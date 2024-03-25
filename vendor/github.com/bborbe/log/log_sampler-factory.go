// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

import "time"

var DefaultSamplerFactory SamplerFactory = SamplerFactoryFunc(func() Sampler {
	return SamplerList{
		NewSampleTime(10 * time.Second),
		NewSamplerGlogLevel(4),
	}
})

// SamplerFactory allow to inject sampler
type SamplerFactory interface {
	Sampler() Sampler
}

type SamplerFactoryFunc func() Sampler

func (s SamplerFactoryFunc) Sampler() Sampler {
	return s()
}
