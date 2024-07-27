// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewProxy(
	transport http.RoundTripper,
	apiUrl *url.URL,
	proxyErrorHandler ProxyErrorHandler,
) http.Handler {
	reverseProxy := httputil.NewSingleHostReverseProxy(apiUrl)
	reverseProxy.ErrorHandler = proxyErrorHandler.HandleError
	reverseProxy.Transport = RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		req.Host = apiUrl.Host
		return transport.RoundTrip(req)
	})
	return reverseProxy
}
