// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

import "context"

// StreamList streams the given list into the given channel
func StreamList[T any](ctx context.Context, list []T, ch chan<- T) error {
	for _, e := range list {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- e:

		}
	}
	return nil
}
