// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"github.com/bborbe/errors"
)

// NewErrorList create a ErrorList with the given errors.
func NewErrorList(errs ...error) error {
	return errors.Join(errs...)
}

// NewErrorListByChan create a ErrorList with the given error channel.
func NewErrorListByChan(ch <-chan error) error {
	var errs []error
	for err := range ch {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
