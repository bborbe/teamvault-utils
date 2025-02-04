// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http"
	"os"
	"runtime/pprof"
)

func NewCpuProfileStartHandler() WithError {
	return WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		f, err := os.Create("cpu.pprof")
		if err != nil {
			return err
		}
		return pprof.StartCPUProfile(f)
	})
}

func NewCpuProfileStopHandler() WithError {
	return WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		pprof.StopCPUProfile()
		return nil
	})
}
