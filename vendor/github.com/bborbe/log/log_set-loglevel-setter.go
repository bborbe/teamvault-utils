// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package log

import (
	"context"
	"flag"
	"strconv"
	"sync"
	"time"

	"github.com/golang/glog"
)

type LogLevelSetter interface {
	Set(ctx context.Context, logLevel glog.Level) error
}

type LogLevelSetterFunc func(ctx context.Context, logLevel glog.Level) error

func (l LogLevelSetterFunc) Set(ctx context.Context, logLevel glog.Level) error {
	return l(ctx, logLevel)
}

func NewLogLevelSetter(
	defaultLoglevel glog.Level,
	autoResetDuration time.Duration,
) LogLevelSetter {
	return &logLevelSetter{
		defaultLoglevel:   defaultLoglevel,
		autoResetDuration: autoResetDuration,
	}

}

type logLevelSetter struct {
	autoResetDuration time.Duration
	defaultLoglevel   glog.Level

	mux             sync.Mutex
	lastSetTime     time.Time
	currentLogLevel glog.Level
}

func (l *logLevelSetter) Set(ctx context.Context, logLevel glog.Level) error {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.lastSetTime = time.Now()
	l.currentLogLevel = logLevel
	_ = flag.Set("v", strconv.Itoa(int(logLevel)))

	glog.V(l.defaultLoglevel).Infof("set loglevel to %d and reset in %v back to %d", logLevel, l.autoResetDuration, l.defaultLoglevel)
	go func() {
		ctx, cancel := context.WithTimeout(ctx, l.autoResetDuration)
		defer cancel()

		select {
		case <-ctx.Done():
			l.resetLogLevel()
		}
	}()
	return nil
}

func (l *logLevelSetter) resetLogLevel() {
	if time.Since(l.lastSetTime) <= l.autoResetDuration {
		glog.V(l.defaultLoglevel).Infof("time since lastSet is to short => skip reset loglevel")
		return
	}

	_ = flag.Set("v", strconv.Itoa(int(l.defaultLoglevel)))
	glog.V(l.defaultLoglevel).Infof("loglevel set back to %d", l.defaultLoglevel)
}
