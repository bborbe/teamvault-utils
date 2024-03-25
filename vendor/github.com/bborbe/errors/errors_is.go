// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import "errors"

func Is(err, target error) bool {
	return errors.Is(err, target)
}
