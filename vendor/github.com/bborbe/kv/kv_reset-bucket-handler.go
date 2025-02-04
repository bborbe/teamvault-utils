// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// NewResetBucketHandler returns a http.Handler
// that allow delete a bucket
func NewResetBucketHandler(db DB, cancel context.CancelFunc) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ctx := context.Background()
		vars := mux.Vars(req)
		bucketName := BucketName(vars["BucketName"])
		if len(bucketName) == 0 {
			http.Error(resp, "parameter bucket missing", http.StatusBadRequest)
			return
		}
		err := db.Update(ctx, func(ctx context.Context, tx Tx) error {
			return tx.DeleteBucket(ctx, bucketName)
		})
		if err != nil {
			http.Error(resp, fmt.Sprintf("remove bucket failed: %v", err), http.StatusInternalServerError)
			return
		}
		resp.WriteHeader(http.StatusOK)
		fmt.Fprint(resp, "reset bucket successful")
		cancel()
	})
}
