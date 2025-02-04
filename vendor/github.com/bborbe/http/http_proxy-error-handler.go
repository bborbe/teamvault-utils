// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "net/http"

//counterfeiter:generate -o mocks/http-proxy-error-handler.go --fake-name HttpProxyErrorHandler . ProxyErrorHandler
type ProxyErrorHandler interface {
	HandleError(resp http.ResponseWriter, req *http.Request, err error)
}

type ProxyErrorHandlerFunc func(resp http.ResponseWriter, req *http.Request, err error)

func (p ProxyErrorHandlerFunc) HandleError(resp http.ResponseWriter, req *http.Request, err error) {
	p(resp, req, err)
}
