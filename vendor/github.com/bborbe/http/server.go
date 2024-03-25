// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bborbe/errors"
	"github.com/bborbe/run"
	"github.com/golang/glog"
)

func NewServerWithPort(port int, router http.Handler) run.Func {
	return NewServer(
		fmt.Sprintf(":%d", port),
		router,
	)
}
func NewServer(addr string, router http.Handler) run.Func {
	return func(ctx context.Context) error {
		server := &http.Server{
			Addr:    addr,
			Handler: router,
		}
		go func() {
			select {
			case <-ctx.Done():
				if err := server.Shutdown(ctx); err != nil {
					glog.Warningf("shutdown failed: %v", err)
				}
			}
		}()
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			glog.V(0).Info(err)
			return nil
		}
		return errors.Wrapf(ctx, err, "httpServer failed")
	}
}
