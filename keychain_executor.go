// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import "context"

//counterfeiter:generate -o mocks/executor.go --fake-name Executor . Executor

// Executor runs an external command and returns its output.
// It exists as an interface to allow test injection in place of os/exec.
type Executor interface {
	Run(
		ctx context.Context,
		name string,
		args []string,
		stdin string,
	) (stdout string, stderr string, exitCode int, err error)
}
