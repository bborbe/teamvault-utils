// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import "github.com/pkg/errors"

func Cause(err error) error {
	return errors.Cause(err)
}
