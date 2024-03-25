// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"context"
	"io"
	"sync"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// ConcurrentRunner allow run N tasks concurrent
type ConcurrentRunner interface {
	Add(ctx context.Context, fn Func)
	Run(ctx context.Context) error
	io.Closer
}

// NewConcurrentRunner returns ConcurrentRunner with the given concurrent limit
func NewConcurrentRunner(maxConcurrent int) ConcurrentRunner {
	return &concurrentRunner{
		maxConcurrent: maxConcurrent,
		fns:           make(chan Func, maxConcurrent),
		closed:        make(chan struct{}),
	}
}

type concurrentRunner struct {
	fns           chan Func
	maxConcurrent int

	mux    sync.Mutex
	closed chan struct{}
}

func (c *concurrentRunner) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	select {
	case <-c.closed:
		glog.V(3).Infof("already closed => skip")
		return errors.Errorf("already closed")
	default:
		glog.V(3).Infof("close concurrent runner")
		close(c.closed)
		close(c.fns)
		return nil
	}
}

func (c *concurrentRunner) Add(ctx context.Context, fn Func) {
	c.mux.Lock()
	defer c.mux.Unlock()
	select {
	case <-c.closed:
		glog.V(3).Infof("close discard added fn")
	default:
		select {
		case <-ctx.Done():
		case c.fns <- fn:
			glog.V(3).Infof("fn add to concurrent runner")
		}
	}
}

func (c *concurrentRunner) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	errs := make(chan error)
	limit := make(chan struct{}, c.maxConcurrent)
	defer func() {
		wg.Wait()
		close(limit)
		close(errs)
	}()

	return CancelOnFirstError(
		ctx,
		func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case fn, ok := <-c.fns:
					if !ok {
						return nil
					}
					limit <- struct{}{}
					wg.Add(1)
					go func() {
						defer func() {
							wg.Done()
							glog.V(3).Infof("fn complete to concurrent runner")
							<-limit
						}()
						err := fn(ctx)
						if err != nil {
							select {
							case <-ctx.Done():
							case errs <- errors.Wrap(err, "execute fn failed"):
							}
						}
					}()
				}
			}
		},
		func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-errs:
				return err
			case <-c.closed:
				return nil
			}
		},
	)
}
