// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "net/http"

//counterfeiter:generate -o mocks/http-round-tripper.go --fake-name HttpRoundTripper . RoundTripper
type RoundTripper interface {
	http.RoundTripper
}
