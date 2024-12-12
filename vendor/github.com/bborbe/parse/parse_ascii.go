// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"context"
	"unicode"

	"github.com/bborbe/errors"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// ParseAscii returns a the given string converted to ascii
func ParseAscii(ctx context.Context, value interface{}) (string, error) {
	str, err := ParseString(ctx, value)
	if err != nil {
		return "", errors.Wrapf(ctx, err, "convert to ascii failed")
	}
	result, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn))), str)
	if err != nil {
		return "", err
	}
	return result, nil
}
