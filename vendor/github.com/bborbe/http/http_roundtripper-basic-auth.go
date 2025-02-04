// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
)

func NewRoundTripperBasicAuth(
	roundTripper RoundTripper,
	username string,
	password string,
) http.RoundTripper {
	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if username != "" && password != "" {
			req.SetBasicAuth(username, password)
		}
		return roundTripper.RoundTrip(req)
	})
}
