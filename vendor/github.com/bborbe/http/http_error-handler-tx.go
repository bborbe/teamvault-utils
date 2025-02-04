// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http"

	libkv "github.com/bborbe/kv"
)

func NewUpdateErrorHandler(db libkv.DB, withErrorTx WithErrorTx) http.Handler {
	return NewErrorHandler(WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		return db.Update(ctx, func(ctx context.Context, tx libkv.Tx) error {
			return withErrorTx.ServeHTTP(ctx, tx, resp, req)
		})
	}))
}

func NewViewErrorHandler(db libkv.DB, withErrorTx WithErrorTx) http.Handler {
	return NewErrorHandler(WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		return db.View(ctx, func(ctx context.Context, tx libkv.Tx) error {
			return withErrorTx.ServeHTTP(ctx, tx, resp, req)
		})
	}))
}
