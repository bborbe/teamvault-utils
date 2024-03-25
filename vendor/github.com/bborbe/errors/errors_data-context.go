// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import (
	"context"
	"sync"
)

type dataCtxKeyType string

const dataCtxKey dataCtxKeyType = "data"

var mutex sync.Mutex

func AddToContext(ctx context.Context, key, value string) context.Context {
	v := ctx.Value(dataCtxKey)
	if v == nil {
		return context.WithValue(ctx, dataCtxKey, map[string]string{
			key: value,
		})
	}
	data, ok := v.(map[string]string)
	if ok {
		mutex.Lock()
		data[key] = value
		mutex.Unlock()
	}
	return ctx
}

func DataFromContext(ctx context.Context) map[string]string {
	value := ctx.Value(dataCtxKey)
	if value == nil {
		return nil
	}
	return value.(map[string]string)
}
