// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/bborbe/errors"
)

func BuildRequest(
	ctx context.Context,
	method string,
	urlString string,
	parameters url.Values,
	body io.Reader,
	header http.Header,
) (*http.Request, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse url failed")
	}
	u.RawQuery = parameters.Encode()

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "create request failed")
	}
	if header != nil {
		req.Header = header
	}
	return req, nil
}
