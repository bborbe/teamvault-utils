// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"time"
)

func CreateDefaultHttpClient() *http.Client {
	return CreateHttpClient(30 * time.Second)
}

func CreateHttpClient(
	timeout time.Duration,
) *http.Client {
	return &http.Client{
		Transport: CreateDefaultRoundTripper(),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: timeout,
	}
}
