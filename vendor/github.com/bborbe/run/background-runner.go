// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"context"

	"github.com/golang/glog"
)

type BackgroundRunner interface {
	Run(runFunc Func) error
}

func NewBackgroundRunner(ctx context.Context) BackgroundRunner {
	return &backgroundRunner{
		ctx:             ctx,
		parallelSkipper: NewParallelSkipper(),
	}
}

type backgroundRunner struct {
	parallelSkipper ParallelSkipper
	ctx             context.Context
}

func (b *backgroundRunner) Run(runFunc Func) error {
	go func() {
		action := b.parallelSkipper.SkipParallel(func(ctx context.Context) error {
			if err := runFunc(ctx); err != nil {
				return err
			}
			return nil
		})
		glog.V(3).Infof("run started")
		if err := action(b.ctx); err != nil {
			glog.Warningf("run failed: %v", err)
		}
		glog.V(3).Infof("run completed")
	}()
	return nil
}
