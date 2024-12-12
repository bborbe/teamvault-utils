// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bborbe/errors"
)

func NewFileDownloader(path string) WithError {
	return WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		open, err := os.Open(path)
		if err != nil {
			return errors.Wrapf(ctx, err, "open %s failed", path)
		}

		fileInfo, err := open.Stat()
		if err != nil {
			return errors.Wrapf(ctx, err, "get stat failed")
		}

		resp.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name()))
		resp.Header().Set("Content-Type", "application/octet-stream")

		if _, err = io.Copy(resp, open); err != nil {
			return errors.Wrapf(ctx, err, "copy content failed")
		}
		return nil
	})
}
