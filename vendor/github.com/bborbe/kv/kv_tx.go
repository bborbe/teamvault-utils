// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"
	"errors"
)

var BucketNotFoundError = errors.New("bucket not found")
var BucketAlreadyExistsError = errors.New("bucket already exists")

//counterfeiter:generate -o mocks/tx.go --fake-name Tx . Tx
type Tx interface {
	Bucket(ctx context.Context, name BucketName) (Bucket, error)
	CreateBucket(ctx context.Context, name BucketName) (Bucket, error)
	CreateBucketIfNotExists(ctx context.Context, name BucketName) (Bucket, error)
	DeleteBucket(ctx context.Context, name BucketName) error
	ListBucketNames(ctx context.Context) (BucketNames, error)
}
