// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"
)

func Count(ctx context.Context, bucket Bucket) (int64, error) {
	var counter int64
	it := bucket.Iterator()
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		select {
		case <-ctx.Done():
			return -1, ctx.Err()
		default:
			counter++
		}
	}
	return counter, nil
}
