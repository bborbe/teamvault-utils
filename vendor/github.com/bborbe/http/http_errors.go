// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"errors"
	"net/http"
	"syscall"

	liberrors "github.com/bborbe/errors"
)

type HasTimeoutError interface {
	Timeout() bool
}

func IsRetryError(err error) bool {
	if errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, http.ErrHandlerTimeout) {
		return true
	}
	if timeoutError, ok := liberrors.Unwrap(err).(HasTimeoutError); ok {
		return timeoutError.Timeout()
	}
	return false
}

type HasTemporaryError interface {
	Temporary() bool
}
