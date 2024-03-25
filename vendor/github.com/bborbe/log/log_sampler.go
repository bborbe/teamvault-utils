// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

//counterfeiter:generate -o mocks/sampler.go --fake-name Sampler . Sampler

// Sampler allow sample glog
//
//	sampler := NewSampleMod(10)
//	if sampler.IsSample() {
//	  glog.V(2).Infof("banana")
//	}
type Sampler interface {
	IsSample() bool
}
