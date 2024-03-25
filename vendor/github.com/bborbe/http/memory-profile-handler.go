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

func NewMemoryProfileHandler() WithError {
	return WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		memoryFile, err := os.Create("memprofile.pprof")
		if err != nil {
			return err
		}
		if err := pprof.WriteHeapProfile(memoryFile); err != nil {
			return err
		}
		return nil
	})
}
