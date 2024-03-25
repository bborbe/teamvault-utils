// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"context"

	"github.com/golang/glog"
)

// LogErrors for the given func
func LogErrors(fn Func) Func {
	return func(ctx context.Context) error {
		if err := fn(ctx); err != nil {
			glog.Warning(err)
			return err
		}
		return nil
	}

}
