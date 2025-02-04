// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime/pprof"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

/*
Example with mux:
router.Path("/profiling/start").Handler(libhttp.NewErrorHandler(libhttp.NewProfilingStart()))
router.Path("/profiling/stop").Handler(libhttp.NewErrorHandler(libhttp.NewProfilingStop()))
router.Path("/download/main").Handler(libhttp.NewErrorHandler(libhttp.NewFileDownloader("/main")))
router.Path("/download/cpu.pprof").Handler(libhttp.NewErrorHandler(libhttp.NewFileDownloader("/cpu.pprof")))
*/

func NewProfilingStart() WithError {
	return WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		f, err := os.Create("cpu.pprof")
		if err != nil {
			return err
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			return errors.Wrap(ctx, err, "start profiling failed")
		}
		glog.V(0).Infof("cpu profiling is enabled")
		fmt.Fprintf(resp, "cpu profiling is enabled")
		return nil
	})
}

func NewProfilingStop() WithError {
	return WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		pprof.StopCPUProfile()
		glog.V(0).Infof("cpu profiling is disabled")
		fmt.Fprintf(resp, "cpu profiling is disabled")
		return nil
	})
}
