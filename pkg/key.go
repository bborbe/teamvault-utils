// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/bborbe/validation"
)

// Key represents a TeamVault secret identifier.
type Key string

// String returns the string representation of the Key.
func (k Key) String() string {
	return string(k)
}

// Validate checks if the Key is not empty.
func (k Key) Validate(ctx context.Context) error {
	if len(k) == 0 {
		return errors.Wrapf(ctx, validation.Error, "Key empty")
	}
	return nil
}
