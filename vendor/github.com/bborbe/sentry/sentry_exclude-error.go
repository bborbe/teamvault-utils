// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sentry

type ExcludeErrors []ExcludeError

func (e ExcludeErrors) IsExcluded(err error) bool {
	for _, ee := range e {
		if ee(err) {
			return true
		}
	}
	return false
}

type ExcludeError func(err error) bool
