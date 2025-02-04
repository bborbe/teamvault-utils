// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bborbe/errors"
	libkv "github.com/bborbe/kv"
)

//counterfeiter:generate -o mocks/http-json-handler-tx.go --fake-name HttpJsonHandlerTx . JsonHandlerTx
type JsonHandlerTx interface {
	ServeHTTP(ctx context.Context, tx libkv.Tx, req *http.Request) (interface{}, error)
}

type JsonHandlerTxFunc func(ctx context.Context, tx libkv.Tx, req *http.Request) (interface{}, error)

func (j JsonHandlerTxFunc) ServeHTTP(ctx context.Context, tx libkv.Tx, req *http.Request) (interface{}, error) {
	return j(ctx, tx, req)
}

func NewJsonHandlerViewTx(db libkv.DB, jsonHandler JsonHandlerTx) WithError {
	return WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		return db.View(ctx, func(ctx context.Context, tx libkv.Tx) error {
			result, err := jsonHandler.ServeHTTP(ctx, tx, req)
			if err != nil {
				return errors.Wrapf(ctx, err, "json handler failed")
			}
			resp.Header().Add(ContentTypeHeaderName, ApplicationJsonContentType)
			if err := json.NewEncoder(resp).Encode(result); err != nil {
				return errors.Wrapf(ctx, err, "encode json failed")
			}
			return nil
		})
	})
}

func NewJsonHandlerUpdateTx(db libkv.DB, jsonHandler JsonHandlerTx) WithError {
	return WithErrorFunc(func(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
		return db.Update(ctx, func(ctx context.Context, tx libkv.Tx) error {
			result, err := jsonHandler.ServeHTTP(ctx, tx, req)
			if err != nil {
				return errors.Wrapf(ctx, err, "json handler failed")
			}
			resp.Header().Add(ContentTypeHeaderName, ApplicationJsonContentType)
			if err := json.NewEncoder(resp).Encode(result); err != nil {
				return errors.Wrapf(ctx, err, "encode json failed")
			}
			return nil
		})
	})
}
