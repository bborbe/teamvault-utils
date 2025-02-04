// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"

	"github.com/bborbe/errors"
)

func ForEach(
	ctx context.Context,
	bucket Bucket,
	fn func(item Item) error,
) error {
	it := bucket.Iterator()
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := fn(it.Item()); err != nil {
				return errors.Wrapf(ctx, err, "fn failed")
			}
		}
	}
	return nil
}
