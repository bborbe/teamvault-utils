// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

import stderrors "errors"

var NotFoundError = stderrors.New("not found")

func Find[T any](list []T, match func(value T) bool) (*T, error) {
	for _, e := range list {
		if match(e) {
			return &e, nil
		}
	}
	return nil, NotFoundError
}
