// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "net/http"

func NewRoundTripperHeader(
	roundTripper http.RoundTripper,
	header http.Header,
) http.RoundTripper {
	return &roundTripperHeader{
		roundTripper: roundTripper,
		header:       header,
	}
}

type roundTripperHeader struct {
	roundTripper http.RoundTripper
	header       http.Header
}

func (a *roundTripperHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, values := range a.header {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return a.roundTripper.RoundTrip(req)
}
