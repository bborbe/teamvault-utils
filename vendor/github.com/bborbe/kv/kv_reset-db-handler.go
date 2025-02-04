// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"context"
	"fmt"
	"net/http"
)

// NewResetHandler returns a http.Handler
// that allow delete the complete database
func NewResetHandler(db DB, cancel context.CancelFunc) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		defer cancel()
		if err := db.Close(); err != nil {
			http.Error(resp, fmt.Sprintf("reset db failed: %v", err), http.StatusInternalServerError)
			return
		}
		if err := db.Remove(); err != nil {
			http.Error(resp, fmt.Sprintf("remove db failed: %v", err), http.StatusInternalServerError)
			return
		}
		resp.WriteHeader(http.StatusOK)
		fmt.Fprint(resp, "reset db successful")
	})
}
