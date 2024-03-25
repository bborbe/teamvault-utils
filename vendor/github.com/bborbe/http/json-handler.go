// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bborbe/errors"
)

//counterfeiter:generate -o mocks/http-json-handler.go --fake-name HttpJsonHandler . JsonHandler
type JsonHandler interface {
	ServeHTTP(ctx context.Context, req *http.Request) (interface{}, error)
}
type JsonHandlerFunc func(ctx context.Context, req *http.Request) (interface{}, error)

func (j JsonHandlerFunc) ServeHTTP(ctx context.Context, req *http.Request) (interface{}, error) {
	return j(ctx, req)
}

func NewJsonHandler(jsonHandler JsonHandler) WithError {
	return WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		result, err := jsonHandler.ServeHTTP(ctx, req)
		if err != nil {
			return errors.Wrapf(ctx, err, "json handler failed")
		}
		resp.Header().Add(ContentTypeHeaderName, ApplicationJsonContentType)
		if err := json.NewEncoder(resp).Encode(result); err != nil {
			return errors.Wrapf(ctx, err, "encode json failed")
		}
		return nil
	})
}
