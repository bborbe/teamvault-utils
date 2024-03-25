// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

//counterfeiter:generate -o mocks/http-with-error.go --fake-name HttpWithError . WithError
type WithError interface {
	ServeHTTP(ctx context.Context, resp http.ResponseWriter, req *http.Request) error
}

type WithErrorFunc func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error

func (w WithErrorFunc) ServeHTTP(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	return w(ctx, resp, req)
}

func NewErrorHandler(handlerWithError WithError) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		glog.V(3).Infof("handle %s request to %s started", req.Method, req.URL.Path)
		if err := handlerWithError.ServeHTTP(ctx, resp, req); err != nil {
			http.Error(resp, fmt.Sprintf("request failed: %v", err), http.StatusInternalServerError)
			glog.V(1).Infof("handle %s request to %s failed: %v", req.Method, req.URL.Path, err)
			return
		}
		glog.V(3).Infof("handle %s request to %s completed", req.Method, req.URL.Path)
	})
}
