// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"
	"errors"
)

var TransactionAlreadyOpenError = errors.New("transaction already open")

//counterfeiter:generate -o mocks/db.go --fake-name DB . DB
type DB interface {
	// Update opens a write transaction
	Update(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error

	// View opens a read only transaction
	View(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error

	// Sync database to disk
	Sync() error

	// Close database
	Close() error

	// Remove database files from disk
	Remove() error
}
