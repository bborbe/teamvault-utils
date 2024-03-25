// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"io"

	"github.com/golang/glog"
)

func WriteAndGlog(w io.Writer, format string, a ...any) (n int, err error) {
	glog.V(2).InfoDepthf(1, format, a...)
	return fmt.Fprintf(w, format+"\n", a...)
}
