// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ProxyErrorHandler interface {
	HandleError(resp http.ResponseWriter, req *http.Request, err error)
}

type ProxyErrorHandlerFunc func(resp http.ResponseWriter, req *http.Request, err error)

func (p ProxyErrorHandlerFunc) HandleError(resp http.ResponseWriter, req *http.Request, err error) {
	p(resp, req, err)
}

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
