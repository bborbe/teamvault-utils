// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"context"
	"flag"
	"runtime"
	"time"

	"github.com/bborbe/argument/v2"
	"github.com/bborbe/run"
	"github.com/golang/glog"
)

func Main(
	ctx context.Context,
	app run.Runnable,
) int {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())
	_ = flag.Set("logtostderr", "true")

	time.Local = time.UTC
	glog.V(2).Infof("set global timezone to UTC")

	if err := argument.Parse(ctx, app); err != nil {
		glog.Errorf("parse app failed: %v", err)
		return 4
	}

	glog.V(3).Infof("application started")
	if err := app.Run(run.ContextWithSig(ctx)); err != nil {
		glog.Error(err)
		return 1
	}
	glog.V(3).Infof("application finished")
	return 0
}
