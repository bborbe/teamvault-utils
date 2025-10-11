// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"fmt"
	"strings"
)

type ApiUrl string

func (a ApiUrl) String() string {
	return string(a)
}

func (a ApiUrl) Key() (Key, error) {
	parts := strings.Split(a.String(), "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("parse key form api-url failed")
	}
	return Key(parts[len(parts)-2]), nil
}
