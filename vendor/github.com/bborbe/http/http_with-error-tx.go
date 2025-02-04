// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http"

	libkv "github.com/bborbe/kv"
)

//counterfeiter:generate -o mocks/http-with-error.go --fake-name HttpWithErrorTx . WithErrorTx
type WithErrorTx interface {
	ServeHTTP(ctx context.Context, tx libkv.Tx, resp http.ResponseWriter, req *http.Request) error
}

type WithErrorTxFunc func(ctx context.Context, tx libkv.Tx, resp http.ResponseWriter, req *http.Request) error

func (w WithErrorTxFunc) ServeHTTP(ctx context.Context, tx libkv.Tx, resp http.ResponseWriter, req *http.Request) error {
	return w(ctx, tx, resp, req)
}
