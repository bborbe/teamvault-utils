// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"
)

type Any []HasValidation

func (l Any) Validate(ctx context.Context) error {
	var err error
	for _, ll := range l {
		if err = ll.Validate(ctx); err == nil {
			return nil
		}
	}
	return err
}
