// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http"
)

//counterfeiter:generate -o mocks/http-with-error.go --fake-name HttpWithError . WithError
type WithError interface {
	ServeHTTP(ctx context.Context, resp http.ResponseWriter, req *http.Request) error
}

type WithErrorFunc func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error

func (w WithErrorFunc) ServeHTTP(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	return w(ctx, resp, req)
}
