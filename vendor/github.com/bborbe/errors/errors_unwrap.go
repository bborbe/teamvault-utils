// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

func Unwrap(err error) error {
	switch e := err.(type) {
	case interface{ Unwrap() error }:
		return e.Unwrap()
	case interface{ Unwrap() []error }:
		if errs := e.Unwrap(); len(errs) > 0 {
			return errs[0]
		}
		return nil
	default:
		return err
	}
}
