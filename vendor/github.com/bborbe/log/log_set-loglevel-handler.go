// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

func NewSetLoglevelHandler(ctx context.Context, logLevelSetter LogLevelSetter) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		level, err := strconv.Atoi(vars["level"])
		if err != nil {
			fmt.Fprintf(resp, "parse loglevel failed: %v\n", err)
			return
		}
		if err := logLevelSetter.Set(ctx, glog.Level(level)); err != nil {
			fmt.Fprintf(resp, "set loglevel failed: %v\n", err)
			return
		}
		fmt.Fprintf(resp, "set loglevel to %d completed\n", level)
	})
}
