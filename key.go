// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bborbe/validation"
)

type Key string

func (k Key) String() string {
	return string(k)
}

func (k Key) Validate(ctx context.Context) error {
	if len(k) == 0 {
		return errors.Wrapf(ctx, validation.Error, "Key empty")
	}
	return nil
}
