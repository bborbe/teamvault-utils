// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

// Staging indicates whether the TeamVault instance is a staging environment.
type Staging bool

// Bool returns the boolean value of Staging.
func (s Staging) Bool() bool {
	return bool(s)
}
